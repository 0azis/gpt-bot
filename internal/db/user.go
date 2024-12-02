package db

import (
	"gpt-bot/internal/db/domain"
	"gpt-bot/utils"

	"github.com/jmoiron/sqlx"
)

type userDb struct {
	db *sqlx.DB
}

func (u userDb) Create(user domain.User) error {
	refCode := utils.GenerateReferralCode(utils.UserRefCode)

	rows, err := u.db.Query(`insert into users (id, avatar, language_code, referral_code) values (?, ?, ?, ?) on duplicate key update avatar = avatar`, user.ID, user.Avatar, user.LanguageCode, refCode)
	defer rows.Close()
	return err
}

func (u userDb) GetByID(jwtUserID int) (domain.User, error) {
	var user domain.User
	rows, err := u.db.Query(`select users.id, subscriptions.name, subscriptions.start, subscriptions.end, limits.o1_preview, limits.gpt_4o, limits.o1_mini, limits.gpt_4o_mini, limits.runware, limits.dall_e_3, users.balance, users.avatar, users.referral_code, users.referred_by from users join subscriptions on subscriptions.user_id = users.id join limits on limits.user_id = users.id where users.id = ?`, jwtUserID)
	if err != nil {
		return user, err
	}
	defer rows.Close()

	for rows.Next() {
		err = rows.Scan(&user.ID, &user.Subscription.Name, &user.Subscription.Start, &user.Subscription.End, &user.Limits.O1Preview, &user.Limits.Gpt4o, &user.Limits.O1Mini, &user.Limits.Gpt4oMini, &user.Limits.Runware, &user.Limits.Dalle3, &user.Balance, &user.Avatar, &user.ReferralCode, &user.ReferredBy)
		if err != nil {
			return user, err
		}
	}

	return user, err
}

func (u userDb) GetAll() ([]domain.User, error) {
	var users []domain.User
	rows, err := u.db.Query(`select users.id, subscriptions.name, users.balance, users.avatar, users.referral_code, users.referred_by from users left join subscriptions on subscriptions.user_id = users.id`)
	if err != nil {
		return users, err
	}
	defer rows.Close()

	for rows.Next() {
		var user domain.User
		err = rows.Scan(&user.ID, &user.Subscription.Name, &user.Balance, &user.Avatar, &user.ReferralCode, &user.ReferredBy)
		if err != nil {
			return users, err
		}
		users = append(users, user)
	}

	return users, nil
}

func (u userDb) IsUserReferred(userID int, refCode string) (int, error) {
	var id int
	rows, err := u.db.Query(`select id from users where (referred_by = ? or referred_by != "") and id = ?`, refCode, userID)
	if err != nil {
		return id, err
	}
	defer rows.Close()

	for rows.Next() {
		err = rows.Scan(&id)
		if err != nil {
			return id, err
		}
	}
	return id, nil
}

func (u userDb) SetReferredBy(userID int, refBy string) error {
	rows, err := u.db.Query(`update users set referred_by = ? where id = ?`, refBy, userID)
	defer rows.Close()
	return err
}

func (u userDb) OwnerOfReferralCode(refCode string) (int, error) {
	var id int
	err := u.db.Get(&id, `select id from users where referral_code = ?`, refCode)
	return id, err
}

func (u userDb) GetBalance(userID int) (int, error) {
	var balance int
	rows, err := u.db.Query(`select balance from users where id = ?`, userID)
	if err != nil {
		return balance, err
	}
	defer rows.Close()

	for rows.Next() {
		err = rows.Scan(&balance)
		if err != nil {
			return balance, err
		}
	}

	return balance, nil
}

func (u userDb) RaiseBalance(userID, sum int) error {
	rows, err := u.db.Query(`update users set balance = balance + ? where id = ?`, sum, userID)
	defer rows.Close()
	return err
}

func (u userDb) ReduceBalance(userID, sum int) error {
	rows, err := u.db.Query(`update users set balance = balance - ? where id = ?`, sum, userID)
	defer rows.Close()
	return err
}

func (u userDb) FillBalance(userID, balance int) error {
	rows, err := u.db.Query(`update users set balance = ? where id = ?`, balance, userID)
	defer rows.Close()
	return err
}

func (u userDb) AllUsersCount() (int, error) {
	var count int
	err := u.db.Get(&count, `select count(id) from users`)
	return count, err
}

func (u userDb) DailyUsersCount() (int, error) {
	var dailyCount int
	err := u.db.Get(&dailyCount, `select count(id) from users where created_at >= curdate()`)
	return dailyCount, err
}

