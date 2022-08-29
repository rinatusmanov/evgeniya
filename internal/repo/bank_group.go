package repo

import (
	"context"
	"time"

	"github.com/jmoiron/sqlx"
)

type BankGroup struct {
	ID         string     `json:"id" db:"id,type:text"`
	CreatedAt  *time.Time `json:"created_at" db:"created_at,type:timestamp"`
	UpdatedAt  *time.Time `json:"updated_at" db:"updated_at,type:timestamp"`
	CurrencyID int        `json:"currency_id" db:"currency_id"`
}

func GetBankGroups(ctx context.Context, db *sqlx.Tx) ([]BankGroup, error) {
	var bankGroup []BankGroup
	err := db.SelectContext(ctx, &bankGroup, "SELECT * FROM billing.ref_bank_groups")

	return bankGroup, err //nolint:wrapcheck // intentional
}

func GetBankByCurrencyID(ctx context.Context, db *sqlx.Tx, currencyID int) (BankGroup, error) {
	var bankGroup BankGroup
	err := db.SelectContext(
		ctx,
		&bankGroup,
		"SELECT * FROM billing.ref_bank_groups WHERE currency_id = $1 LIMIT 1",
		currencyID,
	)

	return bankGroup, err //nolint:wrapcheck // intentional
}

func NewBankGroup(ctx context.Context, db *sqlx.Tx, id string, currencyID int) (*BankGroup, error) {
	bankGroup := BankGroup{
		ID:         id,
		CurrencyID: currencyID,
	}

	if errGetContext := db.GetContext(
		ctx,
		&bankGroup,
		"INSERT INTO billing.ref_bank_groups(id, currency_id) VALUES ($1, $2) returning *",
		id,
		currencyID,
	); errGetContext != nil {
		return nil, errGetContext //nolint:wrapcheck // intentional
	}

	return &bankGroup, nil
}
