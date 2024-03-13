package audiofilesapp

import (
	"context"
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"time"

	"github.com/RecoBattle/internal/app/asr"
	"github.com/google/uuid"
	"github.com/sirupsen/logrus"
)

const (
	StatusNEW        = "NEW"
	StatusPROCESSING = "PROCESSING"
	StatusINVALID    = "INVALID"
	StatusPROCESSED  = "PROCESSED"

	//ASRYaSpeachKit = "yandexSpeachKit"
	//ASRSalut       = "salut"
	//ASRVosk        = "vosk"
	//ASR3iTech      = "3iTech"
)

type AudioFile struct {
	UUID       uuid.UUID `json:"uuid"`
	FileID     string    `json:"id_file"`
	FileName   string    `json:"file_name"`
	ASR        string    `json:"asr"`
	Status     string    `json:"status"`
	UploadedAt time.Time `json:"uploaded_at"`
	UserID     string    `json:"-"`
}

type ResultASR struct {
	UUID       uuid.UUID `json:"-"`
	ChannelTag string    `json:"channelTag"`
	Text       string    `json:"text"`
	StartTime  float32   `json:"startTime"`
	EndTime    float32   `json:"endTime"`
}

type AudioFileStore interface {
	CreateFile(ctx context.Context, audioFile AudioFile) error
	CreateASR(ctx context.Context, audioFile AudioFile) error
	UpdateStatusASR(ctx context.Context, audioFileUUID, status string) error
	CreateResultASR(ctx context.Context, resultASR ResultASR) error
	GetAudioFiles(ctx context.Context, userID string) (*[]AudioFile, error)
	GetResultASR(ctx context.Context, uuid string) (*[]ResultASR, error)
}

type AudioFiles struct {
	audioFileStore AudioFileStore
}

func NewAudioFile(audioFileStore AudioFileStore) *AudioFiles {
	return &AudioFiles{
		audioFileStore: audioFileStore,
	}
}

func (af *AudioFiles) Create(ctx context.Context, audiofile AudioFile) (string, error) {

	audiofile.FileID = hex.EncodeToString(af.writeHash(audiofile.FileName, audiofile.UserID))

	if err := af.audioFileStore.CreateFile(ctx, audiofile); err != nil {
		return "", err
	}

	return audiofile.FileID, nil
}

func (af *AudioFiles) AddASRProcessing(audiofile AudioFile, asr asr.ASR, data []byte) {

	ctx := context.Background()

	audiofile.UUID = uuid.New()
	if err := af.audioFileStore.CreateASR(ctx, audiofile); err != nil {
		logrus.Error(err.Error())
		return
	}

	time.Sleep(30 * time.Second)

	if err := af.audioFileStore.UpdateStatusASR(ctx, audiofile.UUID.String(), StatusPROCESSING); err != nil {
		logrus.Error(err.Error())
		return
	}

	result, err := asr.TextFromASRModel(data)
	if err != nil {
		logrus.Errorf("error in sending request to ASR. error: %#v", result)
		if err := af.audioFileStore.UpdateStatusASR(ctx, audiofile.UUID.String(), StatusINVALID); err != nil {
			return
		}
		return
	}

	resASR := ResultASR{
		UUID:       audiofile.UUID,
		ChannelTag: "1",
		Text:       result,
	}

	if err := af.audioFileStore.CreateResultASR(ctx, resASR); err != nil {
		logrus.Errorf("error in writing the ASR result. error: %v", err)
		if err := af.audioFileStore.UpdateStatusASR(ctx, audiofile.UUID.String(), StatusINVALID); err != nil {
			return
		}
		return
	}

	if err := af.audioFileStore.UpdateStatusASR(ctx, audiofile.UUID.String(), StatusPROCESSED); err != nil {
		logrus.Error(err.Error())
		return
	}

}

func (af *AudioFiles) GetAudioFiles(ctx context.Context, userID string) (*[]AudioFile, error) {

	files, err := af.audioFileStore.GetAudioFiles(ctx, userID)

	if err != nil {
		return nil, err
	}

	return files, nil
}

func (af *AudioFiles) GetResultASR(ctx context.Context, uuid string) (*[]ResultASR, error) {

	resultASR, err := af.audioFileStore.GetResultASR(ctx, uuid)

	if err != nil {
		return nil, err
	}

	return resultASR, nil
}

func (af *AudioFiles) writeHash(filename, userID string) []byte {

	secretKey := "file2468"

	hash := hmac.New(sha256.New, []byte(secretKey))
	hash.Write([]byte(fmt.Sprintf("%s:%s:%s", filename, userID, secretKey)))

	return hash.Sum(nil)
}
