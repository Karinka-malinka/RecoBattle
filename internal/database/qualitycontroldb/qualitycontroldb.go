package qualitycontroldb

import (
	"context"
	"database/sql"
	"errors"

	"github.com/RecoBattle/internal/app/qualitycontrolapp"
	"github.com/RecoBattle/internal/database"
	"github.com/jackc/pgerrcode"
	"github.com/jackc/pgx/v5/pgconn"
)

var _ qualitycontrolapp.QualityControlStore = &QualityControlStore{}

type QualityControlStore struct {
	db *sql.DB
}

func NewQCStore(ctx context.Context, db *sql.DB) (*QualityControlStore, error) {

	_, err := db.ExecContext(ctx, `CREATE TABLE IF NOT EXISTS quality_control (
		"uuid" TEXT PRIMARY KEY,
		"file_id" TEXT PRIMARY KEY,
		"channelTag" TEXT,
		"text" TEXT,
		FOREIGN KEY (file_id) REFERENCES audiofiles(file_id)
	  )`)

	if err != nil {
		return nil, err
	}

	return &QualityControlStore{db: db}, nil
}

func (d *QualityControlStore) Create(ctx context.Context, it qualitycontrolapp.IdealText) error {

	tx, err := d.db.Begin()
	if err != nil {
		return err
	}

	defer tx.Rollback()

	_, err = tx.ExecContext(ctx, "INSERT INTO quality_control (uuid, file_id, channelTag, text) VALUES($1,$2,$3,$4)", it.UUID.String(), it.FileID, it.ChannelTag, it.Text)

	if err != nil {
		var pgErr *pgconn.PgError
		if errors.As(err, &pgErr) && pgerrcode.IsIntegrityConstraintViolation(pgErr.Code) {
			return database.NewErrorConflict(err)
		}

		return err
	}

	return tx.Commit()
}

func (d *QualityControlStore) GetTextASRIdeal(ctx context.Context, fileID string) (*[]qualitycontrolapp.QualityControl, error) {

}
