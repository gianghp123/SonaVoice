package mocks

import (
	"context"

	"github.com/gianghp123/SonaVoice/api/internal/database/models"
	"github.com/stretchr/testify/mock"
)

type SessionConfigRepository struct {
	mock.Mock
}

func (m *SessionConfigRepository) Get(ctx context.Context) (*models.SessionConfig, error) {
	args := m.Called(ctx)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*models.SessionConfig), args.Error(1)
}

func (m *SessionConfigRepository) Save(ctx context.Context, model *models.SessionConfig) error {
	args := m.Called(ctx, model)
	return args.Error(0)
}
