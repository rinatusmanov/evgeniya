package repo

import (
	"context"
	"errors"

	"github.com/jmoiron/sqlx"
)

type Bonus struct {
	Bonus string `json:"bonus" db:"type:text;"`
	Used  bool   `json:"used" db:"type:boolean;"`
}

func WriteBonus(ctx context.Context, db *sqlx.Tx, bonus string) error {
	if _, errExecContext := db.ExecContext(
		ctx,
		"INSERT INTO billing.bonuses(bonus) VALUES ($1)",
		bonus,
	); errExecContext != nil {
		return errExecContext //nolint:wrapcheck // intentional
	}

	return nil
}

var errBonusNotFound = errors.New("bonus not found")

func UseBonus(ctx context.Context, db *sqlx.Tx, bonus string) error {
	result, errExecContext := db.ExecContext(
		ctx,
		"UPDATE billing.bonuses SET used = true WHERE bonus = $1",
		bonus,
	)
	if errExecContext != nil {
		return errExecContext //nolint:wrapcheck // intentional
	}

	if count, _ := result.RowsAffected(); count == 0 {
		return errBonusNotFound
	}

	return nil
}
