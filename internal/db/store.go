package db

import (
	"gpt-bot/config"
	"gpt-bot/internal/db/repository"

	_ "github.com/go-sql-driver/mysql"
	"github.com/jmoiron/sqlx"
)

// Store is a main struct with all repositories
type Store struct {
	User         repository.UserRepository
	Chat         repository.ChatRepository
	Message      repository.MessageRepository
	Bonus        repository.BonusRepository
	Subscription repository.SubscriptionRepository
	Limits       repository.LimitsRepository
}

func New(cfg config.Database) (Store, error) {
	db, err := sqlx.Connect("mysql", cfg.Addr())

	store := Store{
		User:         userDb{db},
		Chat:         chatDb{db},
		Message:      messageDb{db},
		Bonus:        bonusDb{db},
		Subscription: subscriptionDb{db},
		Limits:       limitsDb{db},
	}

	return store, err
}
