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
	sqlResult, err := c.db.Exec(`insert into chats (user_id, model, type) values (?, ?, ?)`, chat.UserID, chat.Model, chat.Type)
	if err != nil {
		return chatID, err
	}
	lastID, err := sqlResult.LastInsertId()
	chatID = int(lastID)
	return chatID, err
}

func (c chatDb) GetByUser(userID int) ([]domain.Chat, error) {
	var chats []domain.Chat
	err := c.db.Select(&chats, `select chats.id, chats.title, chats.model, chats.type from chats join messages on messages.chat_id = chats.id where user_id = ? group by chats.id order by max(messages.created_at) desc`, userID)
	return chats, err
}

func (c chatDb) GetByID(userID, chatID int) (domain.Chat, error) {
	var chat domain.Chat
	err := c.db.Get(&chat, `select id, title, model, type from chats where id = ? and user_id = ?`, chatID, userID)
	return chat, err
}

func (c chatDb) GetByMessage(userID, messageID int) (domain.Chat, error) {
	var chat domain.Chat
	err := c.db.Get(&chat, `select id, title, model, type from chats join messages on messages.chat_id = chats.id where messages.id = ? and chats.user_id = ?`, messageID, userID)
	return chat, err
}

func (c chatDb) UpdateTitle(chatID int, title string) error {
	rows, err := c.db.Query(`update chats set title = ? where id = ?`, title, chatID)
	defer rows.Close()
	return err
}
