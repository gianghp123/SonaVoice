package mocks

import (
	"context"

	"github.com/gianghp123/SonaVoice/api/internal/core/response"
	"github.com/gianghp123/SonaVoice/api/internal/database"
	"github.com/gianghp123/SonaVoice/api/internal/database/models"
	"github.com/stretchr/testify/mock"
)

type MessageRepository struct {
	mock.Mock
}

func (m *MessageRepository) Create(ctx context.Context, msg *models.Message) error {
	args := m.Called(ctx, msg)
	return args.Error(0)
}

func (m *MessageRepository) CreateBatch(ctx context.Context, msgs []*models.Message) error {
	args := m.Called(ctx, msgs)
	return args.Error(0)
}

func (m *MessageRepository) ListBySessionID(ctx context.Context, sessionID string, q *database.Query) (*response.PaginatedResult[*models.Message], error) {
	args := m.Called(ctx, sessionID, q)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*response.PaginatedResult[*models.Message]), args.Error(1)
}

func (m *MessageRepository) GetByID(ctx context.Context, id string) (*models.Message, error) {
	args := m.Called(ctx, id)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*models.Message), args.Error(1)
}
