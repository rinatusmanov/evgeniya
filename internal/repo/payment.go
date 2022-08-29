package repo

import (
	"context"
	"errors"
	"time"

	"github.com/jmoiron/sqlx"
)

type Payment struct {
	ID             int        `json:"id" db:"id"`
	CreatedAt      *time.Time `json:"created_at" db:"created_at,type:timestamp"`
	RollBackAt     *time.Time `json:"rollback_at" db:"rollback_at,type:timestamp"`
	UserID         int        `json:"user_id" db:"user_id"`
	CurrencyID     int        `json:"currency_id" db:"currency_id"`
	Withdraw       int        `json:"withdraw" db:"withdraw"`
	Deposit        int        `json:"deposit" db:"deposit"`
	TransactionRef string     `json:"transaction_ref" db:"transaction_ref"`
}

func GetCurrencyID(ctx context.Context, db *sqlx.Tx, userID int) (int, error) {
	var payments []Payment
	err := db.SelectContext(
		ctx,
		&payments,
		"SELECT * FROM billing.payments WHERE user_id = $1 limit 1",
		userID,
	)

	if len(payments) != 0 {
		return payments[0].CurrencyID, nil
	}

	return -1, err //nolint:wrapcheck // intentional
}

func NewPayment(
	ctx context.Context,
	db *sqlx.Tx,
	userID, currencyID, withdraw, deposit int,
	transactionRef string,
) (*Payment, error) {
	payment := Payment{
		UserID:         userID,
		CurrencyID:     currencyID,
		Withdraw:       withdraw,
		Deposit:        deposit,
		TransactionRef: transactionRef,
	}

	if errGetContext := db.GetContext(
		ctx,
		&payment,
		"INSERT INTO billing.payments(user_id, currency_id, withdraw, deposit, transaction_ref) VALUES ($1, $2, $3, $4, $5) returning *", //nolint:lll // intentional
		userID,
		currencyID,
		withdraw,
		deposit,
		transactionRef,
	); errGetContext != nil {
		return nil, errGetContext //nolint:wrapcheck // intentional
	}

	return &payment, nil
}

func GetDepositByUserID(ctx context.Context, db *sqlx.Tx, userID int) (int, error) {
	var payments []Payment
	if errSelectContext := db.SelectContext(
		ctx,
		&payments,
		"SELECT * FROM billing.payments WHERE user_id = $1 and rollback_at is NULL",
		userID,
	); errSelectContext != nil {
		return 0, errSelectContext //nolint:wrapcheck // intentional
	}

	var result int

	for _, payment := range payments {
		result += payment.Deposit - payment.Withdraw
	}

	return result, nil
}

var errNotUniqueTransactionRef = errors.New("not unique transactionRef")

func CheckUniqueTransactionRef(ctx context.Context, db *sqlx.Tx, transactionRef string) error {
	var payments []Payment
	if errSelectContext := db.SelectContext(
		ctx,
		&payments,
		"SELECT * FROM billing.payments WHERE transaction_ref = $1",
		transactionRef,
	); errSelectContext != nil {
		return errSelectContext //nolint:wrapcheck // intentional
	}

	if len(payments) != 0 {
		return errNotUniqueTransactionRef
	}

	return nil
}

var errPaymentNotFound = errors.New("payment not found")

func RollbackPayment(ctx context.Context, db *sqlx.Tx, rollbackPayment string) error {
	result, errGetContext := db.ExecContext(
		ctx,
		"UPDATE billing.payments SET rollback_at = now() WHERE transaction_ref = $1",
		rollbackPayment,
	)
	if errGetContext != nil {
		return errGetContext //nolint:wrapcheck // intentional
	}

	if count, _ := result.RowsAffected(); count == 0 {
		return errPaymentNotFound
	}

	return nil
}
