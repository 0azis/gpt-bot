package db

import (
	"gpt-bot/config"

	_ "github.com/go-sql-driver/mysql"
	"github.com/jmoiron/sqlx"
)

// Store is a main struct with all repositories
type Store struct {
	User         userRepository
	Chat         chatRepository
	Message      messageRepository
	Bonus        bonusRepository
	Subscription subscriptionRepository
}

func New(cfg config.Database) (Store, error) {
	db, err := sqlx.Connect("mysql", cfg.Addr())

	store := Store{
		User:         user{db},
		Chat:         chat{db},
		Message:      message{db},
		Bonus:        bonus{db},
		Subscription: subscription{db},
	}

	return store, err
}
