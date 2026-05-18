package mocks

import (
	"context"

	"github.com/gianghp123/SonaVoice/api/internal/core/errors"
	"github.com/gianghp123/SonaVoice/api/internal/database/models"
	"github.com/gianghp123/SonaVoice/api/internal/modules/model-gateway/dtos/res"
	"github.com/stretchr/testify/mock"
)

type StartConnectionService struct {
	mock.Mock
}

func (m *StartConnectionService) Start(ctx context.Context, session *models.Session, userID string, dailyQuota int) (*res.CreateSessionRes, *errors.AppError) {
	args := m.Called(ctx, session, userID, dailyQuota)
	if args.Get(0) == nil {
		return nil, func() *errors.AppError {
			if args.Get(1) == nil {
				return nil
			}
			return args.Get(1).(*errors.AppError)
		}()
	}
	return args.Get(0).(*res.CreateSessionRes), func() *errors.AppError {
		if args.Get(1) == nil {
			return nil
		}
		return args.Get(1).(*errors.AppError)
	}()
}
