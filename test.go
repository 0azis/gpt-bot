package main

import (
	"fmt"
	"gpt-bot/config"
	"log"
	"log/slog"

	"github.com/arthurshafikov/cryptobot-sdk-golang/cryptobot"
	"github.com/subosito/gotenv"
)

func main() {
	if err := gotenv.Load("./.env"); err != nil {
		slog.Error(err.Error())
	}
	cfg := config.New()
	fmt.Println(cfg)
	client := cryptobot.NewClient(cryptobot.Options{
		APIToken: cfg.Api.CryptoBot(),
	})

	appInfo, err := client.GetMe()
	if err != nil {
		log.Fatalln(err)
	}

	fmt.Printf(
		"AppID - %v, Name - %s, PaymentProcessingBotUsername - %s \n",
		appInfo.AppID,
		appInfo.Name,
		appInfo.PaymentProcessingBotUsername,
	)
}
