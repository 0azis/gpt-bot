package controllers

import (
	"encoding/json"
	"fmt"
	"gpt-bot/internal/api"
	"gpt-bot/internal/db"
	"gpt-bot/internal/db/domain"
	"gpt-bot/tgbot"
	"gpt-bot/utils"
	"io"
	"log/slog"
	"strconv"

	"github.com/arthurshafikov/cryptobot-sdk-golang/cryptobot"
	"github.com/labstack/echo/v4"
)

type paymentControllers interface {
	CreateInvoiceLink(c echo.Context) error
	Webhook(c echo.Context) error
	YooMoneyWebhook(c echo.Context) error
}

type payment struct {
	store db.Store
	api   api.Interface
	b     tgbot.BotInterface
}

func (p payment) CreateInvoiceLink(c echo.Context) error {
	jwtUserID := utils.ExtractUserID(c)

	var payment domain.Payment
	payment.UserID = jwtUserID
	err := c.Bind(&payment)
	if err != nil || !payment.Valid() {
		return c.JSON(400, nil)
	}
	payment.ToReadable()

	id, err := p.store.Subscription.UserSubscription(int64(jwtUserID), payment.SubscriptionName)
	if err != nil {
		return c.JSON(500, nil)
	}
	if id != 0 {
		return c.JSON(400, nil)
	}

	payload, err := json.Marshal(payment)
	if err != nil {
		slog.Error(err.Error())
		return c.JSON(500, nil)
	}

	switch payment.Type {
	case "telegram":
		link, err := p.b.CreateInvoiceLink(payload, payment)
		if err != nil {
			slog.Error(err.Error())
			return c.JSON(500, nil)
		}
		return c.JSON(200, link)
	case "crypto":
		amountStr := strconv.Itoa(payment.Amount)
		link, err := p.api.Crypto.CreateInvoiceLink(cryptobot.CreateInvoiceRequest{
			Amount:    amountStr,
			Asset:     payment.Asset,
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

func (p payment) Webhook(c echo.Context) error {
	req := c.Request()
	body, err := io.ReadAll(req.Body)
	if err != nil {
		slog.Error(err.Error())
	}
	defer req.Body.Close()

	payload, err := p.api.Crypto.Webhook(body)
	if err != nil {
		slog.Error(err.Error())
		return c.JSON(500, nil)
	}

	var payment domain.Payment
	err = json.Unmarshal(payload, &payment)
	if err != nil {
		slog.Error(err.Error())
		return c.JSON(500, nil)
	}

	err = p.store.Subscription.Update(payment.UserID, payment.SubscriptionName, payment.End)
	if err != nil {
		slog.Error(err.Error())
		p.b.PaymentInfo(payment, false)
		return c.JSON(500, nil)
	}
	diamonds, err := p.store.Subscription.DailyDiamonds(payment.SubscriptionName)
	if err != nil {
		slog.Error(err.Error())
		p.b.PaymentInfo(payment, false)
		return c.JSON(500, nil)
	}
	err = p.store.User.FillBalance(payment.UserID, diamonds)
	if err != nil {
		slog.Error(err.Error())
		p.b.PaymentInfo(payment, false)
		return c.JSON(500, nil)
	}
	limits := domain.NewLimits(payment.UserID, payment.SubscriptionName)
	err = p.store.Limits.Update(limits)
	if err != nil {
		slog.Error(err.Error())
		p.b.PaymentInfo(payment, false)
		return c.JSON(500, nil)
	}

	p.b.PaymentInfo(payment, true)
	return c.JSON(200, nil)
}

func (p payment) YooMoneyWebhook(c echo.Context) error {
	b, err := io.ReadAll(c.Request().Body)
	fmt.Println(string(b), err)
	fmt.Println(c.Request())
	return c.JSON(200, nil)
}

func NewPaymentControllers(store db.Store, b tgbot.BotInterface, api api.Interface) paymentControllers {
	return payment{
		store, api, b,
	}
}
