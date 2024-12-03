package db

import (
	"gpt-bot/utils"

	"github.com/jmoiron/sqlx"
)

type statDb struct {
	db *sqlx.DB
}

func (s statDb) Count(userID int64) error {
	time := utils.Timestamp()
	_, err := s.db.Exec(`insert into stats (user_id, clicked_at) values (?, ?)`, userID, time)
	return err
}

func (s statDb) Daily() (int, error) {
	var daily int
	err := s.db.Get(&daily, `select count(*) from stats where clicked_at >= curdate()`)
	return daily, err
}

func (s statDb) Monthly() (int, error) {
	var monthly int
	err := s.db.Get(&monthly, `select count(*) from stats where clicked_at >= date_sub(curdate(), interval dayofmonth(curdate())-1 day) `)
	return monthly, err
}

func (s statDb) All() (int, error) {
	var all int
	err := s.db.Get(&all, `select count(*) from stats`)
	return all, err
}
