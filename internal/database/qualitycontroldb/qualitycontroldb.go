package qualitycontroldb

import (
	"context"
	"database/sql"
	"errors"

	"github.com/Masterminds/squirrel"
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

	return &QualityControlStore{db: db}, nil
}

func (d *QualityControlStore) Create(ctx context.Context, it qualitycontrolapp.IdealText) error {

	tx, err := d.db.Begin()
	if err != nil {
		return err
	}

	defer tx.Rollback()

	_, err = tx.ExecContext(ctx, "INSERT INTO quality_control (uuid, file_id, channel_tag, text) VALUES($1,$2,$3,$4)", it.UUID.String(), it.FileID, it.ChannelTag, it.Text)

	if err != nil {
		var pgErr *pgconn.PgError
		if errors.As(err, &pgErr) && pgerrcode.IsIntegrityConstraintViolation(pgErr.Code) {
			return database.NewErrorConflict(err)
		}

		return err
	}

	return tx.Commit()
}

func (d *QualityControlStore) GetTextASRIdeal(ctx context.Context, fileID string) ([]qualitycontrolapp.QualityControl, string, error) {

	var rows *sql.Rows
	var qcs []qualitycontrolapp.QualityControl

	qb := squirrel.StatementBuilder.PlaceholderFormat(squirrel.Dollar)

	query, args, err := qb.Select("text").
		From("quality_control").
		Where(squirrel.Eq{"file_id": fileID}).
		ToSql()

	if err != nil {
		return nil, "", err
	}

	row := d.db.QueryRowContext(ctx, query, args...)

	if row.Err() != nil {
		return nil, "", row.Err()
	}

	var idealText string
	if err = row.Scan(&idealText); err != nil {
		return qcs, "", nil
	}

	rows, err = qb.Select("asr.asr", "res.text").
		From("asr").
		InnerJoin("result_asr res ON asr.uuid = res.uuid").
		Where(squirrel.Eq{"file_id": fileID}).
		RunWith(d.db).
		Query()

	if err != nil {
		return nil, "", err
	}

	if rows.Err() != nil {
		return nil, "", rows.Err()
	}

	defer rows.Close()

	for rows.Next() {
		var qc qualitycontrolapp.QualityControl
		if err = rows.Scan(&qc.ASR, &qc.TextASR); err != nil {
			return nil, "", err
		}
		qcs = append(qcs, qc)
	}

	return qcs, idealText, nil
}
