package mocks

import (
	"context"

	"github.com/RecoBattle/internal/app/qualitycontrolapp"
	"github.com/google/uuid"
	"github.com/stretchr/testify/mock"
)

type MockQualityControlStore struct {
	mock.Mock
}

func (m *MockQualityControlStore) Create(ctx context.Context, qualityControl qualitycontrolapp.IdealText) error {
	qualityControl.UUID = uuid.MustParse("2d53b244-8844-40a6-ab37-e5b89019af0a")
	args := m.Called(ctx, qualityControl)
	return args.Error(0)
}

func (m *MockQualityControlStore) GetTextASRIdeal(ctx context.Context, fileID string) ([]qualitycontrolapp.QualityControl, string, error) {
	args := m.Called(ctx, fileID)
	return args.Get(0).([]qualitycontrolapp.QualityControl), args.Get(0).(string), args.Error(2)
}
