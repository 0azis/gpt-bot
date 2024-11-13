package tgbot

import (
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	"gpt-bot/config"
	"gpt-bot/internal/db"
	"gpt-bot/utils"
	"log/slog"
	"os"
	"os/signal"
	"strings"

	"github.com/0azis/bot"
	"github.com/0azis/bot/models"
)

type BotInterface interface {
	Start()
	Instance() *bot.Bot
	InitHandlers()

	// handlers
	startHandler(ctx context.Context, b *bot.Bot, update *models.Update)
	defaultHandler(ctx context.Context, b *bot.Bot, update *models.Update)

	// helpers
	getTelegramAvatar(ctx context.Context, userID int64) string
	informUser(ctx context.Context, userID int64, errMsg string)
}

type tgBot struct {
	telegram config.Telegram
	store    db.Store
	b        *bot.Bot
}

func New(cfg config.Telegram, store db.Store) (BotInterface, error) {
	b, err := bot.New(cfg.GetToken())
	return tgBot{
		telegram: cfg,
		store:    store,
		b:        b,
	}, err
}

func (tb tgBot) Start() {
	ctx, cancel := signal.NotifyContext(context.Background(), os.Interrupt)
	defer cancel()
	tb.b.Start(ctx)
}

func (tb tgBot) Instance() *bot.Bot {
	return tb.b
}

func (tg tgBot) InitHandlers() {
	tg.b.RegisterHandler(bot.HandlerTypeMessageText, "/start", bot.MatchTypeContains, tg.startHandler)
	tg.b.RegisterHandler(bot.HandlerTypeMessageText, "/admin", bot.MatchTypeContains, tg.adminHandler)
	tg.b.SetDefaultHandler(tg.defaultHandler)
}

func (tb tgBot) startHandler(ctx context.Context, b *bot.Bot, update *models.Update) {
	var user db.UserModel
	user.ID = int(update.Message.From.ID)
	user.Avatar = tb.getTelegramAvatar(ctx, int64(user.ID))

	err := tb.store.User.Create(user)
	if err != nil {
		slog.Error(err.Error())
		tb.informUser(ctx, int64(user.ID), "")
		return
	}
	err = tb.store.Subscription.InitStandard(user.ID)
	if err != nil {
		slog.Error(err.Error())
		tb.informUser(ctx, int64(user.ID), "")
		return
	}

	token := utils.NewToken()
	token.SetUserID(user.ID)
	err = token.SignJWT()
	if err != nil {
		slog.Error(err.Error())
		tb.informUser(ctx, int64(user.ID), "")
		return
	}

	msgSlice := strings.Split(update.Message.Text, " ")
	if len(msgSlice) == 2 {
		refCode := &msgSlice[1]
		id, err := tb.store.User.IsUserReferred(user.ID, *refCode)
		if err != nil {
			tb.informUser(ctx, int64(user.ID), "")
			return
		}
		if id != 0 {
			tb.informUser(ctx, int64(user.ID), "")
			return
		}

		ownerID, err := tb.store.User.OwnerOfReferralCode(*refCode)
		if errors.Is(err, sql.ErrNoRows) {
			tb.informUser(ctx, int64(user.ID), "")
			return
		}
		if err != nil {
			slog.Error(err.Error())
			tb.informUser(ctx, int64(user.ID), "")
			return
		}

		award, err := tb.store.Bonus.GetAward("referral")
		if err != nil {
			slog.Error(err.Error())
			tb.informUser(ctx, int64(user.ID), "")
			return
		}
		err = tb.store.User.RaiseBalance(ownerID, award)
		if err != nil {
			slog.Error(err.Error())
			tb.informUser(ctx, int64(user.ID), "")
			return
		}
		err = tb.store.User.SetReferredBy(user.ID, *refCode)
		if err != nil {
			slog.Error(err.Error())
			tb.informUser(ctx, int64(user.ID), "")
			return
		}
	}

	b.SetChatMenuButton(ctx, &bot.SetChatMenuButtonParams{
		ChatID: update.Message.Chat.ID,
		MenuButton: models.MenuButtonWebApp{
			Type: "web_app",
			Text: "Open Mini App",
			WebApp: models.WebAppInfo{
				URL: tb.telegram.GetWebAppUrl() + "?" + token.GetStrToken(),
			},
		},
	})

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

func (tb tgBot) informUser(ctx context.Context, userID int64, errorMsg string) {
	tb.b.SendMessage(ctx, &bot.SendMessageParams{
		ChatID: userID,
		Text:   fmt.Sprintf("Error: %s", errorMsg),
	})
}

func (tb tgBot) defaultHandler(ctx context.Context, b *bot.Bot, update *models.Update) {
	var subscriptionPayload db.SubscriptionModel

	if update.PreCheckoutQuery != nil {
		err := json.Unmarshal([]byte(update.PreCheckoutQuery.InvoicePayload), &subscriptionPayload)
		if err != nil {
			return
		}
		err = tb.store.Subscription.Update(subscriptionPayload.UserID, subscriptionPayload.Name, subscriptionPayload.End)
		if err != nil {
			slog.Error(err.Error())
			tb.informUser(ctx, int64(subscriptionPayload.UserID), "")
			return
		}
		diamonds, err := tb.store.Subscription.DailyDiamonds(subscriptionPayload.Name)
		if err != nil {
			slog.Error(err.Error())
			tb.informUser(ctx, int64(subscriptionPayload.UserID), "")
			return
		}
		err = tb.store.User.FillBalance(subscriptionPayload.UserID, diamonds)
		if err != nil {
			slog.Error(err.Error())
			tb.informUser(ctx, int64(subscriptionPayload.UserID), "")
			return
		}

		b.AnswerPreCheckoutQuery(ctx, &bot.AnswerPreCheckoutQueryParams{
			PreCheckoutQueryID: update.PreCheckoutQuery.ID,
			OK:                 true,
			ErrorMessage:       "",
		})
		return
	}

	if update.Message != nil {
		if update.Message.SuccessfulPayment != nil {
			err := json.Unmarshal([]byte(update.Message.SuccessfulPayment.InvoicePayload), &subscriptionPayload)
			if err != nil {
				return
			}
			b.SendMessage(ctx, &bot.SendMessageParams{
				ChatID: update.Message.Chat.ID,
				Text:   fmt.Sprintf("Вы успешно купили подписку **%s** за *%d* звезд!\nВаша подписка доступна до: %s", subscriptionPayload.Name, subscriptionPayload.Amount, subscriptionPayload.End),
			})
			return
		}
	}
}

func (tb tgBot) adminHandler(ctx context.Context, b *bot.Bot, update *models.Update) {
	msgSlice := strings.Split(update.Message.Text, " ")
	if len(msgSlice) == 1 {
		tb.informUser(ctx, update.Message.From.ID, "")
		return
	}
	password := msgSlice[1]
	if tb.telegram.GetAdminPassword() != password {
		tb.informUser(ctx, update.Message.From.ID, "")
	}

	b.SendMessage(ctx, &bot.SendMessageParams{
		ChatID: update.Message.From.ID,
		Text:   "You are an admin!",
	})
}
