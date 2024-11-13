package controllers

import (
	"encoding/json"
	"gpt-bot/internal/api"
	"gpt-bot/internal/db"
	"gpt-bot/tgbot"
	"gpt-bot/utils"
	"io"
	"log/slog"
	"strconv"

	"github.com/arthurshafikov/cryptobot-sdk-golang/cryptobot"
	"github.com/labstack/echo/v4"
)

type subscriptionControllers interface {
	CreateInvoiceLink(c echo.Context) error
	Webhook(c echo.Context) error
}

type subscription struct {
	api api.Interface
	b   tgbot.BotInterface
}

func (s subscription) CreateInvoiceLink(c echo.Context) error {
	jwtUserID := utils.ExtractUserID(c)

	var paymentCredentials db.SubscriptionModel
	paymentCredentials.UserID = jwtUserID
	err := c.Bind(&paymentCredentials)
	if err != nil || !paymentCredentials.Valid() {
		return c.JSON(400, nil)
	}
	paymentCredentials.ToReadable()

	payload, err := json.Marshal(paymentCredentials)
	if err != nil {
		slog.Error(err.Error())
		return c.JSON(500, nil)
	}

	switch paymentCredentials.Type {
	case "stars":
		link, err := s.b.CreateInvoiceLink(payload, paymentCredentials)
		if err != nil {
			slog.Error(err.Error())
			return c.JSON(500, nil)
		}
		return c.JSON(200, link)
	case "crypto":
		amountStr := strconv.Itoa(paymentCredentials.Amount)
		link, err := s.api.Crypto.CreateInvoiceLink(cryptobot.CreateInvoiceRequest{
			Amount:    amountStr,
			Asset:     paymentCredentials.Asset,
			Payload:   string(payload),
			ExpiresIn: 30,
		})
		if err != nil {
			slog.Error(err.Error())
			return c.JSON(500, nil)
		}
		return c.JSON(200, link)
	default:
		return c.JSON(400, nil)
	}
}

func (s subscription) Webhook(c echo.Context) error {
	req := c.Request()
	body, err := io.ReadAll(req.Body)
	if err != nil {
		slog.Error(err.Error())
	}
	defer req.Body.Close()

	payload, err := s.api.Crypto.Webhook(body)
	if err != nil {
		slog.Error(err.Error())
		return c.JSON(500, nil)
	}
	var paymentCredentials db.SubscriptionModel
	err = json.Unmarshal(payload, &paymentCredentials)
	if err != nil {
		slog.Error(err.Error())
		return c.JSON(500, nil)
	}

	s.b.PaymentInfo(paymentCredentials)
	return c.JSON(200, nil)
}

func NewSubscriptionControllers(b tgbot.BotInterface, api api.Interface) subscriptionControllers {
	return subscription{
		api, b,
	}
}
