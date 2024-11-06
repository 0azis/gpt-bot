package db

import "github.com/jmoiron/sqlx"

type userModel struct {
	ID           int  `json:"id" db:"id"`
	Subscription bool `json:"subscription" db:"subscription"`
	Requests     int  `json:"requestsCount" db:"requests"`
}

type userRepository interface {
	Create(userID int) error
	GetUser(jwtUserID int) (userModel, error)
}

type user struct {
	db *sqlx.DB
}

// Create return nil (test)
func (u user) Create(userID int) error {
	_, err := u.db.Query(`insert into users (id) values (?) on duplicate key update id = id`, userID)
	return err
}

func (u user) GetUser(jwtUserID int) (userModel, error) {
	var user userModel
	err := u.db.Get(&user, `select * from users where id = ?`, jwtUserID)
	return user, err
}
