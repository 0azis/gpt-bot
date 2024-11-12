package db

import (
	"gpt-bot/utils"

	"github.com/jmoiron/sqlx"
)

type UserModel struct {
	ID           int     `json:"id" db:"id"`
	Subscription string  `json:"subscription"`
	Avatar       string  `json:"avatar"`
	Balance      int     `json:"balance"`
	ReferralCode *string `json:"referralCode"`
	ReferredBy   *string `json:"referredBy"`
}

type userRepository interface {
	// user
	Create(user UserModel) error
	GetByID(jwtUserID int) (UserModel, error)
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

type user struct {
	db *sqlx.DB
}

func (u user) Create(user UserModel) error {
	refCode := utils.GenerateReferralCode()

	_, err := u.db.Query(`insert into users (id, avatar, referral_code) values (?, ?, ?) on duplicate key update avatar = avatar`, user.ID, user.Avatar, refCode)
	return err
}

func (u user) GetByID(jwtUserID int) (UserModel, error) {
	var user UserModel
	rows, err := u.db.Query(`select users.id, subscriptions.name, users.balance, users.avatar, users.balance, users.referral_code, users.referred_by from users left join subscriptions on subscriptions.user_id = users.id where id = ?`, jwtUserID)
	if err != nil {
		return user, err
	}
	defer rows.Close()

	for rows.Next() {
		err = rows.Scan(&user.ID, &user.Subscription, &user.Balance, &user.Avatar, &user.Balance, &user.ReferralCode, &user.ReferredBy)
		if err != nil {
			return user, err
		}
	}

	return user, err
}

func (u user) IsUserReferred(userID int, refCode string) (int, error) {
	var id int
	rows, err := u.db.Query(`select id from users where referred_by = ? and id = ?`, refCode, userID)
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

func (u user) SetReferredBy(userID int, refBy string) error {
	_, err := u.db.Query(`update users set referred_by = ? where id = ?`, refBy, userID)
	return err
}

func (u user) OwnerOfReferralCode(refCode string) (int, error) {
	var id int
	err := u.db.Get(&id, `select id from users where referral_code = ?`, refCode)
	return id, err
}

func (u user) GetBalance(userID int) (int, error) {
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

func (u user) RaiseBalance(userID, sum int) error {
	_, err := u.db.Query(`update users set balance = balance + ? where id = ?`, sum, userID)
	return err
}

func (u user) ReduceBalance(userID, sum int) error {
	_, err := u.db.Query(`update users set balance = balance - ? where id = ?`, sum, userID)
	return err
}

func (u user) FillBalance(userID, balance int) error {
	_, err := u.db.Query(`update users set balance = ? where id = ?`, balance, userID)
	return err
}
