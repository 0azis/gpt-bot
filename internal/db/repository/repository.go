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
	AllUsersCount() (int, error)
	DailyUsersCount() (int, error)
	WeeklyUsersCount() (int, error)
	MonthlyUsersCount() (int, error)

	AllUsers() ([]domain.User, error)
	DailyUsers() ([]domain.User, error)
	WeeklyUsers() ([]domain.User, error)
	MonthlyUsers() ([]domain.User, error)

	AllUsersReferred() (int, error)
	DailyUsersReferred() (int, error)
	WeeklyUsersReferred() (int, error)
	MonthlyUsersReferred() (int, error)

	PremiumUsers() ([]domain.User, error)
	PremiumUsersCount() (int, error)

	ActiveUsersAll() (int, error)
	ActiveUsersDaily() (int, error)
	ActiveUsersWeekly() (int, error)
	ActiveUsersMonthly() (int, error)
}

type SubscriptionRepository interface {
	InitStandard(userID int) error
	UserSubscription(userID int64, name string) (int64, error)
	EndTime() error
	Update(userID int, name string, end string) error
	DailyDiamonds(name string) (int, error)
	GetSubscription(userID int) (string, error)
}

type LimitsRepository interface {
	Create(limits domain.Limits) error
	Update(newLimits domain.Limits) error
	Reduce(userID int, model string) error
	GetLimitsByModel(userID int, model string) (int, error)
	GetByUser(userID int) (domain.Limits, error)
}

type MessageRepository interface {
	Create(msg domain.Message) error
	GetByChat(userID, chatID int) ([]domain.Message, error)

	// admin
	RequestsDaily() (domain.LimitsMap, error)
	RequestsWeekly() (domain.LimitsMap, error)
	RequestsMontly() (domain.LimitsMap, error)
	// UsersDaily() (int, error)
	// UsersWeekly() (int, error)
	// UsersMonthly() (int, error)
	UsersDailyTwice() (int, error)
	UsersWeeklyTwice() (int, error)
	UsersMonthlyTwice() (int, error)
	MessagesDaily() (int, error)
	MessagesWeekly() (int, error)
	MessagesMonthly() (int, error)
	MessagesAll() (int, error)
	RequestsByUser(userID int) (domain.LimitsMap, error)
	LastMessageUser(userID int) (string, error)
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
	DailyBonusesCount() (int, error)
	MonthlyBonusesCount() (int, error)
	AllBonusesCount() (int, error)
	AllBonuses() ([]*domain.Bonus, error)
	BonusesByID(bonusID int) (int, error)
	BonusesByUser(userID int) (int, error)
}

type ReferralRepository interface {
	Create() error
	GetOne(code string) (int, error)
	GetAll() ([]domain.Referral, error)
	Delete(id int) error
	AddUser(userID, refId int) error
	CountUsers(id int) (int, error)
	AllUsers() (int, error)
	MonthlyUsers() (int, error)
	DailyUsers() (int, error)
}

type StatRepository interface {
	Count(userID int64) error
	Daily() (int, error)
	Monthly() (int, error)
	All() (int, error)
}

type AdminRepository interface {
	MakeAdmin(userID int) error
	CheckID(userID int) bool
}
