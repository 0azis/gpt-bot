package db

import (
	"gpt-bot/internal/db/domain"
	"gpt-bot/utils"

	"github.com/jmoiron/sqlx"
)

type bonusDb struct {
	db *sqlx.DB
}

func (b bonusDb) Create() error {
	var bonusID int64
	sqlResult, err := b.db.Exec(`insert into bonuses values ()`)
	if err != nil {
		return err
	}
	bonusID, err = sqlResult.LastInsertId()
	if err != nil {
		return err
	}

	rows, err := b.db.Query(`insert into user_bonuses (bonus_id, user_id) select ?, id from users`, bonusID)
	defer rows.Close()
	return err
}

func (b bonusDb) UpdateName(id int, name string) error {
	_, err := b.db.Exec(`update bonuses set name = ? where id = ?`, name, id)
	return err
}

func (b bonusDb) UpdateChannel(id int, channelID int, link string) error {
	_, err := b.db.Exec(`update bonuses set channel_id = ?, link = ? where id = ?`, channelID, link, id)
	return err
}

func (b bonusDb) UpdateAward(id int, award int) error {
	_, err := b.db.Exec(`update bonuses set award = ? where id = ?`, award, id)
	return err
}

func (b bonusDb) UpdateStatus(id int, status bool) error {
	_, err := b.db.Exec(`update bonuses set is_check = ? where id = ?`, status, id)
	return err
}

func (b bonusDb) UpdateMaxUsers(id int, maxUsers int) error {
	_, err := b.db.Exec(`update bonuses set max_users = ? where id = ?`, maxUsers, id)
	return err
}

func (b bonusDb) GetAll(userID int) ([]*domain.Bonus, error) {
	var bonuses []*domain.Bonus
	rows, err := b.db.Query(`select distinct bonuses.ID, bonuses.link, bonuses.channel_id, bonuses.award, user_bonuses.awarded from bonuses join user_bonuses on user_bonuses.bonus_id = bonuses.id where user_bonuses.user_id = ? and is_check = 1`, userID)
	if err != nil {
		return bonuses, err
	}
	defer rows.Close()

	for rows.Next() {
		var bonus domain.Bonus
		err = rows.Scan(&bonus.ID, &bonus.Link, &bonus.Channel.ID, &bonus.Award, &bonus.Awarded)
		if err != nil {
			return bonuses, err
		}
		bonuses = append(bonuses, &bonus)
	}

	return bonuses, nil
}

func (b bonusDb) GetOne(bonusID int) (domain.Bonus, error) {
	var bonus domain.Bonus
	rows, err := b.db.Query(`select bonuses.id, bonuses.name, bonuses.award, bonuses.channel_id, bonuses.is_check, bonuses.max_users, bonuses.created_at from bonuses join user_bonuses on user_bonuses.bonus_id = bonuses.id where bonuses.id = ?`, bonusID)
	if err != nil {
		return bonus, err
	}
	defer rows.Close()

	for rows.Next() {
		err = rows.Scan(&bonus.ID, &bonus.Name, &bonus.Award, &bonus.Channel.ID, &bonus.Check, &bonus.MaxUsers, &bonus.CreatedAt)
		if err != nil {
			return bonus, err
		}
	}
	return bonus, err
}

// func (b bonusDb) GetCompleted(userID int) (completedBonuses []*domain.Bonus, err error) {
// 	rows, err := b.db.Query(`select bonuses.* from user_bonuses join bonuses on bonuses.id = user_bonuses.bonus_id where user_bonuses.completed = 1`)
// 	if err != nil {
// 		return completedBonuses, err
// 	}

// 	for rows.Next() {
// 		var bonus domain.Bonus
// 		err = rows.Scan(&bonus.ID, &bonus.Channel.Name, &bonus.Award)
// 		if err != nil {
// 			return completedBonuses, err
// 		}
// 		completedBonuses = append(completedBonuses, &bonus)
// 	}

// 	return completedBonuses, err

// }

