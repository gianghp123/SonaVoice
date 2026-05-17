package mocks

import (
	"context"

	"github.com/gianghp123/SonaVoice/api/internal/database/models"
	"github.com/stretchr/testify/mock"
)

type GlobalConfigRepository struct {
	mock.Mock
}

func (m *GlobalConfigRepository) Get(ctx context.Context) (*models.GlobalConfig, error) {
	args := m.Called(ctx)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*models.GlobalConfig), args.Error(1)
}

func (m *GlobalConfigRepository) Save(ctx context.Context, model *models.GlobalConfig) error {
	args := m.Called(ctx, model)
	return args.Error(0)
}
