package mocks

import (
	"context"

	"github.com/gianghp123/SonaVoice/api/internal/database/transaction"
	"github.com/stretchr/testify/mock"
)

type SessionJanitorService struct {
	mock.Mock
}

func (m *SessionJanitorService) CleanupStaleSessions(ctx context.Context, p transaction.IProvider, userID string, maxSessionLockTTL int64, dailyVoiceSeconds int64) error {
	args := m.Called(ctx, p, userID, maxSessionLockTTL, dailyVoiceSeconds)
	return args.Error(0)
}
