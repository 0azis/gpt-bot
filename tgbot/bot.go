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
	IsUserMember(channelName string, userID int) bool
	PaymentInfo(paymentCredentials domain.Payment, status bool)
	GetChannelInfo(channelName string) (domain.Channel, error)
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
	b.SetMyDescription(context.Background(), &bot.SetMyDescriptionParams{
		Description: "✨ Я - умная нейросеть, которая поможет тебе справиться с любой задачей!\n\n🔥Вы можете воспользоваться самыми современными инструментами:\n\n- DALLE + Runware + Midjourney: Генерация изображений на основе текста!\n\n- ChatGPT : получение ответов на вопросы, советы и помощь в разговоре на любую тему!\n\n➕ Редактор изображений, создание видео, увеличенная скорость ответа и генераций и многое другое.",
	})
	tgBot := tgBot{
		telegram: cfg,
		store:    store,
		b:        b,
	}

	f := fsm.New(stateDefault, map[fsm.StateID]fsm.Callback{
		//diamonds
		diamondStateAskUserID:         tgBot.callbackUserID,
		diamondStateAskDiamondsAmount: tgBot.callbackDiamondsAmount,
		// subscription
		subscriptionStateAskUserID: tgBot.callbackUserID,
		subscriptionStateAskName:   tgBot.callbackSubscriptionName,
		// bonus
		bonusStateAskChannel: tgBot.callbackChannelName,
		bonusStateAskAward:   tgBot.callbackBonusAward,
		bonusStateDelete:     tgBot.callbackChannelName,
		// referral
		referralStateDelete: tgBot.callbackReferralId,

		// default
		// stateDefault: tgBot.stateDefault,
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
	tb.b.RegisterHandler(bot.HandlerTypeCallbackQueryData, "btn_", bot.MatchTypePrefix, tb.callbackHandler)

	// other
	tb.b.RegisterHandler(bot.HandlerTypeMessageText, "/app", bot.MatchTypeExact, tb.appHandler)
	tb.b.RegisterHandler(bot.HandlerTypeMessageText, "/menu", bot.MatchTypeExact, tb.menuHandler)

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
	err = tb.store.Bonus.InitBonuses(user.ID)
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
		if len(*refCode) == utils.UserRefCode {
			id, err := tb.store.User.IsUserReferred(user.ID, *refCode)
			if err != nil {
				slog.Error(err.Error())
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
		if len(*refCode) == utils.AdRefCode {
			refID, err := tb.store.Referral.GetOne(*refCode)
			if errors.Is(err, sql.ErrNoRows) {
				tb.informUser(ctx, int64(user.ID), referralInvalid)
				return
			}
			if err != nil {
				slog.Error(err.Error())
				tb.informUser(ctx, int64(user.ID), internalError)
				return
			}
			tb.store.Referral.AddUser(user.ID, refID)
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
		ChatID:    update.Message.Chat.ID,
		Text:      "Привет! 👋\n\n🥳 Добро пожаловать в WEBAI App! Вы можете воспользоваться любимыми нейросетями в удобном формате мини-приложения.\n\nКоманды:\n/help - техническая поддержка\n/menu - управление ботом\n/app - мини-приложение\n/start - перезапустить бота",
		ParseMode: models.ParseModeHTML,
	})
}

func (tb tgBot) appHandler(ctx context.Context, b *bot.Bot, update *models.Update) {
	token := utils.NewToken()
	token.SetUserID(int(update.Message.From.ID))
	err := token.SignJWT()
	if err != nil {
		slog.Error(err.Error())
		tb.informUser(ctx, update.Message.From.ID, userCreationError)
		return
	}

	kb := &models.InlineKeyboardMarkup{
		InlineKeyboard: [][]models.InlineKeyboardButton{
			{
				{Text: "Открыть WebAI", WebApp: &models.WebAppInfo{
					URL: tb.telegram.GetWebAppUrl() + "?token=" + token.GetStrToken(),
				}},
			},
		},
	}
	b.SendMessage(ctx, &bot.SendMessageParams{
		ChatID:      update.Message.From.ID,
		Text:        "Нажмите на кнопку ниже для открытия мини-приложения.",
		ReplyMarkup: kb,
	})
}

func (tb tgBot) menuHandler(ctx context.Context, b *bot.Bot, update *models.Update) {
	userID := int(update.Message.From.ID)
	subscriptionName, err := tb.store.Subscription.GetSubscription(userID)
	if err != nil {
		slog.Error(err.Error())
		tb.informUser(ctx, int64(userID), internalError)
		return

	}
	limits, err := tb.store.Limits.GetByUser(userID)
	if err != nil {
		slog.Error(err.Error())
		tb.informUser(ctx, int64(userID), internalError)
		return
	}

	b.SendMessage(ctx, &bot.SendMessageParams{
		ChatID:    userID,
		Text:      fmt.Sprintf("ID: %d\nПодписка: <i>%s</i>\n\n<pre><code><b>Лимиты запросов:</b>\nGPT o1: %d\nGPT o1-mini: %d\nGPT 4o: %d\nGPT 4o-mini: %d\nDALL-E 3: %d\nRunware: %d</code></pre>\n\n<i>Лимиты для пользователей обновляются каждый день</i>", userID, subscriptionName, limits.O1Preview, limits.O1Mini, limits.Gpt4o, limits.Gpt4oMini, limits.Dalle3, limits.Runware),
		ParseMode: models.ParseModeHTML,
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

func (tb tgBot) IsUserMember(channelName string, userID int) bool {
	channelName = "@" + channelName
	member, _ := tb.b.GetChatMember(context.Background(), &bot.GetChatMemberParams{
		ChatID: channelName,
		UserID: int64(userID),
	})

	if member.Member == nil {
		return false
	}

	return true
}

func (tb tgBot) GetChannelInfo(channelName string) (domain.Channel, error) {
	var channel domain.Channel
	channel.Name = channelName
	channelName = "@" + channelName

	tgChannel, err := tb.b.GetChat(context.Background(), &bot.GetChatParams{
		ChatID: channelName,
	})
	if err != nil {
		slog.Error(err.Error())
		return channel, err
	}

	channel.Title = tgChannel.Title
	channel.Link = "https://t.me/" + tgChannel.Username

	file, err := tb.b.GetFile(context.Background(), &bot.GetFileParams{
		FileID: tgChannel.Photo.SmallFileID,
	})
	if err != nil {
		slog.Error(err.Error())
		return channel, err
	}
	url := tb.b.FileDownloadLink(file)

	channel.Avatar = url
	return channel, nil
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
		if update.Message.Text != "" {
			switch tb.f.Current(update.Message.From.ID) {
			case stateDefault:
				b.SendMessage(ctx, &bot.SendMessageParams{
					ChatID: update.Message.From.ID,
					Text:   "Введена некорректная команда",
				})
				return
			case diamondStateAskUserID:
				id, err := strconv.Atoi(update.Message.Text)
				if err != nil {
					tb.informUser(ctx, update.Message.From.ID, incorrectData)
					return
				}
				diamondsScheme.userID = id
				tb.f.Transition(update.Message.From.ID, diamondStateAskDiamondsAmount, update.Message.Chat.ID)
			case diamondStateAskDiamondsAmount:
				amount, err := strconv.Atoi(update.Message.Text)
				if err != nil {
					tb.informUser(ctx, update.Message.From.ID, incorrectData)
					return
				}
				diamondsScheme.amount = amount
				tb.giveDiamonds(diamondsScheme, update.Message)
				tb.f.Transition(update.Message.From.ID, stateDefault, update.Message.Chat.ID)

			// subscription
			case subscriptionStateAskUserID:
				id, err := strconv.Atoi(update.Message.Text)
				if err != nil {
					tb.informUser(ctx, update.Message.From.ID, incorrectData)
					return
				}
				subscriptionScheme.userID = id
				tb.f.Transition(update.Message.From.ID, subscriptionStateAskName, update.Message.Chat.ID)
			case subscriptionStateAskName:
				subscriptionScheme.name = update.Message.Text
				tb.giveSubscription(subscriptionScheme, update.Message)
				tb.f.Transition(update.Message.From.ID, stateDefault, update.Message.Chat.ID)
			// bonus
			case bonusStateAskChannel:
				bonusScheme.channel_name = update.Message.Text
				tb.f.Transition(update.Message.From.ID, bonusStateAskAward, update.Message.Chat.ID)
			case bonusStateAskAward:
				award, err := strconv.Atoi(update.Message.Text)
				if err != nil {
					tb.informUser(ctx, update.Message.From.ID, incorrectData)
					return
				}
				bonusScheme.award = award
				tb.createBonus(bonusScheme, update.Message)
				tb.f.Transition(update.Message.From.ID, stateDefault, update.Message.Chat.ID)

			case bonusStateDelete:
				bonusScheme.channel_name = update.Message.Text
				tb.deleteBonus(bonusScheme, update.Message)
				tb.f.Transition(update.Message.From.ID, stateDefault, update.Message.Chat.ID)
			// referral
			case referralStateDelete:
				refID, err := strconv.Atoi(update.Message.Text)
				if err != nil {
					tb.informUser(ctx, update.Message.From.ID, incorrectData)
					return
				}
				tb.deteleReferral(refID, update.Message)
				tb.f.Transition(update.Message.From.ID, stateDefault, update.Message.Chat.ID)
			}
		}
	}
}
