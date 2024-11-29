package db

import (
	"fmt"
	"gpt-bot/internal/db/domain"
	"strings"

	"github.com/jmoiron/sqlx"
)

type limitsDb struct {
	db *sqlx.DB
}

func (l limitsDb) Create(limits domain.Limits) error {
	rows, err := l.db.Query(`insert into limits values (?, ?, ?, ?, ?, ?, ?) on duplicate key update user_id = user_id`, limits.UserID, limits.O1Preview, limits.Gpt4o, limits.O1Mini, limits.Gpt4oMini, limits.Runware, limits.Dalle3)
	defer rows.Close()
	return err
}

func (l limitsDb) Update(newLimits domain.Limits) error {
	rows, err := l.db.Query(`update limits set o1_preview=?, gpt_4o=?, o1_mini=?, gpt_4o_mini=?, runware=?, dall_e_3=? where user_id = ?`, newLimits.O1Preview, newLimits.Gpt4o, newLimits.O1Mini, newLimits.Gpt4oMini, newLimits.Runware, newLimits.Dalle3, newLimits.UserID)
	defer rows.Close()
	return err
}

func (l limitsDb) Reduce(userID int, model string) error {
	model = strings.Replace(model, "-", "_", -1)
	query := fmt.Sprintf(`update limits set %s = %s - 1 where user_id = %d`, model, model, userID)
	rows, err := l.db.Query(query)
	defer rows.Close()
	return err
}

func (l limitsDb) GetLimitsByModel(userID int, model string) (int, error) {
	model = strings.Replace(model, "-", "_", -1)
	query := fmt.Sprintf(`select %s from limits where user_id = %d`, model, userID)
	var modelLimits int
	err := l.db.Get(&modelLimits, query)
	return modelLimits, err
}

func (l limitsDb) GetByUser(userID int) (domain.Limits, error) {
	var limits domain.Limits
	err := l.db.Get(&limits, `select * from limits where user_id = ?`, userID)
	return limits, err
}
