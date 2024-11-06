package db

import "github.com/jmoiron/sqlx"

type userRepository interface {
	Create(userID int) error
}

type user struct {
	db *sqlx.DB
}

// Create return nil (test)
func (u user) Create(userID int) error {
	_, err := u.db.Query(`insert into users (id) values (?) on duplicate key update id = id`, userID)
	return err
}
