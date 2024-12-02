package db

import (
	"gpt-bot/internal/db/domain"
	"gpt-bot/utils"

	"github.com/jmoiron/sqlx"
)

type bonusDb struct {
	db *sqlx.DB
}

func (b bonusDb) Create(bonus domain.Bonus) error {
	var bonusID int64
	sqlResult, err := b.db.Exec(`insert into bonuses (name, channel_name, max_users) values(?)`, bonus.Name, bonus.Channel.Name,
		bonus.MaxUsers)
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

func (b bonusDb) UpdateChannel(id int, channel string) error {
	_, err := b.db.Exec(`update bonuses set channel_name = ? where id = ?`, channel, id)
	return err
}

func (b bonusDb) UpdateAward(id int, award int) error {
	_, err := b.db.Exec(`update bonuses set award = ? where id = ?`, award, id)
	return err
}

func (b bonusDb) UpdateStatus(id int, status bool) error {
	_, err := b.db.Exec(`update bonuses set status = ? where id = ?`, status, id)
	return err
}

func (b bonusDb) GetAll(userID int) ([]*domain.Bonus, error) {
	var bonuses []*domain.Bonus
	rows, err := b.db.Query(`select distinct bonuses.*, user_bonuses.awarded from bonuses join user_bonuses on user_bonuses.bonus_id = bonuses.id where user_bonuses.user_id = ?`, userID)
	if err != nil {
		return bonuses, err
	}
	defer rows.Close()

	for rows.Next() {
		var bonus domain.Bonus
		err = rows.Scan(&bonus.ID, &bonus.Channel.Name, &bonus.Award, &bonus.Awarded)
		if err != nil {
			return bonuses, err
		}
		bonuses = append(bonuses, &bonus)
	}

	return bonuses, nil
}

func (b bonusDb) GetOne(bonusID int) (domain.Bonus, error) {
	var bonus domain.Bonus
	rows, err := b.db.Query(`select bonuses.id, bonuses.name, bonuses.channel_name, bonuses.status, bonuses.max_users, bonuses.created_at from bonuses join user_bonuses on user_bonuses.bonus_id = bonuses.id where bonuses.id = ?`, bonusID)
	if err != nil {
		return bonus, err
	}
	defer rows.Close()

	for rows.Next() {
		err = rows.Scan(&bonus.ID, &bonus.Name, &bonus.Channel.Name, &bonus.Status, &bonus.MaxUsers, &bonus.CreatedAt)
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
	rows, err := b.db.Query(`select bonuses.id, bonuses.name, bonuses.channel_name, bonuses.status, bonuses.max_users, bonuses.created_at from bonuses join user_bonuses on user_bonuses.bonus_id = bonuses.id`)
	if err != nil {
		return bonuses, err
	}
	defer rows.Close()

	for rows.Next() {
		var bonus domain.Bonus
		err = rows.Scan(&bonus.ID, &bonus.Name, &bonus.Channel.Name, &bonus.Status, &bonus.MaxUsers, &bonus.CreatedAt)
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

func (b bonusDb) TurnBonus(status bool, bonusID int) error {
	_, err := b.db.Exec(`update bonuses set status = ? where id = ?`, status, bonusID)
	return err
}

func (b bonusDb) Delete(channel_name string) error {
	rows, err := b.db.Query(`delete from bonuses where channel_name = ?`, channel_name)
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
