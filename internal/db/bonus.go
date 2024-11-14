package db

import "github.com/jmoiron/sqlx"

type bonusDb struct {
	db *sqlx.DB
}

func (b bonusDb) ChangeAward(bonusType string, award int) error {
	_, err := b.db.Query(`update bonuses set award = ? where bonus_type = ?`, award, bonusType)
	return err
}

func (b bonusDb) GetAward(bonus_type string) (int, error) {
	var award int
	err := b.db.Get(&award, `select award from bonuses where bonus_type = ?`, bonus_type)
	return award, err
}
