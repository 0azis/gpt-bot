package db

import "github.com/jmoiron/sqlx"

type userModel struct {
	ID           int     `json:"id" db:"id"`
	Subscription string  `json:"subscription" db:"subscription"`
	Requests     int     `json:"requestsCount" db:"requests"`
	Avatar       string  `json:"avatar" db:"avatar"`
	Balance      int     `json:"balance" db:"balance"`
	ReferralCode *string `json:"referralCode" db:"referral_code"`
	ReferralBy   *string `json:"referralBy" db:"referral_by"`
}

type userRepository interface {
	Create(userID int, avatarUrl string) error
	GetUser(jwtUserID int) (userModel, error)
	SetReferralCode(userID int, refCode string) error
	SetReferralBy(userID int, refBy string) error
	CheckReferredUser(userID int, refCode string) error
}

type user struct {
	db *sqlx.DB
}

// Create return nil (test)
func (u user) Create(userID int, avatarUrl string) error {
	_, err := u.db.Query(`insert into users (id, avatar) values (?, ?) on duplicate key update id = id, avatar = avatar`, userID, avatarUrl)
	return err
}

func (u user) GetUser(jwtUserID int) (userModel, error) {
	var user userModel
	err := u.db.Get(&user, `select * from users where id = ?`, jwtUserID)
	return user, err
}

func (u user) SetReferralCode(userID int, refCode string) error {
	_, err := u.db.Query(`update users set referral_code = ? where id = ?`, refCode, userID)
	return err
}

func (u user) SetReferralBy(userID int, refBy string) error {
	_, err := u.db.Query(`update users set referral_by = ? where id = ?`, refBy, userID)
	return err
}

func (u user) CheckReferredUser(userID int, refCode string) error {
	var id int
	err := u.db.Get(&id, `select id from users where referral_by = ? and id = ?`, refCode, userID)
	return err
}
