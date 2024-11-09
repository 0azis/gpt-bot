package db

import (
	_ "github.com/go-sql-driver/mysql"
	"github.com/jmoiron/sqlx"
)

// Store is a main struct with all repositories
type Store struct {
	User    userRepository
	Chat    chatRepository
	Message messageRepository
	Bonus   bonusRepository
}

func New(uri string) (Store, error) {
	db, err := sqlx.Connect("mysql", uri)

	store := Store{
		User:    user{db},
		Chat:    chat{db},
		Message: message{db},
		Bonus:   bonus{db},
	}

	return store, err
}
