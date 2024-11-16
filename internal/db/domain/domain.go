package domain

import (
	"gpt-bot/utils"
	"slices"
	"strings"

	"github.com/sashabaranov/go-openai"
)

type User struct {
	ID           int          `json:"id" db:"id"`
	Subscription Subscription `json:"subscription"`
	Avatar       string       `json:"avatar"`
	Balance      int          `json:"balance"`
	ReferralCode *string      `json:"referralCode"`
	ReferredBy   *string      `json:"referredBy"`
}

type Subscription struct {
	UserId int     `json:"-"`
	Name   string  `json:"name"`
	Start  string  `json:"start"`
	End    *string `json:"end"`
}

var paymentTypes []string = []string{"stars", "crypto"}
var paymentAsset []string = []string{"USDT", "TON", "BTC", "ETH", "LTC", "BNB", "TRX", "USDC"}
var paymentPrices map[string]int = map[string]int{
	"advanced-month": 1,
	"advanced-year":  1,
	"ultimate-month": 1,
	"ultimate-year":  1,
}

type Payment struct {
	UserID           int    `json:"userId"`
	SubscriptionName string `json:"name"`
	Type             string `json:"type"`
	Asset            string `json:"asset"`
	Amount           int    `json:"amount"`
	End              string `json:"end"`
}

func (p Payment) Valid() bool {
	if p.UserID == 0 || p.Amount == 0 {
		return false
	}
	if !slices.Contains(paymentAsset, p.Asset) {
		return false
	}
	if !slices.Contains(paymentTypes, p.Type) {
		return false
	}
	if paymentPrices[p.SubscriptionName] != p.Amount {
		return false
	}

	return true
}

func (p *Payment) ToReadable() {
	oldName := p.SubscriptionName
	newName := strings.Split(oldName, "-")
	p.SubscriptionName = newName[0]

	if newName[1] == "month" {
		p.End = utils.AddMonth()
	}
	if newName[1] == "year" {
		p.End = utils.AddYear()
	}
}

const PriceOfMessage int = 10

type Message struct {
	ID      int    `json:"-"`
	ChatID  int    `json:"-"`
	Content string `json:"content"`
	Role    string `json:"role"`
}

func (m Message) Valid() bool {
	return m.Content != ""
}

func NewUserMessage(chatID int, content string) Message {
	return Message{
		ChatID:  chatID,
		Content: content,
		Role:    openai.ChatMessageRoleUser,
	}
}

func NewAssistantMessage(chatID int, content string) Message {
	return Message{
		ChatID:  chatID,
		Content: content,
		Role:    openai.ChatMessageRoleAssistant,
	}
}

type chatType string

var (
	ChatText  chatType = "text"
	ChatImage chatType = "image"
)

var modelNames = map[chatType][]string{
	ChatText:  []string{"o1-preview", "gpt-4o", "o1-mini", "gpt-4o-mini"},
	ChatImage: []string{"dall-e-3", "runware"},
}

type Chat struct {
	ID     int      `json:"id" db:"id"`
	UserID int      `json:"-"`
	Title  *string  `json:"title" db:"title"`
	Model  string   `json:"-" db:"model"`
	Type   chatType `json:"type" db:"type"`
}

func (c *Chat) SetType() bool {
	for t, models := range modelNames {
		for _, model := range models {
			if model == c.Model {
				c.Type = t
				break
			}
		}
	}
	return c.Type != ""
}

type bonusType string

const (
	BonusReferral bonusType = "referral"
)

type Bonus struct {
	ID        int       `json:"id" db:"id"`
	Award     int       `json:"award" db:"award"`
	BonusType bonusType `json:"bonusType" db:"bonus_type"`
}
