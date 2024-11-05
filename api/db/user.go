package db

import "github.com/jmoiron/sqlx"

type userRepository interface {
	Create() error
}

type user struct {
	db *sqlx.DB
}

// Create return nil (test)
func (u user) Create() error {
	return nil
}
