package qualitycontrolapp

import (
	"context"
	"strings"
	"unicode"

	"github.com/google/uuid"
)

type IdealText struct {
	UUID       uuid.UUID `json:"uuid"`
	FileID     string    `json:"id_file"`
	ChannelTag string    `json:"channelTag"`
	Text       string    `json:"text"`
}

type QualityControl struct {
	ASR       string  `json:"asr"`
	TestIdeal string  `json:"-"`
	TextASR   string  `json:"-"`
	Quality   float32 `json:"quality"`
}

type QualityControlStore interface {
	Create(ctx context.Context, qualityControl IdealText) error
	GetTextASRIdeal(ctx context.Context, fileID string) ([]QualityControl, string, error)
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

	data, idealText, err := qc.QualityControlStore.GetTextASRIdeal(ctx, fileID)
	if err != nil {
		return nil, err
	}

	idealText = removeSpecialCharacters(idealText)

	for i := range data {
		resASR := removeSpecialCharacters(data[i].TextASR)
		data[i].Quality = compareStrings(idealText, resASR)
		data[i].TestIdeal = idealText
	}

	return &data, nil
}

func removeSpecialCharacters(s string) string {

	var cleanedString strings.Builder

	for _, r := range s {
		if unicode.IsLetter(r) || unicode.IsSpace(r) {
			cleanedString.WriteRune(r)
		}
	}

	return strings.ToLower(cleanedString.String())
}

func compareStrings(str1, str2 string) float32 {

	words1 := strings.Fields(str1)
	words2 := strings.Fields(str2)

	minLength := len(words1)

	if len(words2) < minLength {
		minLength = len(words2)
	}

	matches := 0

	for i := 0; i < minLength; i++ {
		if words1[i] == words2[i] {
			matches++
		}
	}

	similarity := float32(matches) / float32(minLength)

	return similarity
}
