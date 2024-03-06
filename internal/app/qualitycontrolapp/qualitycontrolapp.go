package qualitycontrolapp

import (
	"context"

	"github.com/google/uuid"
)

type IdealText struct {
	UUID       uuid.UUID `json:"uuid"`
	FileID     string    `json:"id_file"`
	ChannelTag string    `json:"channelTag"`
	Text       string    `json:"text"`
}

type QualityControl struct {
	ASR       string `json:"asr"`
	TestIdeal string `json:"-"`
	TextASR   string `json:"-"`
	Quality   int    `json:"quality"`
}

type QualityControlStore interface {
	Create(ctx context.Context, qualityControl IdealText) error
	GetTextASRIdeal(ctx context.Context, fileID string) (*[]QualityControl, error)
}

type QualityControls struct {
	QualityControlStore QualityControlStore
}

func NewQualityControl(qualityControlStore QualityControlStore) *QualityControls {
	return &QualityControls{
		QualityControlStore: qualityControlStore,
	}
}

func (qc *QualityControls) Create(ctx context.Context, qualityControl IdealText) error {

	qualityControl.UUID = uuid.New()

	if err := qc.QualityControlStore.Create(ctx, qualityControl); err != nil {
		return err
	}

	return nil
}

func (qc *QualityControls) QualityControl(ctx context.Context, fileID string) (*[]QualityControl, error) {

	data, err := qc.QualityControlStore.GetTextASRIdeal(ctx, fileID)
	if err != nil {
		return nil, err
	}

	return data, nil
}
