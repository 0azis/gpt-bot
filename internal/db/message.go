package db

import (
	"github.com/jmoiron/sqlx"
	"github.com/sashabaranov/go-openai"
)

type MessageModel struct {
	ID      int    `json:"id" db:"id"`
	ChatID  int    `json:"chatID" db:"chat_id"`
	Content string `json:"content" db:"content"`
	Role    string `json:"role" db:"role"`
}

func NewUserMessage(chatID int, content string) MessageModel {
	return MessageModel{
		ChatID:  chatID,
		Content: content,
		Role:    openai.ChatMessageRoleUser,
	}
}

func NewAssistantMessage(chatID int, content string) MessageModel {
	return MessageModel{
		ChatID:  chatID,
		Content: content,
		Role:    openai.ChatMessageRoleAssistant,
	}
}

type MessageCredentials struct {
	ChatID  int    `json:"chatId"`
	Content string `json:"content"`
}

func (mc MessageCredentials) Valid() bool {
	return mc.Content != ""
}

type messageRepository interface {
	Create(msg MessageModel) error
	GetMessages(userID, chatID int) ([]MessageModel, error)
}

type message struct {
	db *sqlx.DB
}

func (m message) Create(msg MessageModel) error {
	_, err := m.db.Query(`insert into messages (chat_id, content, role) values (?, ?, ?)`, msg.ChatID, msg.Content, msg.Role)
	return err
}

func (m message) GetMessages(userID, chatID int) ([]MessageModel, error) {
	var messages []MessageModel
	err := m.db.Select(&messages, `select messages.id, messages.chat_id, messages.content, messages.role from messages inner join chats on chats.id = messages.chat_id where messages.chat_id = ? and chats.user_id = ?`, chatID, userID)
	return messages, err
}
