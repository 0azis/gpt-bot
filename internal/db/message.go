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

func (m messageDb) Delete(messageID int) error {
	_, err := m.db.Exec(`delete from messages where id = ?`, messageID)
	return err
}

func (m messageDb) RequestsDaily() (domain.LimitsMap, error) {
	modelsCount := domain.LimitsMap{}
	rows, err := m.db.Query(`select distinct chats.id, chats.model from messages join chats on chats.id = messages.chat_id where date(created_at) >= curdate()`)
	if err != nil {
		return modelsCount, err
	}
	defer rows.Close()

	for rows.Next() {
		var id int
		var model string
		err = rows.Scan(&id, &model)
		modelsCount[model] += 1
	}

	return modelsCount, err
}

func (m messageDb) RequestsAll() (domain.LimitsMap, error) {
	modelsCount := domain.LimitsMap{}
	rows, err := m.db.Query(`select distinct chats.id, chats.model from messages join chats on chats.id = messages.chat_id`)
	if err != nil {
		return modelsCount, err
	}
	defer rows.Close()

	for rows.Next() {
		var id int
		var model string
		err = rows.Scan(&id, &model)
		modelsCount[model] += 1
	}

	return modelsCount, err
}

func (m messageDb) RequestsWeekly() (domain.LimitsMap, error) {
	modelsCount := domain.LimitsMap{}
	rows, err := m.db.Query(`select distinct chats.id, chats.model from messages join chats on chats.id = messages.chat_id where date(created_at) >= date_sub(curdate(), interval dayofweek(curdate())-1 day)`)
	if err != nil {
		return modelsCount, err
	}
	defer rows.Close()

	for rows.Next() {
		var id int
		var model string
		err = rows.Scan(&id, &model)
		modelsCount[model] += 1
	}
	return modelsCount, err
}

func (m messageDb) RequestsMontly() (domain.LimitsMap, error) {
	modelsCount := domain.LimitsMap{}
	rows, err := m.db.Query(`select distinct chats.id, chats.model from messages join chats on chats.id = messages.chat_id where date(created_at) >= date_sub(curdate(), interval dayofmonth(curdate())-1 day)`)
	if err != nil {
		return modelsCount, err
	}
	defer rows.Close()

	for rows.Next() {
		var id int
		var model string
		err = rows.Scan(&id, &model)
		modelsCount[model] += 1
	}
	return modelsCount, err
}

// func (m messageDb) UsersDaily() (int, error) {
// 	var usersCount int
// 	err := m.db.Get(&usersCount, `select count(distinct chats.user_id) from chats join messages on messages.chat_id = chats.id where role = "user" and date(messages.created_at) >= curdate()`)

// 	return usersCount, err
// }

// func (m messageDb) UsersWeekly() (int, error) {
// 	var usersCount int
// 	err := m.db.Get(&usersCount, `select count(distinct chats.user_id) from chats join messages on messages.chat_id = chats.id where role = "user" and date(messages.created_at) >= date_sub(curdate(), interval dayofweek(curdate())-1 day)`)

// 	return usersCount, err
// }

// func (m messageDb) UsersMonthly() (int, error) {
// 	var usersCount int
// 	err := m.db.Get(&usersCount, `select count(distinct chats.user_id) from chats join messages on messages.chat_id = chats.id where role = "user" and date(messages.created_at) >= date_sub(curdate(), interval dayofmonth(curdate())-1 day)`)

// 	return usersCount, err
// }

func (m messageDb) UsersDailyTwice() (int, error) {
	var usersCount int
	err := m.db.Get(&usersCount, `select count(user_id) from (select c.user_id from chats c join messages m on c.id = m.chat_id where date(m.created_at) >= curdate() group by c.user_id having count (distinct c.id) > 1)as qq`)
	return usersCount, err
}

func (m messageDb) UsersWeeklyTwice() (int, error) {
	var usersCount int
	err := m.db.Get(&usersCount, `select count(user_id) from (select c.user_id from chats c join messages m on c.id = m.chat_id where date(m.created_at) >= date_sub(curdate(), interval dayofweek(curdate())-1 day) group by c.user_id having count (distinct c.id) > 1)as qq`)
	return usersCount, err
}

func (m messageDb) UsersMonthlyTwice() (int, error) {
	var usersCount int
	err := m.db.Get(&usersCount, `select count(user_id) from (select c.user_id from chats c join messages m on c.id = m.chat_id where date(m.created_at) >= date_sub(curdate(), interval dayofmonth(curdate())-1 day) group by c.user_id having count (distinct c.id) > 1)as qq`)
	return usersCount, err
}

func (m messageDb) MessagesDaily() (int, error) {
	var messagesCount int
	err := m.db.Get(&messagesCount, `select count(*) from messages where date(created_at) >= curdate() and role = "user"`)
	return messagesCount, err
}

func (m messageDb) MessagesWeekly() (int, error) {
	var messagesCount int
	err := m.db.Get(&messagesCount, `select count(*) from messages where date(created_at) >= date_sub(curdate(), interval dayofweek(curdate())-1 day) and role = "user"`)
	return messagesCount, err
}

func (m messageDb) MessagesMonthly() (int, error) {
	var messagesCount int
	err := m.db.Get(&messagesCount, `select count(*) from messages where date(created_at) >= date_sub(curdate(), interval dayofmonth(curdate())-1 day)and role = "user"`)
	return messagesCount, err
}

func (m messageDb) MessagesAll() (int, error) {
	var messagesCount int
	err := m.db.Get(&messagesCount, `select count(*) from messages where role = "user"`)
	return messagesCount, err
}

func (m messageDb) RequestsByUser(userID int) (domain.LimitsMap, error) {
	modelsCount := domain.LimitsMap{}
	rows, err := m.db.Query(`select chats.model from messages join chats on chats.id = messages.chat_id where role = "user" and chats.user_id = ?`, userID)
	if err != nil {
		return modelsCount, err
	}
	defer rows.Close()

	for rows.Next() {
		var model string
		err = rows.Scan(&model)
		modelsCount[model] += 1
	}

	return modelsCount, err
}

func (m messageDb) LastMessageUser(userID int) (string, error) {
	var time []string
	err := m.db.Select(&time, `select created_at from messages join chats on chats.id = messages.chat_id where chats.user_id = ? order by created_at desc`, userID)
	if err != nil {
		return "", err
	}

	if len(time) == 0 {
		return "", nil
	}

	return time[0], nil
}
