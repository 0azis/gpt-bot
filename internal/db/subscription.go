package db

import (
	"fmt"
	"gpt-bot/utils"
	"strings"

	"github.com/jmoiron/sqlx"
)

var subscriptions map[string]int = map[string]int{
	"advanced-month": 1,
	"advanced-year":  1,
	"ultimate-month": 1,
	"ultimate-year":  1,
}

type subscriptionRepository interface {
	InitStandard(userID int) error
	Update(userID int, name string, end string) error
	Info(name string) (int, error)
}

type subscription struct {
	db *sqlx.DB
}

type SubscriptionModel struct {
	UserID int    `json:"userId"`
	Name   string `json:"name"`
	Amount int    `json:"amount"`
	End    string `json:"end"`
}

func (sc SubscriptionModel) Valid() bool {
	fmt.Println(sc)
	if _, ok := subscriptions[sc.Name]; !ok {
		return false
	}
	if subscriptions[sc.Name] != sc.Amount {
		return false
	}
	if sc.UserID == 0 {
		return false
	}
	return true
}

func (sc *SubscriptionModel) ToReadable() {
	oldName := sc.Name
	newName := strings.Split(oldName, "-")
	sc.Name = newName[0]

	if newName[1] == "month" {
		sc.End = utils.AddMonth()
	}
	if newName[1] == "year" {
		sc.End = utils.AddYear()
	}
}

func (s subscription) InitStandard(userID int) error {
	_, err := s.db.Query(`insert into subscriptions (user_id) values (?) on duplicate key update user_id = user_id`, userID)
	return err
}

func (s subscription) Update(userID int, name string, end string) error {
	_, err := s.db.Query(`update subscriptions set name = ?, end = ? where user_id = ?`, name, end, userID)
	return err
}

func (s subscription) Info(name string) (int, error) {
	var diamonds int
	err := s.db.Get(&diamonds, `select diamonds from subscriptions_info where name = ?`, name)
	return diamonds, err
}
