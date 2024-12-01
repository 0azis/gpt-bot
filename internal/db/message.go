package db

import (
	"gpt-bot/internal/db/domain"

	"github.com/jmoiron/sqlx"
)

type messageDb struct {
	db *sqlx.DB
}

func (m messageDb) Create(msg domain.Message) error {
	rows, err := m.db.Query(`insert into messages (chat_id, content, role, type) values (?, ?, ?, ?)`, msg.ChatID, msg.Content, msg.Role, msg.Type)
	defer rows.Close()
	return err
}

func (m messageDb) GetByChat(userID, chatID int) ([]domain.Message, error) {
	var messages []domain.Message
	err := m.db.Select(&messages, `select messages.content, messages.role, messages.type from messages inner join chats on chats.id = messages.chat_id where messages.chat_id = ? and chats.user_id = ?`, chatID, userID)
	return messages, err
}

func (m messageDb) RequestsDaily() (int, error) {
	var dailyCount int
	err := m.db.Get(&dailyCount, `select count(*) from messages where date(created_at) >= curdate() and role = "assistant"`)
	return dailyCount, err
}

func (m messageDb) RequestsWeekly() (int, error) {
	var weeklyCount int
	err := m.db.Get(&weeklyCount, `select count(*) from messages where date(created_at) >= date_sub(curdate(), interval dayofweek(curdate())-1 day) and role = "assistant"`)
	return weeklyCount, err
}

func (m messageDb) RequestsMontly() (int, error) {
	var montlyCount int
	err := m.db.Get(&montlyCount, `select count(*) from messages where date(created_at) >= date_sub(curdate(), interval dayofmonth(curdate())-1 day) and role = "assistant"`)
	return montlyCount, err
}
