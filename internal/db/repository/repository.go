package repository

import "gpt-bot/internal/db/domain"

type UserRepository interface {
	// user
	Create(user domain.User) error
	GetByID(jwtUserID int) (domain.User, error)
	GetAll() ([]domain.User, error)
	// referral
	IsUserReferred(userID int, refCode string) (int, error)
	SetReferredBy(userID int, refBy string) error
	OwnerOfReferralCode(refCode string) (int, error)
	//balance
	GetBalance(userID int) (int, error)
	RaiseBalance(userID, sum int) error
	ReduceBalance(userID, sum int) error
	FillBalance(userID, balance int) error
}

type SubscriptionRepository interface {
	InitStandard(userID int) error
	UserSubscription(userID int64, name string) (int64, error)
	EndTime() error
	Update(userID int, name string, end string) error
	DailyDiamonds(name string) (int, error)
}

type MessageRepository interface {
	Create(msg domain.Message) error
	GetByChat(userID, chatID int) ([]domain.Message, error)
}

type ChatRepository interface {
	Create(chat domain.Chat) (int, error)
	GetByUser(userID int) ([]domain.Chat, error)
	GetByID(chatID int) (domain.Chat, error)
	UpdateTitle(chatID int, title string) error
}

type BonusRepository interface {
	ChangeAward(bonusType string, award int) error
	GetAward(bonusType string) (int, error)
}
