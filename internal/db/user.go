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
	CheckReferredUser(userID int, refCode string) error
	CheckReferralCode(refCode string) error
}

type user struct {
	db *sqlx.DB
}

// Create return nil (test)
func (u user) Create(user UserModel) error {
	var refBy *string
	refBy = user.ReferredBy
	if refBy != nil {
		refBy = *&user.ReferredBy
	}
	refCode := utils.ReferralCode()

	_, err := u.db.Query(`insert into users (id, avatar, referral_code, referred_by) values (?, ?, ?, ?) on duplicate key update avatar = avatar`, user.ID, user.Avatar, refCode, refBy)
	return err
}

func (u user) GetUser(jwtUserID int) (UserModel, error) {
	var user UserModel
	err := u.db.Get(&user, `select * from users where id = ?`, jwtUserID)
	return user, err
}

func (u user) CheckReferredUser(userID int, refCode string) error {
	var id int
	err := u.db.Get(&id, `select id from users where referral_by = ? and id = ?`, refCode, userID)
	return err
}

func (u user) CheckReferralCode(refCode string) error {
	var id int
	err := u.db.Get(&id, `select id from users where referral_code = ?`, refCode)
	return err
}
