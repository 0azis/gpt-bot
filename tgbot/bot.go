package tgbot

import (
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	"gpt-bot/config"
	"gpt-bot/internal/db"
	"gpt-bot/internal/db/domain"
	"gpt-bot/utils"
	"log/slog"
	"os"
	"os/signal"
	"strconv"
	"strings"

	"github.com/0azis/bot"
	"github.com/0azis/bot/models"
	"github.com/go-telegram/fsm"
)

type BotInterface interface {
	Start()
	Instance() *bot.Bot
	InitHandlers()

	// handlers
	startHandler(ctx context.Context, b *bot.Bot, update *models.Update)
	defaultHandler(ctx context.Context, b *bot.Bot, update *models.Update)

	// helpers
	CreateInvoiceLink(payload []byte, paymentCredentials domain.Payment) (string, error)
	PaymentInfo(paymentCredentials domain.Payment, status bool)
	getTelegramAvatar(ctx context.Context, userID int64) string
	informUser(ctx context.Context, userID int64, errMsg string)

	// admin
	adminHandler(ctx context.Context, b *bot.Bot, update *models.Update)
	usersStatisticsCallback(ctx context.Context, b *bot.Bot, update *models.Update)
}

type tgBot struct {
	telegram config.Telegram
	store    db.Store
	b        *bot.Bot
	f        *fsm.FSM
}

func New(cfg config.Telegram, store db.Store) (BotInterface, error) {
	b, err := bot.New(cfg.GetToken())

	tgBot := tgBot{
		telegram: cfg,
		store:    store,
		b:        b,
	}

	f := fsm.New(stateDefault, map[fsm.StateID]fsm.Callback{
		stateAskUserID:         tgBot.callbackUserID,
		stateAskDiamondsAmount: tgBot.callbackDiamondsAmount,
	})
	tgBot.f = f
	return tgBot, err
}

func (tb tgBot) Start() {
	ctx, cancel := signal.NotifyContext(context.Background(), os.Interrupt)
	defer cancel()
	tb.b.Start(ctx)
}

func (tb tgBot) Instance() *bot.Bot {
	return tb.b
}

func (tb tgBot) InitHandlers() {
	// base
	tb.b.RegisterHandler(bot.HandlerTypeMessageText, "/start", bot.MatchTypeContains, tb.startHandler)
	tb.b.SetDefaultHandler(tb.defaultHandler)

	// admin
	tb.b.RegisterHandler(bot.HandlerTypeMessageText, "/admin", bot.MatchTypeContains, tb.adminHandler)
	tb.b.RegisterHandler(bot.HandlerTypeCallbackQueryData, userStatistics, bot.MatchTypePrefix, tb.usersStatisticsCallback)
	tb.b.RegisterHandler(bot.HandlerTypeCallbackQueryData, requestsStatistics, bot.MatchTypePrefix, tb.requestsStatisticsCallback)
}

func (tb tgBot) startHandler(ctx context.Context, b *bot.Bot, update *models.Update) {
	var user domain.User
	user.ID = int(update.Message.From.ID)
	user.Avatar = tb.getTelegramAvatar(ctx, int64(user.ID))

	err := tb.store.User.Create(user)
	if err != nil {
		slog.Error(err.Error())
		tb.informUser(ctx, int64(user.ID), userCreationError)
		return
	}
	err = tb.store.Subscription.InitStandard(user.ID)
	if err != nil {
		slog.Error(err.Error())
		tb.informUser(ctx, int64(user.ID), userCreationError)
		return
	}
	limits := domain.NewLimits(user.ID, domain.SubscriptionStandard)
	err = tb.store.Limits.Create(limits)
	if err != nil {
		slog.Error(err.Error())
		tb.informUser(ctx, int64(user.ID), userCreationError)
		return
	}

	token := utils.NewToken()
	token.SetUserID(user.ID)
	err = token.SignJWT()
	if err != nil {
		slog.Error(err.Error())
		tb.informUser(ctx, int64(user.ID), userCreationError)
		return
	}

	msgSlice := strings.Split(update.Message.Text, " ")
	if len(msgSlice) == 2 {
		refCode := &msgSlice[1]
		id, err := tb.store.User.IsUserReferred(user.ID, *refCode)
		if err != nil {
			tb.informUser(ctx, int64(user.ID), internalError)
			return
		}
		if id != 0 {
			tb.informUser(ctx, int64(user.ID), userAlreadyReferred)
			return
		}

		ownerID, err := tb.store.User.OwnerOfReferralCode(*refCode)
		if errors.Is(err, sql.ErrNoRows) {
			tb.informUser(ctx, int64(user.ID), referralInvalid)
			return
		}
		if ownerID == user.ID {
			tb.informUser(ctx, int64(user.ID), referralSameUser)
			return
		}
		if err != nil {
			slog.Error(err.Error())
			tb.informUser(ctx, int64(user.ID), internalError)
			return
		}

		award := domain.ReferralAward
		err = tb.store.User.RaiseBalance(ownerID, award)
		if err != nil {
			slog.Error(err.Error())
			tb.informUser(ctx, int64(user.ID), internalError)
			return
		}
		err = tb.store.User.SetReferredBy(user.ID, *refCode)
		if err != nil {
			slog.Error(err.Error())
			tb.informUser(ctx, int64(user.ID), internalError)
			return
		}
	}

	_, err = b.SetChatMenuButton(ctx, &bot.SetChatMenuButtonParams{
		ChatID: update.Message.Chat.ID,
		MenuButton: models.MenuButtonWebApp{
			Type: "web_app",
			Text: "Open Mini App",
			WebApp: models.WebAppInfo{
				URL: tb.telegram.GetWebAppUrl() + "?token=" + token.GetStrToken(),
			},
		},
	})
	if err != nil {
		slog.Error(err.Error())
		tb.informUser(ctx, update.Message.From.ID, miniAppError)
	}

	b.SendMessage(ctx, &bot.SendMessageParams{
		ChatID: update.Message.Chat.ID,
		Text:   fmt.Sprintf("Hello!, %s", update.Message.From.Username),
	})
}

