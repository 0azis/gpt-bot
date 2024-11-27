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
	_, err := b.db.Query(`insert into bonuses (channel_name, award) values (?, ?)`, bonus.Channel.Name, bonus.Award)
	if err != nil {
		return err
	}
	_, err = b.db.Query(`insert into user_bonuses (bonus_id, user_id) select last_insert_id(), id from users`)
	return err
}

func (b bonusDb) GetCompleted(userID int) (completedBonuses []domain.Bonus, err error) {
	err = b.db.Select(&completedBonuses, `select bonuses.* from user_bonuses join bonuses on bonuses.id = user_bonuses.bonus_id where user_bonuses.completed = 1`)
	return completedBonuses, err
}

func (b bonusDb) GetUncompleted(userID int) (uncompletedBonuses []domain.Bonus, err error) {
	err = b.db.Select(&uncompletedBonuses, `select bonuses.* from user_bonuses join bonuses on bonuses.id = user_bonuses.bonus_id where user_bonuses.completed = 0`)
	return uncompletedBonuses, err
}

func (b bonusDb) DailyBonuses() (dailyBonuses []domain.Bonus, err error) {
	err = b.db.Select(&dailyBonuses, `select count(*) from user_bonuses where completed = 1 and completed_at >= curdate()`)
	return dailyBonuses, err
}

func (b bonusDb) AllBonuses() (allBonuses []domain.Bonus, err error) {
	err = b.db.Select(&allBonuses, `select count(*) from user_bonuses where completed = 1`)
	return allBonuses, err
}

func (b bonusDb) Delete(bonusID int) error {
	_, err := b.db.Query(`delete from bonuses where id = ?`, bonusID)
	return err
}

func (b bonusDb) MakeCompleted(bonusID, userID int) error {
	timestamp := utils.Timestamp()
	_, err := b.db.Query(`update user_bonuses set completed = 1 and completed_at = ? where bonus_id = ? and user_id = ?`, timestamp, bonusID, userID)
	return err
}

func (b bonusDb) GetAward(bonusID, userID int) (int, error) {
	var award int
	err := b.db.Get(&award, `select award from bonuses join user_bonuses on user_bonuses.bonus_id = bonuses.id where user_bonuses.user_id = ? and where bonuses.id = ?`, userID, bonusID)
	return award, err
}
