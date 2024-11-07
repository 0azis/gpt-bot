package db

import "github.com/jmoiron/sqlx"

type messageModel struct {
	ID      int    `json:"id" db:"id"`
	ChatID  int    `json:"chatID" db:"chat_id"`
	Content string `json:"content" db:"content"`
	IsUser  bool   `json:"isUser" db:"is_user"`
}

type messageRepository interface {
	Create(message messageModel) error
	GetMessages(chatID int) ([]messageModel, error)
}

type message struct {
	db *sqlx.DB
}

func (m message) Create(message messageModel) error {
	_, err := m.db.Query(`insert into messages (chat_id, content, is_user) values (?, ?, ?)`, message.ChatID, message.Content, message.IsUser)
	return err
}

func (m message) GetMessages(chatID int) ([]messageModel, error) {
	var messages []messageModel
	err := m.db.Select(&messages, `select * from messages where chat_id = ?`, chatID)
	return messages, err
}