func (tb tgBot) getTelegramAvatar(ctx context.Context, userID int64) string {
	var url string

	// get list of photos
	photos, _ := tb.b.GetUserProfilePhotos(ctx, &bot.GetUserProfilePhotosParams{
		UserID: userID,
	})

	if photos.TotalCount == 0 {
		url = "" // if user don't have an avatar
	} else {
		file, _ := tb.b.GetFile(ctx, &bot.GetFileParams{
			FileID: photos.Photos[0][0].FileID, // take the first avatar
		})
		url = tb.b.FileDownloadLink(file) // generate link
	}

	return url
}

func (tb tgBot) CreateInvoiceLink(payload []byte, payment domain.Payment) (string, error) {
	link, err := tb.b.CreateInvoiceLink(context.Background(), &bot.CreateInvoiceLinkParams{
		Title:       fmt.Sprintf("%s subscription", payment.SubscriptionName),
		Description: "Buy subscription",
		Payload:     string(payload),
		Currency:    "XTR",
		Prices: []models.LabeledPrice{
			{Label: payment.SubscriptionName, Amount: payment.Amount},
		},
	})
	return link, err
}

func (tb tgBot) informUser(ctx context.Context, userID int64, errorMsg string) {
	tb.b.SendMessage(ctx, &bot.SendMessageParams{
		ChatID: userID,
		Text:   fmt.Sprintf("%s", errorMsg),
	})
}

func (tb tgBot) PaymentInfo(payment domain.Payment, status bool) {
	if status {
		tb.b.SendMessage(context.Background(), &bot.SendMessageParams{
			ChatID:    payment.UserID,
			Text:      fmt.Sprintf(`Вы успешно купили подписку <b>%s</b> за <i>%d</i> звезд. Ваша подписка доступна до: %s`, payment.SubscriptionName, payment.Amount, payment.End),
			ParseMode: models.ParseModeHTML,
		})
	} else {
		tb.b.SendMessage(context.Background(), &bot.SendMessageParams{
			ChatID:    payment.UserID,
			Text:      fmt.Sprintf(`Неудалось купить подписку <b>%s</b> за <i>%d</i> звезд. Попробуйте позже`, payment.SubscriptionName, payment.Amount),
			ParseMode: models.ParseModeHTML,
		})
	}
}

func (tb tgBot) defaultHandler(ctx context.Context, b *bot.Bot, update *models.Update) {
	var payment domain.Payment

	if update.PreCheckoutQuery != nil {
		err := json.Unmarshal([]byte(update.PreCheckoutQuery.InvoicePayload), &payment)
		if err != nil {
			slog.Error(err.Error())
			return
		}

		_, err = b.AnswerPreCheckoutQuery(ctx, &bot.AnswerPreCheckoutQueryParams{
			PreCheckoutQueryID: update.PreCheckoutQuery.ID,
			OK:                 true,
			ErrorMessage:       "",
		})
		if err != nil {
			slog.Error(err.Error())
			tb.informUser(ctx, int64(payment.UserID), paymentError)
			return
		}

		return
	}

	if update.Message != nil {
		if update.Message.SuccessfulPayment != nil {
			err := json.Unmarshal([]byte(update.Message.SuccessfulPayment.InvoicePayload), &payment)
			if err != nil {
				slog.Error(err.Error())
				tb.informUser(ctx, update.Message.From.ID, internalError)
				return
			}
			err = tb.store.Subscription.Update(payment.UserID, payment.SubscriptionName, payment.End)
			if err != nil {
				slog.Error(err.Error())
				tb.informUser(ctx, int64(payment.UserID), internalError)
				return
			}
			diamonds, err := tb.store.Subscription.DailyDiamonds(payment.SubscriptionName)
			if err != nil {
				slog.Error(err.Error())
				tb.informUser(ctx, int64(payment.UserID), internalError)
				return
			}
			err = tb.store.User.FillBalance(payment.UserID, diamonds)
			if err != nil {
				slog.Error(err.Error())
				tb.informUser(ctx, int64(payment.UserID), internalError)
				return
			}
			limits := domain.NewLimits(payment.UserID, payment.SubscriptionName)
			err = tb.store.Limits.Update(limits)
			if err != nil {
				slog.Error(err.Error())
				tb.informUser(ctx, int64(payment.UserID), internalError)
				return
			}
			tb.PaymentInfo(payment, true)
			return
		}
		if update.CallbackQuery != nil {
			diamonds := diamondsScheme{}
			switch tb.f.Current(update.CallbackQuery.From.ID) {
			case stateDefault:
				tb.b.SendMessage(ctx, &bot.SendMessageParams{
					ChatID: update.CallbackQuery.From.ID,
					Text:   "Type",
				})
				return
			case stateAskUserID:
				id, err := strconv.Atoi(update.Message.Text)
				if err != nil {
				}
				diamonds.userID = id
				tb.f.Transition(update.Message.From.ID, stateAskDiamondsAmount, update.Message.Chat.ID)
			case stateAskDiamondsAmount:
				amount, err := strconv.Atoi(update.Message.Text)
				if err != nil {
				}
				diamonds.amount = amount
				fmt.Println(diamonds)
				tb.giveDiamonds(diamonds.userID, diamonds.amount)
			}
		}
	}
}
