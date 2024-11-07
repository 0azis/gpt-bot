package db

import "github.com/jmoiron/sqlx"

type chatModel struct {
	ID     int    `json:"id" db:"id"`
	UserID int    `json:"userID" db:"user_id"`
	Title  string `json:"title" db:"title"`
}

type chatRepository interface {
	Create(userID int) error
	GetChats(userID int) ([]chatModel, error)
}

type chat struct {
	db *sqlx.DB
}

func (c chat) Create(userID int) error {
	_, err := c.db.Query(`insert into chats (userID) values (?)`, userID)
	return err
}

func (c chat) GetChats(userID int) ([]chatModel, error) {
	var chats []chatModel
	err := c.db.Select(&chats, `select * from chats where user_id = ?`, userID)
	return chats, err
}
