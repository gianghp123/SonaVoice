package mocks

import (
	"context"

	"github.com/gianghp123/SonaVoice/api/internal/database/models"
	"github.com/stretchr/testify/mock"
)

type SessionRepository struct {
	mock.Mock
}

func (m *SessionRepository) Create(ctx context.Context, model *models.Session) error {
	args := m.Called(ctx, model)
	return args.Error(0)
}

func (m *SessionRepository) Get(ctx context.Context, sessionId string) (*models.Session, error) {
	args := m.Called(ctx, sessionId)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*models.Session), args.Error(1)
}

func (m *SessionRepository) Update(ctx context.Context, model *models.Session) error {
	args := m.Called(ctx, model)
	return args.Error(0)
}

func (m *SessionRepository) GetBySpeechSessionID(ctx context.Context, speechSessionId string) (*models.Session, error) {
	args := m.Called(ctx, speechSessionId)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*models.Session), args.Error(1)
}
