package tgbot

import (
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
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
	getTelegramAvatar(ctx context.Context, userID int64) string
	defaultHandler(ctx context.Context, b *bot.Bot, update *models.Update)
}

type tgBot struct {
	webAppUrl string
	store     db.Store
	b         *bot.Bot
}

func New(token string, store db.Store, url string) (BotInterface, error) {
	b, err := bot.New(token)
	return tgBot{
		webAppUrl: url,
		store:     store,
		b:         b,
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
	tg.b.SetDefaultHandler(tg.defaultHandler)
}

func (tb tgBot) startHandler(ctx context.Context, b *bot.Bot, update *models.Update) {
	var user db.UserModel
	user.ID = int(update.Message.From.ID)
	user.Avatar = tb.getTelegramAvatar(ctx, int64(user.ID))

	err := tb.store.User.Create(user)
	if err != nil {
		slog.Error(err.Error())
		tb.botError(ctx, update, "internal error while creating user")
		return
	}
	err = tb.store.Subscription.CreateStandardSubscription(user.ID)
	if err != nil {
		slog.Error(err.Error())
		tb.botError(ctx, update, "internal error while creating user")
		return
	}

	token, err := utils.SignJWT(user.ID)
	if err != nil {
		slog.Error(err.Error())
		tb.botError(ctx, update, "internal erorr while create jwt token")
		return
	}

	msgSlice := strings.Split(update.Message.Text, " ")
	if len(msgSlice) == 2 {
		refCode := &msgSlice[1]
		id, err := tb.store.User.IsUserReferred(user.ID, *refCode)
		if err != nil {
			tb.botError(ctx, update, "internal error")
			return
		}
		if id != 0 {
			tb.botError(ctx, update, "referral link can't be used twice")
			return
		}

		ownerID, err := tb.store.User.OwnerReferralCode(*refCode)
		if errors.Is(err, sql.ErrNoRows) {
			tb.botError(ctx, update, "referral link doesn't exists")
			return
		}
		if err != nil {
			slog.Error(err.Error())
			tb.botError(ctx, update, "internal error")
			return
		}

		award, err := tb.store.Bonus.GetReferralAward()
		if err != nil {
			slog.Error(err.Error())
			tb.botError(ctx, update, "internal error while realese ref link")
			return
		}
		err = tb.store.User.RaiseBalance(ownerID, award)
		if err != nil {
			slog.Error(err.Error())
			tb.botError(ctx, update, "internal error while raise referred balance")
			return
		}
		err = tb.store.User.SetReferredBy(user.ID, *refCode)
		if err != nil {
			slog.Error(err.Error())
			tb.botError(ctx, update, "internal error")
			return
		}
	}

	b.SetChatMenuButton(ctx, &bot.SetChatMenuButtonParams{
		ChatID: update.Message.Chat.ID,
		MenuButton: models.MenuButtonWebApp{
			Type: "web_app",
			Text: "Open Mini App",
			WebApp: models.WebAppInfo{
				URL: tb.webAppUrl + "?" + token,
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

func (tb tgBot) botError(ctx context.Context, update *models.Update, errorMsg string) {
	tb.b.SendMessage(ctx, &bot.SendMessageParams{
		ChatID: update.Message.Chat.ID,
		Text:   fmt.Sprintf("Error: %s", errorMsg),
	})
}

func (tb tgBot) defaultHandler(ctx context.Context, b *bot.Bot, update *models.Update) {
	if update.PreCheckoutQuery != nil {
		fmt.Println(update.PreCheckoutQuery)
		fmt.Println(update.PreCheckoutQuery.InvoicePayload)
		var subscriptionPayload db.SubscriptionPaymentModel
		err := json.Unmarshal([]byte(update.PreCheckoutQuery.InvoicePayload), &subscriptionPayload)
		if err != nil {
			return
		}
		err = tb.store.Subscription.UpdateSubscription(subscriptionPayload.UserID, subscriptionPayload.Name, subscriptionPayload.End)
		if err != nil {
			slog.Error(err.Error())
			return
		}
		diamonds, err := tb.store.Subscription.SubscriptionInfo(subscriptionPayload.Name)
		if err != nil {
			slog.Error(err.Error())
			return
		}
		err = tb.store.User.FillBalance(subscriptionPayload.UserID, diamonds)
		if err != nil {
			slog.Error(err.Error())
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
			b.SendMessage(ctx, &bot.SendMessageParams{
				ChatID: update.Message.Chat.ID,
				Text:   "Payment was done!",
			})
			return
		}
	}
}
