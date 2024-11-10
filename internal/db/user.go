package db

import (
	"gpt-bot/utils"

	"github.com/jmoiron/sqlx"
)

type UserModel struct {
	ID           int     `json:"id" db:"id"`
	Subscription string  `json:"subscription" db:"subscription"`
	Requests     int     `json:"requestsCount" db:"requests"`
	Avatar       string  `json:"avatar" db:"avatar"`
	Balance      int     `json:"balance" db:"balance"`
	ReferralCode *string `json:"referralCode" db:"referral_code"`
	ReferredBy   *string `json:"referralBy" db:"referred_by"`
}

type userRepository interface {
	Create(user UserModel) error
	GetUser(jwtUserID int) (UserModel, error)
	IsUserReferred(userID int, refCode string) (int, error)
	SetReferredBy(userID int, refBy string) error
	OwnerReferralCode(refCode string) (int, error)
	RaiseBalance(userID, award int) error
	ReduceBalance(userID, sum int) error
}

type user struct {
	db *sqlx.DB
}

// Create return nil (test)
func (u user) Create(user UserModel) error {
	refCode := utils.ReferralCode()

	_, err := u.db.Query(`insert into users (id, avatar, referral_code) values (?, ?, ?) on duplicate key update avatar = avatar`, user.ID, user.Avatar, refCode)
	return err
}

func (u user) GetUser(jwtUserID int) (UserModel, error) {
	var user UserModel
	err := u.db.Get(&user, `select * from users where id = ?`, jwtUserID)
	return user, err
}

func (u user) IsUserReferred(userID int, refCode string) (int, error) {
	var id int
	rows, err := u.db.Query(`select id from users where referred_by = ? and id = ?`, refCode, userID)
	if err != nil {
		return id, err
	}
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

func (u user) OwnerReferralCode(refCode string) (int, error) {
	var id int
	err := u.db.Get(&id, `select id from users where referral_code = ?`, refCode)
	return id, err
}

func (u user) RaiseBalance(userID, award int) error {
	_, err := u.db.Query(`update users set balance = balance + ? where id = ?`, award, userID)
	return err
}

func (u user) ReduceBalance(userID, sum int) error {
	_, err := u.db.Query(`update users set balance = balance - ? where id = ?`, sum, userID)
	return err
}