func (u userDb) WeeklyUsersCount() (int, error) {
	var monthlyUsers int
	err := u.db.Get(&monthlyUsers, `select count(id) from users where date(created_at) >= date_sub(curdate(), interval dayofweek(curdate())-1 day)`)
	return monthlyUsers, err
}

func (u userDb) MonthlyUsersCount() (int, error) {
	var monthlyUsers int
	err := u.db.Get(&monthlyUsers, `select count(id) from users where date(created_at) >= date_sub(curdate(), interval dayofmonth(curdate())-1 day)`)
	return monthlyUsers, err
}

func (u userDb) AllUsers() ([]domain.User, error) {
	var users []domain.User
	err := u.db.Get(&users, `select * from users`)
	return users, err
}

func (u userDb) DailyUsers() ([]domain.User, error) {
	var users []domain.User
	err := u.db.Get(&users, `select * from users where date(created_at) >= curdate()`)
	return users, err
}

func (u userDb) WeeklyUsers() ([]domain.User, error) {
	var users []domain.User
	err := u.db.Get(&users, `select * from users where date(created_at) >= date_sub(curdate(), interval dayofweek(curdate())-1 day)`)
	return users, err
}

func (u userDb) MonthlyUsers() ([]domain.User, error) {
	var users []domain.User
	err := u.db.Get(&users, `select * from users where date(created_at) >= date_sub(curdate(), interval dayofmonth(curdate())-1 day)`)
	return users, err
}

func (u userDb) AllUsersReferred() (int, error) {
	var allUsersReferred int
	err := u.db.Get(&allUsersReferred, `select count(*) from users where referred_by != ""`)
	return allUsersReferred, err
}

func (u userDb) DailyUsersReferred() (int, error) {
	var dailyUsersReferred int
	err := u.db.Get(&dailyUsersReferred, `select count(*) from users where referred_by != "" and date(created_at) >= curdate()`)
	return dailyUsersReferred, err
}

func (u userDb) WeeklyUsersReferred() (int, error) {
	var dailyUsersReferred int
	err := u.db.Get(&dailyUsersReferred, `select count(*) from users where referred_by != "" and date(created_at) >= date_sub(curdate(), interval dayofweek(curdate())-1 day)`)
	return dailyUsersReferred, err
}

func (u userDb) MonthlyUsersReferred() (int, error) {
	var dailyUsersReferred int
	err := u.db.Get(&dailyUsersReferred, `select count(*) from users where referred_by != "" and date(created_at) >= date_sub(curdate(), interval dayofmonth(curdate())-1 day)`)
	return dailyUsersReferred, err
}

func (u userDb) PremiumUsers() ([]domain.User, error) {
	var premiumUsers []domain.User
	err := u.db.Get(&premiumUsers, `select * from users join subscriptions on subscriptions.user_id = users.id where subscriptions.name = "advanced" or subscriptions.name = "ultimate"`)
	return premiumUsers, err
}

func (u userDb) PremiumUsersCount() (int, error) {
	var premiumUsers int
	err := u.db.Get(&premiumUsers, `select count(*) from users join subscriptions on subscriptions.user_id = users.id where subscriptions.name = "advanced" or subscriptions.name = "ultimate"`)
	return premiumUsers, err
}

func (u userDb) ActiveUsersAll() (int, error) {
	var activeUsers int
	err := u.db.Get(&activeUsers, `select count(distinct users.id) from users join chats on chats.user_id = users.id`)
	return activeUsers, err
}

func (u userDb) ActiveUsersDaily() (int, error) {
	var activeUsers int
	err := u.db.Get(&activeUsers, `select count(distinct users.id) from users join chats on chats.user_id = users.id where date(created_at) >= curdate()`)
	return activeUsers, err
}

func (u userDb) ActiveUsersWeekly() (int, error) {
	var activeUsers int
	err := u.db.Get(&activeUsers, `select count(distinct users.id) from users join chats on chats.user_id = users.id where date(created_at) >= date_sub(curdate(), interval dayofweek(curdate())-1 day)`)
	return activeUsers, err
}

func (u userDb) ActiveUsersMonthly() (int, error) {
	var activeUsers int
	err := u.db.Get(&activeUsers, `select count(distinct users.id) from users join chats on chats.user_id = users.id where date(created_at) >= date_sub(curdate(), interval dayofmonth(curdate())-1 day)`)
	return activeUsers, err
}
