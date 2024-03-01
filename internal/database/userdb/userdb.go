package userdb

import (
	"context"
	"database/sql"
	"errors"

	"github.com/Masterminds/squirrel"
	"github.com/RecoBattle/internal/app/userapp"
	"github.com/jackc/pgerrcode"
	"github.com/jackc/pgx/v5/pgconn"

	_ "github.com/jackc/pgx/v5/stdlib"
)

var _ userapp.UserStore = &UserStore{}

type UserStore struct {
	db *sql.DB
}

func NewUserStore(db *sql.DB) *UserStore {

	return &UserStore{db: db}
}

func (d *UserStore) Create(ctx context.Context, user userapp.User) error {

	tx, err := d.db.Begin()
	if err != nil {
		return err
	}

	defer tx.Rollback()

	_, err = tx.ExecContext(ctx, "INSERT INTO users (uuid, login, hash_pass) VALUES($1,$2,$3)", user.UUID.String(), user.Username, user.Password)

	if err != nil {
		var pgErr *pgconn.PgError
		if errors.As(err, &pgErr) && pgerrcode.IsIntegrityConstraintViolation(pgErr.Code) {
			return NewErrorConflict(err)
		}

		return err
	}

	return tx.Commit()
}

func (d *UserStore) Read(ctx context.Context, login string) (*userapp.User, error) {

	var rows *sql.Rows

	qb := squirrel.StatementBuilder.PlaceholderFormat(squirrel.Dollar)

	query, args, err := qb.Select("uuid, login, hash_pass").
		From("users").
		Where(squirrel.Eq{"login": login}).
		ToSql()

	if err != nil {
		return nil, err
	}

	rows, err = d.db.QueryContext(ctx, query, args...)

	if err != nil {
		return nil, err
	}

	if rows.Err() != nil {
		return nil, rows.Err()
	}

	defer rows.Close()

	var user userapp.User

	for rows.Next() {

		if err = rows.Scan(&user.UUID, &user.Username, &user.Password); err != nil {
			return nil, errors.New("401")
		}
	}

	return &user, nil
}