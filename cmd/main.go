package main

import (
	"gpt-bot/config"
	"gpt-bot/cron"
	"gpt-bot/internal/api"
	"gpt-bot/internal/db"
	"gpt-bot/internal/server"
	"gpt-bot/tgbot"
	"log/slog"

	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"
	"github.com/subosito/gotenv"
)

func main() {
	// load env file
	if err := gotenv.Load("../.env"); err != nil {
		slog.Error("environment not found")
		return
	}

	// init config
	config := config.New()
	if !config.IsValid() {
		slog.Error("environment data isn't full")
		return
	}

	// init database
	store, err := db.New(config.Database)
	if err != nil {
		slog.Error(err.Error())
		slog.Error("database connection was failed")
		return
	}

	// init and start bot
	bot, err := tgbot.New(config.Telegram, store)
	if err != nil {
		slog.Error("bot running failed")
		return
	}
	bot.InitHandlers()
	go bot.Start()

	// init and http server
	e := echo.New()
	// plug middlewares
	e.Use(middleware.CORSWithConfig(middleware.CORSConfig{
		AllowOrigins:     []string{"*"},
		AllowHeaders:     []string{echo.HeaderOrigin, echo.HeaderContentType, echo.HeaderAccept, echo.HeaderAuthorization, echo.HeaderContentLength},
		AllowMethods:     []string{echo.GET, echo.POST, echo.PUT, echo.DELETE, echo.OPTIONS},
		AllowCredentials: true,
	}))

	// init api
	api := api.New(config.Api)
	_ = api.YooKassa.CreatePayment("4100118919927024.D44A54E351E103EE25C66F386FAD9AEEF4A46235E7F4BBECFD7EEA366CD8A58B0B924D9F6296D1AE2D6C58C0CC60F34119EEB8C923CE1B255252E4733768E8A64497F74F735E25C84EF45BC5A51748235B53FCF49A9C1F7574A6AA299D8D8F34FAA9BC9334FAA61721D68DE9D76AA40D4407DF4EFEB3EA40E595F8D063486D34", 10)

	// init cron manager
	cronManager := cron.Init(store)
	cronManager.Run()

	// init routes to it
	server.InitRoutes(e, store, api, bot, config.Server.SavePath())

	err = e.Start(config.Server.Addr())
	if err != nil {
		slog.Error("server was failed")
	}
}
