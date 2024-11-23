package api

import (
	"log/slog"
	"strconv"

	"github.com/arthurshafikov/cryptobot-sdk-golang/cryptobot"
)

type cryptoInterface interface {
	CreateInvoiceLink(invoice cryptobot.CreateInvoiceRequest) (string, error)
	NewInvoiceModel(amount int, asset string, payload string) invoiceModel
	Webhook(requestBody []byte) ([]byte, error)
}

type cryptoClient struct {
	*cryptobot.Client
}

func newCrypto(token string) cryptoInterface {
	client := cryptobot.NewClient(cryptobot.Options{
		APIToken: token,
	})
	return cryptoClient{client}
}

type invoiceModel struct {
	Amount    string `json:"amount"`
	Asset     string `json:"asset"`
	Payload   string `json:"payload"`
	ExpiresIn int    `json:"-"`
}

func (c cryptoClient) NewInvoiceModel(amount int, asset string, payload string) invoiceModel {
	amountStr := strconv.Itoa(amount)
	return invoiceModel{
		Amount:    amountStr,
		Asset:     asset,
		Payload:   payload,
		ExpiresIn: 30,
	}
}

func (c cryptoClient) CreateInvoiceLink(invoice cryptobot.CreateInvoiceRequest) (string, error) {
	invoiceObj, err := c.CreateInvoice(invoice)
	if err != nil {
		return "", err
	}

	return invoiceObj.PayUrl, err
}

func (c cryptoClient) Webhook(requestBody []byte) ([]byte, error) {
	webhookUpdate, err := cryptobot.ParseWebhookUpdate(requestBody)
	if err != nil {
		slog.Error(err.Error())
		return []byte{}, err
	}

	switch webhookUpdate.UpdateType {
	case cryptobot.InvoicePaidWebhookUpdateType:
		invoice, err := cryptobot.ParseInvoice(requestBody)
		if err != nil {
			slog.Error(err.Error())
			return []byte{}, err
		}
		return []byte(invoice.Payload), nil
	default:
		return []byte{}, nil
	}
}
