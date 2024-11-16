package db

import (
	"gpt-bot/internal/db/domain"

	"github.com/jmoiron/sqlx"
)

type chatDb struct {
	db *sqlx.DB
}

func (c chatDb) Create(chat domain.Chat) (int, error) {
	var chatID int
	rows, err := c.db.Exec(`insert into chats (user_id, model, type) values (?, ?, ?)`, chat.UserID, chat.Model, chat.Type)
	if err != nil {
		return chatID, err
	}
	lastID, err := rows.LastInsertId()
	chatID = int(lastID)
	return chatID, err
}

func (c chatDb) GetByUser(userID int) ([]domain.Chat, error) {
	var chats []domain.Chat
	err := c.db.Select(&chats, `select id, title, model, type from chats where user_id = ?`, userID)
	return chats, err
}

func (c chatDb) GetByID(chatID int) (domain.Chat, error) {
	var chat domain.Chat
	err := c.db.Get(&chat, `select id, title, model, type from chats where id = ?`, chatID)
	return chat, err
}

func (c chatDb) UpdateTitle(chatID int, title string) error {
	_, err := c.db.Query(`update chats set title = ? where id = ?`, title, chatID)
	return err
}
