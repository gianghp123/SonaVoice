package mocks

import (
	"context"
	"time"

	"github.com/stretchr/testify/mock"
)

type UserQuotaRepository struct {
	mock.Mock
}

func (m *UserQuotaRepository) ReserveAll(ctx context.Context, userID string, quotaKey string, quotaDate time.Time, dailyLimit int64) (int64, error) {
	args := m.Called(ctx, userID, quotaKey, quotaDate, dailyLimit)
	return args.Get(0).(int64), args.Error(1)
}

func (m *UserQuotaRepository) Release(ctx context.Context, userID string, quotaKey string, quotaDate time.Time, amount int64) error {
	args := m.Called(ctx, userID, quotaKey, quotaDate, amount)
	return args.Error(0)
}
