package mocks

import (
	"context"
	"time"

	"github.com/gianghp123/SonaVoice/api/internal/core/errors"
	"github.com/stretchr/testify/mock"
)

type SessionLockService struct {
	mock.Mock
}

func (m *SessionLockService) Acquire(ctx context.Context, userID string, ttl time.Duration) (string, *errors.AppError) {
	args := m.Called(ctx, userID, ttl)
	lockValue, _ := args.Get(0).(string)
	if args.Get(1) == nil {
		return lockValue, nil
	}
	return lockValue, args.Get(1).(*errors.AppError)
}

func (m *SessionLockService) Release(ctx context.Context, userID, lockValue string) *errors.AppError {
	args := m.Called(ctx, userID, lockValue)
	if args.Get(0) == nil {
		return nil
	}
	return args.Get(0).(*errors.AppError)
}
