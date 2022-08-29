package repo

import (
	"context"
	"time"

	"github.com/jmoiron/sqlx"
)

type User struct {
	ID        int       `db:"id"`
	CreatedAt time.Time `db:"created_at"`
	Name      string    `db:"name"`
}

func FindUserByID(ctx context.Context, db *sqlx.Tx, id int) (*User, error) {
	var user User

	if errGetContext := db.GetContext(ctx, &user, "SELECT * FROM public.users WHERE id = $1", id); errGetContext != nil {
		return nil, errGetContext //nolint:wrapcheck // intentional
	}

	return &user, nil
}

func FindUserByName(ctx context.Context, db *sqlx.Tx, name string) (*User, error) {
	var user User

	if errGetContext := db.GetContext(
		ctx,
		&user,
		"SELECT * FROM public.users WHERE name = $1",
		name,
	); errGetContext != nil {
		return nil, errGetContext //nolint:wrapcheck // intentional
	}

	return &user, nil
}

func NewUser(ctx context.Context, db *sqlx.Tx, name string) (*User, error) {
	user := User{
		Name: name,
	}

	if errGetContext := db.GetContext(
		ctx,
		&user,
		"INSERT INTO public.users(\"name\") VALUES ($1) returning *",
		name,
	); errGetContext != nil {
		return nil, errGetContext //nolint:wrapcheck // intentional
	}

	return &user, nil
}
