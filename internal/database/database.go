package database

import (
	"context"
	"database/sql"
	"log"

	"github.com/golang-migrate/migrate/v4"
	"github.com/golang-migrate/migrate/v4/database/postgres"
	_ "github.com/golang-migrate/migrate/v4/source/file"
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

	driver, err := postgres.WithInstance(db, &postgres.Config{})
	if err != nil {
		log.Fatal("Error when creating the driver:", err)
	}

	m, err := migrate.NewWithDatabaseInstance("file://../migrations", "postgres", driver)
	if err != nil {
		log.Fatal("Error initializing migrations:", err)
	}

	err = m.Up()
	if err != nil && err != migrate.ErrNoChange {
		log.Fatal("Error during migration:", err)
	}

	d := PostgresDatabase{DB: db}

	return &d, nil
}

func (d *PostgresDatabase) Close() error {
	return d.DB.Close()
}
