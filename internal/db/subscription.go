package db

import (
	"gpt-bot/utils"

	"github.com/jmoiron/sqlx"
)

type subscriptionDb struct {
	db *sqlx.DB
}

func (s subscriptionDb) InitStandard(userID int) error {
	rows, err := s.db.Query(`insert into subscriptions (user_id) values (?) on duplicate key update user_id = user_id`, userID)
	defer rows.Close()
	return err
}

func (s subscriptionDb) UserSubscription(userID int64, name string) (int64, error) {
	var id int64
	rows, err := s.db.Query(`select user_id from subscriptions where user_id = ? and name = ?`, userID, name)
	if err != nil {
		return id, err
	}
	defer rows.Close()

	for rows.Next() {
		err = rows.Scan(&id)
		if err != nil {
			return id, err
		}
	}

	return id, nil
}

func (s subscriptionDb) EndTime() error {
	rows, err := s.db.Query(`update subscriptions set name = 'standard', start = (current_date()), end = null where end < now()`)
	defer rows.Close()
	return err
}

func (s subscriptionDb) Update(userID int, name string, end string) error {
	time := utils.Timestamp()
	rows, err := s.db.Query(`update subscriptions set name = ?, end = ?, start = ? where user_id = ?`, name, end, time, userID)
	defer rows.Close()
	return err
}

func (s subscriptionDb) DailyDiamonds(name string) (int, error) {
	var diamonds int
	err := s.db.Get(&diamonds, `select diamonds from subscriptions_info where name = ?`, name)
	return diamonds, err
}

func (s subscriptionDb) GetSubscription(userID int) (string, error) {
	var name string
	err := s.db.Get(&name, `select name from subscriptions where user_id = ?`, userID)
	return name, err
}
