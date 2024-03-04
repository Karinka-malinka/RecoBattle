package audiofilesapp

import (
	"context"
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"time"

	"github.com/google/uuid"
)

const (
	StatusNEW        = "NEW"
	StatusPROCESSING = "PROCESSING"
	StatusINVALID    = "INVALID"
	StatusPROCESSED  = "PROCESSED"

	ASRYaSpeachKit = "yandexSpeachKit"
	ASRSalut       = "salut"
	ASRVosk        = "vosk"
	ASR3iTech      = "3iTech"
)

type AudioFile struct {
	UUID       uuid.UUID
	FileID     string
	FileName   string
	ASR        string
	Status     string
	UploadedAt time.Time
	UserID     uuid.UUID
}

type AudioFileStore interface {
	Create(ctx context.Context, audioFile AudioFile) error
	//Read(ctx context.Context, userID string) (*[]AudioFile, error)
}

type AudioFiles struct {
	audioFileStore AudioFileStore
}

func NewAudioFile(audioFileStore AudioFileStore) *AudioFiles {
	return &AudioFiles{
		audioFileStore: audioFileStore,
	}
}

func (af *AudioFiles) Create(ctx context.Context, audiofile AudioFile) error {

	audiofile.FileID = hex.EncodeToString(af.writeHash(audiofile.FileName, audiofile.UserID.String()))

	if err := af.audioFileStore.Create(ctx, audiofile); err != nil {
		return err
	}

	return nil
}

func (af *AudioFiles) writeHash(filename, userID string) []byte {

	secretKey := "file2468"

	hash := hmac.New(sha256.New, []byte(secretKey))
	hash.Write([]byte(fmt.Sprintf("%s:%s:%s", filename, userID, secretKey)))

	return hash.Sum(nil)
}
