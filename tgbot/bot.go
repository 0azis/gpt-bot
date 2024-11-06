package tgbot

import (
	"context"
	"gpt-bot/internal/db"
	"gpt-bot/utils"
	"os"
	"os/signal"

	"github.com/go-telegram/bot"
	"github.com/go-telegram/bot/models"
)

type BotInterface interface {
	Start()
	InitHandlers()

	// handlers
	startHandler(ctx context.Context, b *bot.Bot, update *models.Update)
	getTelegramAvatar(ctx context.Context, userID int64) string
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

func (tg tgBot) InitHandlers() {
	tg.b.RegisterHandler(bot.HandlerTypeMessageText, "/start", bot.MatchTypeExact, tg.startHandler)
}

func (tb tgBot) startHandler(ctx context.Context, b *bot.Bot, update *models.Update) {
	userID := update.Message.From.ID
	err := tb.store.User.Create(int(userID), tb.getTelegramAvatar(ctx, userID))
	if err != nil {
		b.SendMessage(ctx, &bot.SendMessageParams{
			ChatID: update.Message.Chat.ID,
			Text:   "Error",
		})
		return
	}

	token, err := utils.SignJWT(int(userID))
	if err != nil {
		b.SendMessage(ctx, &bot.SendMessageParams{
			ChatID: update.Message.Chat.ID,
			Text:   "Error",
		})
		return
	}

	b.SetChatMenuButton(ctx, &bot.SetChatMenuButtonParams{
		ChatID: update.Message.Chat.ID,
		MenuButton: models.MenuButtonWebApp{
			Type: "web_app",
			Text: "Open App",
			WebApp: models.WebAppInfo{
				URL: tb.webAppUrl + "?" + token,
			},
		},
	})

	b.SendMessage(ctx, &bot.SendMessageParams{
		ChatID: update.Message.Chat.ID,
		Text:   "Hello!",
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
