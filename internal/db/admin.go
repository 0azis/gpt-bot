package db

import (
	"github.com/jmoiron/sqlx"
)

type adminDb struct {
	db *sqlx.DB
}

func (a adminDb) MakeAdmin(userID int) error {
	_, err := a.db.Exec(`insert into admins values (?)`, userID)
	return err
}

func (a adminDb) CheckID(userID int) bool {
	var id int
	err := a.db.Get(&id, `select * from admins where user_id = ?`, userID)
	return err == nil
}
