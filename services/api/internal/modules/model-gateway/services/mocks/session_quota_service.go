package mocks

import (
	"context"

	"github.com/gianghp123/SonaVoice/api/internal/core/errors"
	"github.com/gianghp123/SonaVoice/api/internal/database/models"
	"github.com/stretchr/testify/mock"
)

type SessionQuotaService struct {
	mock.Mock
}

func (m *SessionQuotaService) Reserve(ctx context.Context, userID string, dailyQuota int) (int64, *errors.AppError) {
	args := m.Called(ctx, userID, dailyQuota)
	if args.Get(1) == nil {
		return args.Get(0).(int64), nil
	}
	return args.Get(0).(int64), args.Get(1).(*errors.AppError)
}

func (m *SessionQuotaService) ReleaseAll(ctx context.Context, userID string, amount int64) *errors.AppError {
	args := m.Called(ctx, userID, amount)
	if args.Get(0) == nil {
		return nil
	}
	return args.Get(0).(*errors.AppError)
}

func (m *SessionQuotaService) ReleaseWithActualUsage(ctx context.Context, session *models.Session, actualUsage int64) *errors.AppError {
	args := m.Called(ctx, session, actualUsage)
	if args.Get(0) == nil {
		return nil
	}
	return args.Get(0).(*errors.AppError)
}
