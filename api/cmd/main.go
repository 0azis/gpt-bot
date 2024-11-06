package main

import (
	"gpt-bot/api/db"
	"gpt-bot/api/server"
	"gpt-bot/config"
	"gpt-bot/tgbot"
	"log/slog"

	"github.com/labstack/echo/v4"
	"github.com/subosito/gotenv"
)

func main() {
	// load env file
	if err := gotenv.Load("../../.env"); err != nil {
		slog.Error("environment not found")
		return
	}

	// init config
	config := config.New()

	// init database
	store, err := db.New(config.Db.Addr())
	if err != nil {
		slog.Error("database connection was failed")
		return
	}

	// init and start bot
	bot, err := tgbot.New(config.Tokens.Telegram(), store, config.WebAppUrl)
	if err != nil {
		slog.Error("bot running failed")
		return
	}
	bot.InitHandlers()
	go bot.Start()

	// init and http server
	e := echo.New()

	// plug middlewares
	e.Use(server.AuthMiddleware)

	// init routes to it
	server.InitRoutes(e, store)

	err = e.Start(config.Server.Addr())
	if err != nil {
		slog.Error("server was failed")
	}
}
