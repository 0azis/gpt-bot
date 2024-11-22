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
	refCode := utils.GenerateReferralCode()

	_, err := u.db.Query(`insert into users (id, avatar, referral_code) values (?, ?, ?) on duplicate key update avatar = avatar`, user.ID, user.Avatar, refCode)
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
	_, err := u.db.Query(`update users set referred_by = ? where id = ?`, refBy, userID)
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

	for rows.Next() {
		err = rows.Scan(&balance)
		if err != nil {
			return balance, err
		}
	}

	return balance, nil
}

func (u userDb) RaiseBalance(userID, sum int) error {
	_, err := u.db.Query(`update users set balance = balance + ? where id = ?`, sum, userID)
	return err
}

func (u userDb) ReduceBalance(userID, sum int) error {
	_, err := u.db.Query(`update users set balance = balance - ? where id = ?`, sum, userID)
	return err
}

func (u userDb) FillBalance(userID, balance int) error {
	_, err := u.db.Query(`update users set balance = ? where id = ?`, balance, userID)
	return err
}
