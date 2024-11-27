package db

import (
	"gpt-bot/internal/db/domain"

	"github.com/jmoiron/sqlx"
)

type bonusDb struct {
	db *sqlx.DB
}

// func (b bonusDb) ChangeAward(bonusType string, award int) error {
// 	_, err := b.db.Query(`update bonuses set award = ? where bonus_type = ?`, award, bonusType)
// 	return err
// }

// func (b bonusDb) GetAward(bonus_type string) (int, error) {
// 	var award int
// 	err := b.db.Get(&award, `select award from bonuses where bonus_type = ?`, bonus_type)
// 	return award, err
// }

func (b bonusDb) Create(bonus domain.Bonus) error {
	_, err := b.db.Query(`insert into bonuses (channel_id, award) values (?, ?)`, bonus.ChannelID, bonus.Award)
	return err
}
