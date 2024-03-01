package database

import (
	"context"
	"database/sql"

	_ "github.com/jackc/pgx/v5/stdlib"
)

type PostgresDatabase struct {
	DB *sql.DB
}

func NewDB(ctx context.Context, ps string) (*PostgresDatabase, error) {

	db, err := sql.Open("pgx", ps)
	if err != nil {
		return nil, err
	}

	_, err = db.ExecContext(ctx, `CREATE TABLE IF NOT EXISTS users (
		        "uuid" TEXT PRIMARY KEY,
				"login" TEXT,
				"hash_pass" TEXT,
				UNIQUE (login)
		      )`)

	if err != nil {
		return nil, err
	}

	/*
			_, err = db.ExecContext(ctx, `CREATE TABLE IF NOT EXISTS shorten (
		        "uuid" TEXT PRIMARY KEY,
				"user_id" TEXT,
				"original_url" TEXT,
		        "short_url" TEXT,
				"correlation_id" TEXT,
				"is_deleted" BOOLEAN DEFAULT false,
				UNIQUE (original_url, correlation_id),
				FOREIGN KEY (user_id) REFERENCES users(uuid)
		      )`)

			if err != nil {
				return nil, err
			}

			_, err = db.ExecContext(ctx, "CREATE INDEX IF NOT EXISTS ix_original_url ON shorten (original_url)")
			if err != nil {
				return nil, err
			}
	*/

	d := PostgresDatabase{DB: db}

	return &d, nil
}

func (d *PostgresDatabase) Close() error {
	return d.DB.Close()
}
