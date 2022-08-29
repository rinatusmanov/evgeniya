package repo

import (
	"context"

	"github.com/jmoiron/sqlx"
)

type Operator struct {
	ID int `json:"id" db:"id"`
}

func NewOperator(ctx context.Context, db *sqlx.DB, id int) (*Operator, error) {
	operator := Operator{
		ID: id,
	}

	if errGetContext := db.GetContext(
		ctx,
		&operator,
		"INSERT INTO public.operators(id) VALUES ($1) returning *",
		id,
	); errGetContext != nil {
		return nil, errGetContext //nolint:wrapcheck // intentional
	}

	return &operator, nil
}
