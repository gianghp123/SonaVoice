package mocks

import (
	"context"
	"github.com/stretchr/testify/mock"
)

type QuotaService struct {
	mock.Mock
}

func (m *QuotaService) CheckRemaining(ctx context.Context, userID string, dailyLimit int64) (int64, error) {
	args := m.Called(ctx, userID, dailyLimit)
	return args.Get(0).(int64), args.Error(1)
}
