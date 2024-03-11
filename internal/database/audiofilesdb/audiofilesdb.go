package audiofilesdb

import (
	"context"
	"database/sql"
	"errors"
	"time"

	"github.com/Masterminds/squirrel"
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

	_, err = db.ExecContext(ctx, `CREATE TABLE IF NOT EXISTS result_asr (
		"uuid" TEXT,
		"channelTag" TEXT,
		"text" TEXT,
		"startTime" REAL,
		"endTime" REAL,
		FOREIGN KEY (uuid) REFERENCES asr(uuid)
	  )`)

	if err != nil {
		return nil, err
	}

	return &AudioFileStore{db: db}, nil
}

func (d *AudioFileStore) CreateFile(ctx context.Context, audioFile audiofilesapp.AudioFile) error {

	tx, err := d.db.Begin()
	if err != nil {
		return err
	}

	defer tx.Rollback()

	_, err = tx.ExecContext(ctx, "INSERT INTO audiofiles (file_id, file_name, user_id, uploaded_at) VALUES($1,$2,$3,$4)", audioFile.FileID, audioFile.FileName, audioFile.UserID, time.Now())

	if err != nil {
		var pgErr *pgconn.PgError
		if errors.As(err, &pgErr) && pgerrcode.IsIntegrityConstraintViolation(pgErr.Code) {
			return database.NewErrorConflict(err)
		}

		return err
	}

	return tx.Commit()
}

func (d *AudioFileStore) CreateASR(ctx context.Context, audioFile audiofilesapp.AudioFile) error {

	tx, err := d.db.Begin()
	if err != nil {
		return err
	}

	defer tx.Rollback()

	_, err = tx.ExecContext(ctx, "INSERT INTO asr (uuid, file_id, asr, status) VALUES($1,$2,$3,$4)", audioFile.UUID.String(), audioFile.FileID, audioFile.ASR, audiofilesapp.StatusNEW)

	if err != nil {
		return err
	}

	return tx.Commit()
}

func (d *AudioFileStore) UpdateStatusASR(ctx context.Context, audioFileUUID, status string) error {

	tx, err := d.db.Begin()
	if err != nil {
		return err
	}

	defer tx.Rollback()

	_, err = tx.ExecContext(ctx, "UPDATE asr SET status=$1 WHERE uuid=$2", status, audioFileUUID)

	if err != nil {
		return err
	}

	return tx.Commit()
}

func (d *AudioFileStore) CreateResultASR(ctx context.Context, resultASR audiofilesapp.ResultASR) error {

	tx, err := d.db.Begin()
	if err != nil {
		return err
	}

	defer tx.Rollback()

	_, err = tx.ExecContext(ctx, "INSERT INTO result_asr (uuid, channelTag, text, startTime, endTime) VALUES($1,$2,$3,$4,$5)",
		resultASR.UUID.String(), resultASR.ChannelTag, resultASR.Text, resultASR.StartTime, resultASR.EndTime)

	if err != nil {
		return err
	}

	return tx.Commit()
}

func (d *AudioFileStore) GetAudioFiles(ctx context.Context, userID string) (*[]audiofilesapp.AudioFile, error) {

	var rows *sql.Rows

	qb := squirrel.StatementBuilder.PlaceholderFormat(squirrel.Dollar)

	_, err := d.db.ExecContext(ctx, "CREATE TEMP TABLE temp_audiofiles AS SELECT (file_id, file_name, uploaded_at) FROM audiofiles WHERE user_id=$1;", userID)
	if err != nil {
		return nil, err
	}

	rows, err = qb.Select("a.file_id", "a.file_name", "a.uploaded_at", "b.uuid", "b.asr", "b.status").
		From("temp_audiofiles a").
		LeftJoin("asr b ON a.file_id = b.file_id").
		OrderBy("a.uploaded_at DESC").
		RunWith(d.db).
		Query()

	if err != nil {
		return nil, err
	}

	if rows.Err() != nil {
		return nil, rows.Err()
	}

	defer rows.Close()

	var files []audiofilesapp.AudioFile

	for rows.Next() {

		var file audiofilesapp.AudioFile
		if err = rows.Scan(&file.FileID, &file.FileName, &file.UploadedAt, &file.UUID, &file.ASR, &file.Status); err != nil {
			return nil, err
		}
		files = append(files, file)
	}

	return &files, nil
}

func (d *AudioFileStore) GetResultASR(ctx context.Context, uuid string) (*[]audiofilesapp.ResultASR, error) {

	var rows *sql.Rows

	qb := squirrel.StatementBuilder.PlaceholderFormat(squirrel.Dollar)

	rows, err := qb.Select("channelTag", "text", "startTime", "endTime").
		From("result_asr").
		Where(squirrel.Eq{"uuid": uuid}).
		RunWith(d.db).
		Query()

	if err != nil {
		return nil, err
	}

	if rows.Err() != nil {
		return nil, rows.Err()
	}

	defer rows.Close()

	var resASR []audiofilesapp.ResultASR

	for rows.Next() {

		var res audiofilesapp.ResultASR
		if err = rows.Scan(&res.ChannelTag, &res.Text, &res.StartTime, &res.EndTime); err != nil {
			return nil, err
		}
		resASR = append(resASR, res)
	}

	return &resASR, nil
}
