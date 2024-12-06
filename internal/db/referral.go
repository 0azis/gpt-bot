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
	_, err := r.db.Exec(`insert into referrals (name, code) values (?, ?)`, "", code)
	return err
}

func (r referralDb) GetOne(code string) (int, error) {
	var refID int
	rows, err := r.db.Query(`select id from referrals where code = ?`, code)
	if err != nil {
		return refID, err
	}
	defer rows.Close()

	for rows.Next() {
		err = rows.Scan(&refID)
		if err != nil {
			return refID, err
		}
	}

	return refID, err
}

func (r referralDb) GetOneByID(id int) (domain.Referral, error) {
	var ref domain.Referral
	rows, err := r.db.Query(`select * from referrals where id = ?`, id)
	if err != nil {
		return ref, err
	}
	defer rows.Close()

	for rows.Next() {
		err = rows.Scan(&ref.ID, &ref.Name, &ref.Code)
		if err != nil {
			return ref, err
		}
	}

	return ref, nil
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
	date := utils.Timestamp()
	rows, err := r.db.Query(`insert into user_referrals values (?, ?, ?) on duplicate key update user_id = user_id`, refID, userID, date)
	defer rows.Close()
	return err
}

func (r referralDb) CountUsers(refID int) (int, error) {
	var countUsers int
	rows, err := r.db.Query(`select count(*) from user_referrals where referral_id = ?`, refID)
	if err != nil {
		return countUsers, err
	}
	defer rows.Close()

	for rows.Next() {
		err = rows.Scan(&countUsers)
		if err != nil {
			return countUsers, err
		}
	}

	return countUsers, nil
}

func (r referralDb) ActiveUsers(code string) (int, error) {
	var countUsers int
	rows, err := r.db.Query(`select count(distinct user_referrals.user_id) from user_referrals join referrals on referrals.id = user_referrals.referral_id join chats on chats.user_id = user_referrals.user_id where referrals.code = ? having count(chats.id) >= 1`, code)
	if err != nil {
		return countUsers, err
	}
	defer rows.Close()

	for rows.Next() {
		err = rows.Scan(&countUsers)
		if err != nil {
			return countUsers, err
		}
	}

	return countUsers, nil
}

func (r referralDb) RunMiniApp(code string) (int, error) {
	var countUsers int
	rows, err := r.db.Query(`select count(distinct user_referrals.user_id) from user_referrals join referrals on referrals.id = user_referrals.referral_id join stats on stats.user_id = user_referrals.user_id where referrals.code = ? having count(stats.user_id) >= 1`, code)
	if err != nil {
		return countUsers, err
	}
	defer rows.Close()

	for rows.Next() {
		err = rows.Scan(&countUsers)
		if err != nil {
			return countUsers, err
		}
	}

	return countUsers, nil
}

func (r referralDb) NotRunMiniApp(code string) (int, error) {
	var countUsers int
	err := r.db.Get(&countUsers, `select count(distinct user_referrals.user_id) from user_referrals join referrals on referrals.id = user_referrals.referral_id join stats on stats.user_id = user_referrals.user_id where referrals.code = ? having count(stats.user_id) < 1`, code)
	return countUsers, err
}

func (r referralDb) UpdateCode(id int, code string) error {
	_, err := r.db.Exec(`update referrals set code = ? where id = ?`, code, id)
	return err
}

func (r referralDb) UpdateName(id int, name string) error {
	_, err := r.db.Exec(`update referrals set name = ? where id = ?`, name, id)
	return err
}

func (r referralDb) AllUsers() (int, error) {
	var countUsers int
	err := r.db.Get(&countUsers, `select count(*) from user_referrals`)
	return countUsers, err
}

func (r referralDb) MonthlyUsers() (int, error) {
	var countUsers int
	err := r.db.Get(&countUsers, `select count(*) from user_referrals where date(created_at) >= date_sub(curdate(), interval dayofmonth(curdate())-1 day)`)
	return countUsers, err
}

func (r referralDb) DailyUsers() (int, error) {
	var countUsers int
	err := r.db.Get(&countUsers, `select count(*) from user_referrals where date(created_at) >= curdate()`)
	return countUsers, err
}
