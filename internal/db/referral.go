package db

import (
	"gpt-bot/internal/db/domain"
	"gpt-bot/utils"

	"github.com/jmoiron/sqlx"
)

type referralDb struct {
	db *sqlx.DB
}

func (r referralDb) Create() error {
	code := utils.GenerateReferralCode(utils.AdRefCode)
	_, err := r.db.Exec(`insert into referrals (code) values (?)`, code)
	return err
}

func (r referralDb) GetOne(code string) (int, error) {
	var refID int
	err := r.db.Get(&refID, `select id from referrals where code = ?`, code)
	return refID, err
}

func (r referralDb) GetAll() ([]domain.Referral, error) {
	var refLinks []domain.Referral
	err := r.db.Select(&refLinks, `select * from referrals`)
	return refLinks, err
}

func (r referralDb) Delete(id int) error {
	rows, err := r.db.Query(`delete from referrals where id = ?`, id)
	defer rows.Close()
	return err
}

func (r referralDb) AddUser(userID, refID int) error {
	rows, err := r.db.Query(`insert into user_referrals values (?, ?) on duplicate key update user_id = user_id`, refID, userID)
	defer rows.Close()
	return err
}

func (r referralDb) CountUsers(refID int) (int, error) {
	var countUsers int
	err := r.db.Get(&countUsers, `select count(*) from user_referrals where referral_id = ?`, refID)
	return countUsers, err
}
