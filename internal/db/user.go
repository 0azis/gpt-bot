package db

import "github.com/jmoiron/sqlx"

type userModel struct {
	ID           int    `json:"id" db:"id"`
	Subscription bool   `json:"subscription" db:"subscription"`
	Requests     int    `json:"requestsCount" db:"requests"`
	Avatar       string `json:"avatar" db:"avatar"`
}

type userRepository interface {
	Create(userID int, avatarUrl string) error
	GetUser(jwtUserID int) (userModel, error)
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
