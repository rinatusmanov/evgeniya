package repo

import (
	"context"
	"errors"
	"time"

	"github.com/jmoiron/sqlx"
)

type Currency struct {
	ID                                     int        `json:"id" db:"id,pk"`
	CreatedAt                              *time.Time `json:"created_at" db:"created_at,type:timestamp"`
	UpdatedAt                              *time.Time `json:"updated_at" db:"updated_at,type:timestamp"`
	Code                                   string     `json:"code" db:"code,type:text"`
	NumberOfDigitsAfterTheDecimalSeparator int        `json:"number_of_digits_after_the_decimal_separator" db:"number_of_digits_after_the_decimal_separator,type:int"` //nolint:lll // intentional
	Name                                   string     `json:"name" db:"name,type:text"`
}

func NewCurrency(
	ctx context.Context,
	db *sqlx.Tx,
	code string,
	numberOfDigitsAfterTheDecimalSeparator int,
	name string,
) (Currency, error) {
	currency := Currency{
		Code:                                   code,
		NumberOfDigitsAfterTheDecimalSeparator: numberOfDigitsAfterTheDecimalSeparator,
		Name:                                   name,
	}
	err := db.GetContext(
		ctx,
		&currency,
		"INSERT INTO billing.ref_currency(code, number_of_digits_after_the_decimal_separator, name) VALUES ($1, $2, $3) returning *", //nolint:lll // intentional
		code,
		numberOfDigitsAfterTheDecimalSeparator,
		name,
	)

	return currency, err //nolint:wrapcheck // intentional
}

func GetCurrencies(ctx context.Context, db *sqlx.Tx) ([]Currency, error) {
	var currencies []Currency
	err := db.SelectContext(ctx, &currencies, "SELECT * FROM billing.ref_currency")

	return currencies, err //nolint:wrapcheck // intentional
}

func GetCurrencyByID(ctx context.Context, db *sqlx.Tx, id int) (Currency, error) {
	var currency Currency
	err := db.GetContext(ctx, &currency, "SELECT * FROM billing.ref_currency WHERE id = $1", id)

	return currency, err //nolint:wrapcheck // intentional
}

var errNotFoundCurrency = errors.New("not found currency")

func GetCurrencyByCode(ctx context.Context, db *sqlx.Tx, code string) (*Currency, error) {
	var currencies []Currency

	err := db.SelectContext(ctx, &currencies, "SELECT * FROM billing.ref_currency WHERE code = $1 LIMIT 1", code)
	if err != nil {
		return nil, err //nolint:wrapcheck // intentional
	}

	if len(currencies) == 0 {
		return nil, errNotFoundCurrency
	}

	return &currencies[0], nil
}
