package tgbot

import (
	"context"
	"gpt-bot/api/db"
	"os"
	"os/signal"

	"github.com/go-telegram/bot"
	"github.com/go-telegram/bot/models"
)

type BotInterface interface {
	Start()
}

type tgBot struct {
	webAppUrl string
	store     db.Store
	b         *bot.Bot
}

func New(token string, store db.Store, url string) (BotInterface, error) {
	// opts := []bot.Option{
	// 	bot.WithDefaultHandler(handler),
	// }
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

func (tb tgBot) startHandler(ctx context.Context, b *bot.Bot, update *models.Update) {
	err := tb.store.User.Create(int(update.Message.From.ID))
	if err != nil {
		b.SendMessage(ctx, &bot.SendMessageParams{
			ChatID: update.Message.Chat.ID,
			Text:   "Error",
		})
	}
	b.SetChatMenuButton(ctx, &bot.SetChatMenuButtonParams{
		ChatID: update.Message.Chat.ID,
		MenuButton: models.MenuButtonWebApp{
			Type: "web_app",
			Text: "Open App",
			WebApp: models.WebAppInfo{
				URL: tb.webAppUrl,
			},
		},
	})

	b.SendMessage(ctx, &bot.SendMessageParams{
		ChatID: update.Message.Chat.ID,
		Text:   "Hello!",
	})
}
