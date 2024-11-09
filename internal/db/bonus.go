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
	ChangeReferralAward(award int) error
	GetReferralAward() (int, error)
}

type bonus struct {
	db *sqlx.DB
}

func (b bonus) ChangeReferralAward(award int) error {
	_, err := b.db.Query(`update bonuses set award = ? where bonus_type = 'referral'`, award)
	return err
}

func (b bonus) GetReferralAward() (int, error) {
	var award int
	err := b.db.Get(&award, `select award from bonuses where bonus_type = 'referral'`)
	return award, err
}
