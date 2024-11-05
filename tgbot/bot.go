package tgbot

import (
	"context"
	"os"
	"os/signal"

	"github.com/go-telegram/bot"
	"github.com/go-telegram/bot/models"
)

type BotInterface interface {
	Start()
}

type tgBot struct {
	b *bot.Bot
}

func New(token string) (BotInterface, error) {
	opts := []bot.Option{
		bot.WithDefaultHandler(handler),
	}
	b, err := bot.New(token, opts...)
	return tgBot{
		b: b,
	}, err
}

func (tb tgBot) Start() {
	ctx, cancel := signal.NotifyContext(context.Background(), os.Interrupt)
	defer cancel()
	tb.b.Start(ctx)
}

func handler(ctx context.Context, b *bot.Bot, update *models.Update) {
	b.SendMessage(ctx, &bot.SendMessageParams{
		ChatID: update.Message.Chat.ID,
		Text:   update.Message.Text,
	})
}