// func (b bonusDb) GetUncompleted(userID int) (uncompletedBonuses []*domain.Bonus, err error) {
// 	rows, err := b.db.Query(`select bonuses.* from user_bonuses join bonuses on bonuses.id = user_bonuses.bonus_id where user_bonuses.completed = 0`)
// 	if err != nil {
// 		return uncompletedBonuses, err
// 	}

// 	for rows.Next() {
// 		var bonus domain.Bonus
// 		err = rows.Scan(&bonus.ID, &bonus.Channel.Name, &bonus.Award)
// 		if err != nil {
// 			return uncompletedBonuses, err
// 		}
// 		uncompletedBonuses = append(uncompletedBonuses, &bonus)
// 	}

// 	return uncompletedBonuses, err
// }

func (b bonusDb) DailyBonusesCount() (int, error) {
	var dailyBonuses int
	err := b.db.Get(&dailyBonuses, `select count(*) from user_bonuses where awarded = 1 and awarded_at >= curdate()`)
	return dailyBonuses, err
}

func (b bonusDb) MonthlyBonusesCount() (int, error) {
	var monthlyBonuses int
	err := b.db.Get(&monthlyBonuses, `select count(*) from user_bonuses where awarded = 1 and awarded_at >= date_sub(curdate(), interval dayofmonth(curdate())-1 day)`)
	return monthlyBonuses, err
}

func (b bonusDb) AllBonusesCount() (int, error) {
	var allBonuses int
	err := b.db.Get(&allBonuses, `select count(*) from user_bonuses where awarded = 1`)
	return allBonuses, err
}

func (b bonusDb) AllBonuses() ([]*domain.Bonus, error) {
	var bonuses []*domain.Bonus
	rows, err := b.db.Query(`select bonuses.id, bonuses.name, bonuses.channel_id, bonuses.is_check, bonuses.link, bonuses.max_users, bonuses.created_at from bonuses`)
	if err != nil {
		return bonuses, err
	}
	defer rows.Close()

	for rows.Next() {
		var bonus domain.Bonus
		err = rows.Scan(&bonus.ID, &bonus.Name, &bonus.Channel.ID, &bonus.Check, &bonus.Link, &bonus.MaxUsers, &bonus.CreatedAt)
		if err != nil {
			return bonuses, err
		}
		bonuses = append(bonuses, &bonus)
	}

	return bonuses, nil
}

func (b bonusDb) BonusesByID(bonusID int) (int, error) {
	var allBonuses int
	err := b.db.Get(&allBonuses, `select count(user_id) from user_bonuses where awarded = 1 and bonus_id = ?`, bonusID)
	return allBonuses, err
}

func (b bonusDb) BonusesByUser(userID int) (int, error) {
	var bonusUser int
	err := b.db.Get(&bonusUser, `select count(user_id) from user_bonuses where user_id = ? and awarded = 1`)
	return bonusUser, err
}

func (b bonusDb) Delete(id int) error {
	rows, err := b.db.Query(`delete from bonuses where id = ?`, id)
	defer rows.Close()
	return err
}

func (b bonusDb) MakeAwarded(bonusID, userID int) error {
	timestamp := utils.Timestamp()
	rows, err := b.db.Query(`update user_bonuses set awarded = 1, awarded_at = ? where bonus_id = ? and user_id = ?`, timestamp, bonusID, userID)
	defer rows.Close()
	return err
}

func (b bonusDb) GetAward(bonusID, userID int) (int, error) {
	var award int
	err := b.db.Get(&award, `select award from bonuses join user_bonuses on user_bonuses.bonus_id = bonuses.id where user_bonuses.user_id = ? and bonuses.id = ?`, userID, bonusID)
	return award, err
}

func (b bonusDb) InitBonuses(userID int) error {
	rows, err := b.db.Query(`insert into user_bonuses (bonus_id, user_id) select id, ? from bonuses on duplicate key update bonus_id=bonus_id`, userID)
	defer rows.Close()
	return err
}
