package mocks

import (
	"context"

	"github.com/RecoBattle/internal/app/userapp"
	"github.com/google/uuid"
	"github.com/stretchr/testify/mock"
)

type MockUserStore struct {
	mock.Mock
}

func (m *MockUserStore) Create(ctx context.Context, user userapp.User) error {
	user.UUID = uuid.MustParse("2d53b244-8844-40a6-ab37-e5b89019af0a")

	args := m.Called(ctx, user)
	return args.Error(0)
}

func (m *MockUserStore) GetUser(ctx context.Context, condition map[string]string) (*userapp.User, error) {
	args := m.Called(ctx, condition)
	return args.Get(0).(*userapp.User), args.Error(1)
}
