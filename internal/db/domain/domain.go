package domain

import (
	"gpt-bot/utils"
	"slices"
	"strings"

	"github.com/sashabaranov/go-openai"
)

const (
	SubscriptionStandard = "standard"
	SubscriptionAdvanced = "advanced"
	SubscriptionUltimate = "ultimate"
)

var subscriptionModels map[string][]string = map[string][]string{
	SubscriptionStandard: []string{"gpt-4o-mini", "runware"},
	SubscriptionAdvanced: []string{"gpt-4o", "o1-mini", "gpt-4o-mini", "runware", "dall-e-3"},
	SubscriptionUltimate: []string{"o1-preview", "gpt-4o", "o1-mini", "gpt-4o-mini", "runware", "dall-e-3"},
}

type User struct {
	ID           int          `json:"id" db:"id"`
	Subscription Subscription `json:"subscription"`
	Limits       Limits       `json:"limits"`
	Avatar       string       `json:"avatar"`
	Balance      int          `json:"balance"`
	LanguageCode string       `json:"language_code"`
	ReferralCode *string      `json:"referralCode"`
	ReferredBy   *string      `json:"referredBy"`
	Username     string
	CreatedAt    string `json:"createdAt"`
}

func (u User) IsModelValid(model string) bool {
	if !slices.Contains(subscriptionModels[u.Subscription.Name], model) {
		return false
	}
	return true
}

type Subscription struct {
	UserId int     `json:"-"`
	Name   string  `json:"name"`
	Start  string  `json:"start"`
	End    *string `json:"end"`
}

type Limits struct {
	UserID    int `json:"-" db:"user_id"`
	O1Preview int `json:"o1-preview" db:"o1_preview"`
	Gpt4o     int `json:"gpt-4o" db:"gpt_4o"`
	O1Mini    int `json:"o1-mini" db:"o1_mini"`
	Gpt4oMini int `json:"gpt-4o-mini" db:"gpt_4o_mini"`
	Runware   int `json:"runware" db:"runware"`
	Dalle3    int `json:"dall-e-3" db:"dall_e_3"`
}

type LimitsMap map[string]int

func NewLimits(userID int, subscription string) Limits {
	if subscription == SubscriptionStandard {
		return Limits{
			UserID:    userID,
			Gpt4oMini: 5,
			Runware:   1,
		}
	}
	if subscription == SubscriptionAdvanced {
		return Limits{
			UserID:    userID,
			Gpt4o:     50,
			O1Mini:    50,
			Gpt4oMini: 50,
			Runware:   10,
			Dalle3:    10,
		}
	}
	if subscription == SubscriptionUltimate {
		return Limits{
			UserID:    userID,
			O1Preview: 50,
			Gpt4o:     50,
			O1Mini:    50,
			Gpt4oMini: 50,
			Runware:   30,
			Dalle3:    30,
		}
	}
	return Limits{}
}

// var paymentTypes []string = []string{"stars", "crypto"}
var paymentAsset map[string][]string = map[string][]string{
	"crypto":   []string{"USDT", "TON", "BTC", "ETH", "LTC", "BNB", "TRX", "USDC"},
	"telegram": []string{"stars"},
}
var paymentPrices map[string]map[string]int = map[string]map[string]int{
	"advanced-month": map[string]int{"telegram": 379, "crypto": 3},
	"advanced-year":  map[string]int{"telegram": 3299, "crypto": 30},
	"ultimate-month": map[string]int{"telegram": 1279, "crypto": 13},
	"ultimate-year":  map[string]int{"telegram": 10999, "crypto": 104},
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
	if p.UserID == 0 {
		return false
	}
	if !slices.Contains(paymentAsset[p.Type], p.Asset) {
		return false
	}
	if _, ok := paymentPrices[p.SubscriptionName]; !ok {
		return false
	}

	return true
}

func (p *Payment) ToReadable() {
	p.Amount = paymentPrices[p.SubscriptionName][p.Type]

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

const PriceOfTextMessage int = 10
const PriceOfImageMessage int = 100

type Message struct {
	ID      int    `json:"-"`
	ChatID  int    `json:"-"`
	Content string `json:"content"`
	Role    string `json:"role"`
	Type    string `json:"type"`
}

func (m Message) Valid() bool {
	return m.Content != ""
}

func NewUserTextMessage(chatID int, content string) Message {
	return Message{
		ChatID:  chatID,
		Content: content,
		Role:    openai.ChatMessageRoleUser,
		Type:    "text",
	}
}

func NewUserImageMessage(chatID int, content string) Message {
	return Message{
		ChatID:  chatID,
		Content: content,
		Role:    openai.ChatMessageRoleUser,
		Type:    "image",
	}
}

func NewAssistantTextMessage(chatID int, content string) Message {
	return Message{
		ChatID:  chatID,
		Content: content,
		Role:    openai.ChatMessageRoleAssistant,
		Type:    "text",
	}
}

func NewAssistantImageMessage(chatID int, content string) Message {
	return Message{
		ChatID:  chatID,
		Content: content,
		Role:    openai.ChatMessageRoleUser,
		Type:    "image",
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

const ReferralAward = 10

type Bonus struct {
	ID        int     `json:"id" db:"id"`
	Name      string  `json:"-" db:"name"`
	Channel   Channel `json:"channel"`
	Award     int     `json:"award"`
	Completed bool    `json:"completed"`
	Awarded   bool    `json:"awarded"`
	Link      string  `json:"link" db:"link"`
	Check     bool    `json:"-" db:"is_check"`
	MaxUsers  int     `json:"-" db:"max_users"`
	CreatedAt string  `json:"-" db:"created_at"`
}

func (b Bonus) Valid() bool {
	return b.Award != 0 || b.Channel.ID != 0
}

type Channel struct {
	ID     int    `json:"id"`
	Title  string `json:"title"`
	Avatar string `json:"avatar"`
}

const baseUrl = "https://t.me/webai_robot?start="

type Referral struct {
	ID   int    `db:"id"`
	Code string `db:"code"`
	Link string
}

func (r *Referral) SetLink() {
	r.Link = baseUrl + r.Code
}
