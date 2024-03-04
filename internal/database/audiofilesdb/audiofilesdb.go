package audiofilesdb

import (
	"context"
	"database/sql"
	"errors"
	"time"

	"github.com/RecoBattle/internal/app/audiofilesapp"
	"github.com/RecoBattle/internal/database"
	"github.com/jackc/pgerrcode"
	"github.com/jackc/pgx/v5/pgconn"
)

var _ audiofilesapp.AudioFileStore = &AudioFileStore{}

type AudioFileStore struct {
	db *sql.DB
}

func NewAudioFileStore(ctx context.Context, db *sql.DB) (*AudioFileStore, error) {

	_, err := db.ExecContext(ctx, `CREATE TABLE IF NOT EXISTS audiofiles (
		"file_id" TEXT PRIMARY KEY,
		"file_name" TEXT,
		"user_id" TEXT,
		"uploaded_at" TIMESTAMP,
		FOREIGN KEY (user_id) REFERENCES users(uuid)
	  )`)

	if err != nil {
		return nil, err
	}

	_, err = db.ExecContext(ctx, `CREATE TABLE IF NOT EXISTS asr (
		"uuid" TEXT PRIMARY KEY,
		"file_id" TEXT,
		"asr" TEXT,
		"status" TEXT,
		FOREIGN KEY (file_id) REFERENCES audiofiles(file_id)
	  )`)

	if err != nil {
		return nil, err
	}

	return &AudioFileStore{db: db}, nil
}

func (d *AudioFileStore) Create(ctx context.Context, audioFile audiofilesapp.AudioFile) error {

	tx, err := d.db.Begin()
	if err != nil {
		return err
	}

	defer tx.Rollback()

	_, err = tx.ExecContext(ctx, "INSERT INTO audiofiles (file_id, file_name, user_id, uploaded_at) VALUES($1,$2,$3,$4)", audioFile.FileID, audioFile.FileName, audioFile.Status, time.Now())

	if err != nil {
		var pgErr *pgconn.PgError
		if errors.As(err, &pgErr) && pgerrcode.IsIntegrityConstraintViolation(pgErr.Code) {
			return database.NewErrorConflict(err)
		}

		return err
	}

	return tx.Commit()
}
