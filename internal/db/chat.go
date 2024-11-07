package db

import (
	"github.com/jmoiron/sqlx"
)

type ChatModel struct {
	ID     int     `json:"id" db:"id"`
	UserID int     `json:"userID" db:"user_id"`
	Title  *string `json:"title" db:"title"`
	Model  string  `json:"model" db:"model"`
}

type chatRepository interface {
	Create(chat ChatModel) error
	GetChats(userID int) ([]ChatModel, error)
	GetModelOfChat(chatID int) (string, error)
}

type chat struct {
	db *sqlx.DB
}

func (c chat) Create(chat ChatModel) error {
	_, err := c.db.Query(`insert into chats (user_id, model) values (?, ?)`, chat.UserID, chat.Model)
	return err
}

func (c chat) GetChats(userID int) ([]ChatModel, error) {
	var chats []ChatModel
	err := c.db.Select(&chats, `select * from chats where user_id = ?`, userID)
	return chats, err
}

func (c chat) GetModelOfChat(chatID int) (string, error) {
	var model string
	err := c.db.Get(&model, `select model from chats where id = ?`, chatID)
	return model, err
}

// func (c chat) UpdateTitle(chatID int, title string) error {
// 	_, err := c.db.Query(`update chats set title = ? where id = ?`, title, chatID)
// 	return err
// }
