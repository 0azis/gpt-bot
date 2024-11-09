package db

import (
	"github.com/jmoiron/sqlx"
)

type chatType string

var (
	ChatText  chatType = "chat"
	ChatImage chatType = "image"
)

type ChatModel struct {
	ID     int      `json:"id" db:"id"`
	UserID int      `json:"userID" db:"user_id"`
	Title  *string  `json:"title" db:"title"`
	Model  string   `json:"model" db:"model"`
	Type   chatType `json:"type" db:"type"`
}

func (cm ChatModel) Valid() bool {
	if cm.Model != "" && (cm.Type == ChatText || cm.Type == ChatImage) {
		return true
	}
	return false
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
	_, err := c.db.Query(`insert into chats (user_id, model, type) values (?, ?, ?)`, chat.UserID, chat.Model, chat.Type)
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
