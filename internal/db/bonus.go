package db

import "github.com/jmoiron/sqlx"

type bonusType string

const (
	BonusReferral bonusType = "referral"
)

type BonusModel struct {
	ID        int       `json:"id" db:"id"`
	Award     int       `json:"award" db:"award"`
	BonusType bonusType `json:"bonusType" db:"bonus_type"`
}

type bonusRepository interface {
	ChangeAward(bonusType string, award int) error
	GetAward(bonusType string) (int, error)
}

type bonus struct {
	db *sqlx.DB
}

func (b bonus) ChangeAward(bonusType string, award int) error {
	_, err := b.db.Query(`update bonuses set award = ? where bonus_type = ?`, award, bonusType)
	return err
}

func (b bonus) GetAward(bonus_type string) (int, error) {
	var award int
	err := b.db.Get(&award, `select award from bonuses where bonus_type = ?`, bonus_type)
	return award, err
}
