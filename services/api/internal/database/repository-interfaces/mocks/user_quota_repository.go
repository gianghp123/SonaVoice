package mocks

import (
	"context"
	"time"

	"github.com/stretchr/testify/mock"
)

type UserQuotaRepository struct {
	mock.Mock
}

func (m *UserQuotaRepository) GetOrCreate(ctx context.Context, userID string, quotaKey string, quotaDate time.Time, initialAmount int64) (int64, error) {
	args := m.Called(ctx, userID, quotaKey, quotaDate, initialAmount)
	return args.Get(0).(int64), args.Error(1)
}

func (m *UserQuotaRepository) Deduct(ctx context.Context, userID string, quotaKey string, quotaDate time.Time, amount int64) error {
	args := m.Called(ctx, userID, quotaKey, quotaDate, amount)
	return args.Error(0)
}

func (m *UserQuotaRepository) GetRemaining(ctx context.Context, userID string, quotaKey string, quotaDate time.Time) (int64, error) {
	args := m.Called(ctx, userID, quotaKey, quotaDate)
	return args.Get(0).(int64), args.Error(1)
}
