package mocks

import (
	"context"

	"github.com/gianghp123/SonaVoice/api/internal/core/quota"
	"github.com/stretchr/testify/mock"
)

type QuotaService struct {
	mock.Mock
}

func (m *QuotaService) ReserveAll(ctx context.Context, userID string, cfg quota.QuotaConfig) (int64, error) {
	args := m.Called(ctx, userID, cfg)
	return args.Get(0).(int64), args.Error(1)
}

func (m *QuotaService) Release(ctx context.Context, userID string, cfg quota.QuotaConfig, reserved, actual int64) error {
	args := m.Called(ctx, userID, cfg, reserved, actual)
	return args.Error(0)
}
