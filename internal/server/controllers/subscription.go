package controllers

import (
	"context"
	"encoding/json"
	"fmt"
	"gpt-bot/internal/db"
	"gpt-bot/utils"

	"github.com/0azis/bot"
	"github.com/0azis/bot/models"
	"github.com/labstack/echo/v4"
)

type subscriptionControllers interface {
	CreateInvoiceLink(c echo.Context) error
}

type subscription struct {
	b *bot.Bot
}

func (s subscription) CreateInvoiceLink(c echo.Context) error {
	jwtUserID := utils.ExtractUserID(c)
	var paymentCredentials db.SubscriptionPaymentModel
	paymentCredentials.UserID = jwtUserID
	err := c.Bind(&paymentCredentials)
	if err != nil || !paymentCredentials.Valid() {
		return c.JSON(400, nil)
	}
	paymentCredentials.Rename()

	payload, err := json.Marshal(paymentCredentials)
	if err != nil {
		return c.JSON(500, nil)
	}

	link, err := s.b.CreateInvoiceLink(context.Background(), &bot.CreateInvoiceLinkParams{
		Title:       fmt.Sprintf("%s subscription", paymentCredentials.Name),
		Description: "Test",
		Payload:     string(payload),
		Currency:    "XTR",
		Prices: []models.LabeledPrice{
			{Label: paymentCredentials.Name, Amount: paymentCredentials.Amount},
		},
	})
	if err != nil {
		return c.JSON(500, nil)
	}

	return c.JSON(200, link)
}

func NewSubscriptionControllers(b *bot.Bot) subscriptionControllers {
	return subscription{
		b,
	}
}
