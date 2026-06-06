package mocks

import (
	"context"

	"github.com/gianghp123/SonaVoice/api/internal/database/models"
	"github.com/stretchr/testify/mock"
)

type UserProfileRepository struct {
	mock.Mock
}

func (m *UserProfileRepository) GetByUserID(ctx context.Context, userID string) (*models.UserProfile, error) {
	args := m.Called(ctx, userID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*models.UserProfile), args.Error(1)
}

func (m *UserProfileRepository) Upsert(ctx context.Context, profile *models.UserProfile) error {
	args := m.Called(ctx, profile)
	return args.Error(0)
}
