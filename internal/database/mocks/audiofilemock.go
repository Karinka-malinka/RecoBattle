package mocks

import (
	"context"

	"github.com/RecoBattle/internal/app/audiofilesapp"
	"github.com/google/uuid"
	"github.com/stretchr/testify/mock"
)

type MockAudioFileStore struct {
	mock.Mock
}

func (m *MockAudioFileStore) CreateFile(ctx context.Context, audioFile audiofilesapp.AudioFile) error {
	audioFile.UUID = uuid.MustParse("2d53b244-8844-40a6-ab37-e5b89019af0a")
	args := m.Called(ctx, audioFile)
	return args.Error(0)
}

func (m *MockAudioFileStore) CreateASR(ctx context.Context, audioFile audiofilesapp.AudioFile) error {
	audioFile.UUID = uuid.MustParse("2d53b244-8844-40a6-ab37-e5b89019af0a")
	args := m.Called(ctx, audioFile)
	return args.Error(0)
}

func (m *MockAudioFileStore) UpdateStatusASR(ctx context.Context, audioFileUUID, status string) error {
	args := m.Called(ctx, audioFileUUID, status)
	return args.Error(0)
}

func (m *MockAudioFileStore) CreateResultASR(ctx context.Context, resultASR audiofilesapp.ResultASR) error {
	args := m.Called(ctx, resultASR)
	return args.Error(0)
}

func (m *MockAudioFileStore) GetAudioFiles(ctx context.Context, userID string) (*[]audiofilesapp.AudioFile, error) {
	args := m.Called(ctx, userID)
	return args.Get(0).(*[]audiofilesapp.AudioFile), args.Error(1)
}

func (m *MockAudioFileStore) GetResultASR(ctx context.Context, uuid string) (*[]audiofilesapp.ResultASR, error) {
	args := m.Called(ctx, uuid)
	return args.Get(0).(*[]audiofilesapp.ResultASR), args.Error(1)
}
