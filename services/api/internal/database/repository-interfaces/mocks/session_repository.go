package mocks

import (
	"context"
	"time"

	"github.com/gianghp123/SonaVoice/api/internal/core/enums"
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

func (m *SessionRepository) GetForUpdate(ctx context.Context, sessionID string) (*models.Session, error) {
	args := m.Called(ctx, sessionID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*models.Session), args.Error(1)
}

func (m *SessionRepository) GetBySpeechSessionID(ctx context.Context, speechSessionId string) (*models.Session, error) {
	args := m.Called(ctx, speechSessionId)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*models.Session), args.Error(1)
}

func (m *SessionRepository) UpdateSpeechSessionID(ctx context.Context, sessionID, speechSessionID string) error {
	args := m.Called(ctx, sessionID, speechSessionID)
	return args.Error(0)
}

func (m *SessionRepository) UpdateStatus(ctx context.Context, sessionID string, status enums.SessionStatus) error {
	args := m.Called(ctx, sessionID, status)
	return args.Error(0)
}

func (m *SessionRepository) SetSessionActive(ctx context.Context, sessionID string, startedAt time.Time) error {
	args := m.Called(ctx, sessionID, startedAt)
	return args.Error(0)
}

func (m *SessionRepository) SetQuotaDate(ctx context.Context, sessionID string, quotaDate time.Time) error {
	args := m.Called(ctx, sessionID, quotaDate)
	return args.Error(0)
}
