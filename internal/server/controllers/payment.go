package controllers

import (
	"encoding/json"
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
	AuthYooMoney(c echo.Context) error
	TokenYooMoney(c echo.Context) error
	CreatePayment(c echo.Context) error
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

	switch payment.Entity {
	case "subscription":
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
				Asset:     "USDT",
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

	case "limits":
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
				Asset:     "USDT",
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
	return c.JSON(400, nil)
}

func (p payment) AuthYooMoney(c echo.Context) error {
	return c.JSON(200, p.api.YooKassa.GetAuthLink())
}

func (p payment) TokenYooMoney(c echo.Context) error {
	code := struct {
		Value string `json:"code"`
	}{}
	if err := c.Bind(&code); err != nil {
		return c.JSON(400, nil)
	}

	string := p.api.YooKassa.GetToken(code.Value)
	return c.JSON(200, string)
}

func (p payment) CreatePayment(c echo.Context) error {
	jwtUserID := utils.ExtractUserID(c)
	var payment domain.Payment
	if err := c.Bind(&payment); err != nil {
		return c.JSON(400, nil)
	}
	status := p.api.YooKassa.CreatePayment("", payment.Amount)
	if !status {
		return c.JSON(500, nil)
	}

	err := p.store.Limits.AddLimits(jwtUserID, payment.LimitModel, payment.Amount)
	if err != nil {
		return c.JSON(500, nil)
	}

	return c.JSON(200, nil)
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
	switch payment.Entity {
	case "subscription":
		err = p.store.Subscription.Update(payment.UserID, payment.SubscriptionName, payment.SubscriptionEnd)
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
	case "limits":
		err = p.store.Limits.AddLimits(payment.UserID, payment.LimitModel, payment.LimitAmount)
		if err != nil {
			slog.Error(err.Error())
			p.b.PaymentInfo(payment, false)
			return c.JSON(500, nil)
		}

		p.b.PaymentInfo(payment, true)
		return c.JSON(200, nil)
	}
	return c.JSON(400, nil)
}

func NewPaymentControllers(store db.Store, b tgbot.BotInterface, api api.Interface) paymentControllers {
	return payment{
		store, api, b,
	}
}
