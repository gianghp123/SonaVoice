package mocks

import (
	"context"
	"time"

	"github.com/gianghp123/SonaVoice/api/internal/core/errors"
	"github.com/stretchr/testify/mock"
)

type QuoteService struct {
	mock.Mock
}

func (m *QuoteService) ReserveAllRemaining(ctx context.Context, userID string, dailyQuota int64) (int64, error) {
	args := m.Called(ctx, userID, dailyQuota)
	return args.Get(0).(int64), args.Error(1)
}

func (m *QuoteService) Release(ctx context.Context, userID string, reservedAmount, actualUsage, dailyQuota int64) error {
	args := m.Called(ctx, userID, reservedAmount, actualUsage, dailyQuota)
	return args.Error(0)
}

func (m *QuoteService) AcquireSessionLock(ctx context.Context, userID string, ttl time.Duration) (string, error) {
	args := m.Called(ctx, userID, ttl)
	return args.String(0), args.Error(1)
}

func (m *QuoteService) ReleaseSessionLock(ctx context.Context, userID, lockValue string) error {
	args := m.Called(ctx, userID, lockValue)
	return args.Error(0)
}

func (m *QuoteService) ReserveQuota(ctx context.Context, userID string, dailyQuota int64) (int64, func(bool), *errors.AppError) {
	args := m.Called(ctx, userID, dailyQuota)
	cleanup, _ := args.Get(1).(func(bool))
	return args.Get(0).(int64), cleanup, func() *errors.AppError {
		if args.Get(2) == nil {
			return nil
		}
		return args.Get(2).(*errors.AppError)
	}()
}
