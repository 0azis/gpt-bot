package db

import "github.com/jmoiron/sqlx"

type MessageModel struct {
	ID      int    `json:"id" db:"id"`
	ChatID  int    `json:"chatID" db:"chat_id"`
	Content string `json:"content" db:"content"`
	IsUser  bool   `json:"isUser" db:"is_user"`
	Model   string `json:"-"`
}

type messageRepository interface {
	Create(msg MessageModel) error
	GetMessages(userID, chatID int) ([]MessageModel, error)
}

type message struct {
	db *sqlx.DB
}

func (m message) Create(msg MessageModel) error {
	_, err := m.db.Query(`insert into messages (chat_id, content, is_user) values (?, ?, ?)`, msg.ChatID, msg.Content, msg.IsUser)
	return err
}

func (m message) GetMessages(userID, chatID int) ([]MessageModel, error) {
	var messages []MessageModel
	err := m.db.Select(&messages, `select messages.id, messages.chat_id, messages.content, messages.is_user from messages inner join chats on chats.id = messages.chat_id where messages.chat_id = ? and chats.user_id = ?`, chatID, userID)
	return messages, err
}
