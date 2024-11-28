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
	// admin
	CountUsers() (int, error)
	DailyUsers() (int, error)
}

type SubscriptionRepository interface {
	InitStandard(userID int) error
	UserSubscription(userID int64, name string) (int64, error)
	EndTime() error
	Update(userID int, name string, end string) error
	DailyDiamonds(name string) (int, error)
}

type LimitsRepository interface {
	Create(limits domain.Limits) error
	Update(newLimits domain.Limits) error
	Reduce(userID int, model string) error
	GetLimitsByModel(userID int, model string) (int, error)
}

type MessageRepository interface {
	Create(msg domain.Message) error
	GetByChat(userID, chatID int) ([]domain.Message, error)

	// admin
	RequestsDaily() (int, error)
	RequestsWeekly() (int, error)
	RequestsMontly() (int, error)
}

type ChatRepository interface {
	Create(chat domain.Chat) (int, error)
	GetByUser(userID int) ([]domain.Chat, error)
	GetByID(chatID int) (domain.Chat, error)
	UpdateTitle(chatID int, title string) error
}

type BonusRepository interface {
	Create(bonus domain.Bonus) error
	GetAll(userID int) ([]*domain.Bonus, error)
	GetOne(bonusID int) (domain.Bonus, error)
	// GetCompleted(userID int) (completedBonuses []*domain.Bonus, err error)
	// GetUncompleted(userID int) (uncompletedBonuses []*domain.Bonus, err error)
	Delete(channel_name string) error
	MakeAwarded(bonusID, userID int) error
	GetAward(bonusID, userID int) (int, error)
	InitBonuses(userID int) error

	// admin
	DailyBonuses() (int, error)
	AllBonuses() (int, error)
}
