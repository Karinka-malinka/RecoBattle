package audiofilesapp

import (
	"context"
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
	FileID     uuid.UUID
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
	audioFile AudioFileStore
}

func NewAudioFile(audioFile AudioFileStore) *AudioFiles {
	return &AudioFiles{
		audioFile: audioFile,
	}
}
