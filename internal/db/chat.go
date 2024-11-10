package db

import (
	"slices"

	"github.com/jmoiron/sqlx"
)

type chatType string

var (
	ChatText  chatType = "text"
	ChatImage chatType = "image"
)

var modelNames = map[chatType][]string{
	ChatText:  []string{"o1-preview", "gpt-4o", "o1-mini", "gpt-4o-mini"},
	ChatImage: []string{"dall-e-3", "runware"},
}

type ChatModel struct {
	ID     int      `json:"id" db:"id"`
	UserID int      `json:"userID" db:"user_id"`
	Title  *string  `json:"title" db:"title"`
	Model  string   `json:"model" db:"model"`
	Type   chatType `json:"type" db:"type"`
}

func (cm ChatModel) Valid() bool {
	if slices.Contains(modelNames[cm.Type], cm.Model) {
		return true
	}
	return false
}

type chatRepository interface {
	Create(chat ChatModel) (int, error)
	GetChats(userID int) ([]ChatModel, error)
	GetChatInfo(chatID int) (ChatModel, error)
}

type chat struct {
	db *sqlx.DB
}

func (c chat) Create(chat ChatModel) (int, error) {
	var chatID int
	rows, err := c.db.Exec(`insert into chats (user_id, model, type) values (?, ?, ?)`, chat.UserID, chat.Model, chat.Type)
	if err != nil {
		return chatID, err
	}
	lastID, err := rows.LastInsertId()
	chatID = int(lastID)
	return chatID, err
}

func (c chat) GetChats(userID int) ([]ChatModel, error) {
	var chats []ChatModel
	err := c.db.Select(&chats, `select * from chats where user_id = ?`, userID)
	return chats, err
}

func (c chat) GetChatInfo(chatID int) (ChatModel, error) {
	var chat ChatModel
	err := c.db.Get(&chat, `select model, type from chats where id = ?`, chatID)
	return chat, err
}

// func (c chat) UpdateTitle(chatID int, title string) error {
// 	_, err := c.db.Query(`update chats set title = ? where id = ?`, title, chatID)
// 	return err
// }
