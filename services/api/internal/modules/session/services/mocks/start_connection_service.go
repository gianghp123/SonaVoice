package mocks

import (
	"context"

	"github.com/gianghp123/SonaVoice/api/internal/core/errors"
	"github.com/gianghp123/SonaVoice/api/internal/database/models"
	"github.com/gianghp123/SonaVoice/api/internal/modules/session/dtos/req"
	"github.com/gianghp123/SonaVoice/api/internal/modules/session/dtos/res"
	"github.com/stretchr/testify/mock"
)

type StartConnectionService struct {
	mock.Mock
}

func (m *StartConnectionService) Start(ctx context.Context, session *models.Session, connReq *req.StartConnectionReq) (*res.WebRTCConnectionRes, *errors.AppError) {
	args := m.Called(ctx, session, connReq)
	if args.Get(0) == nil {
		return nil, func() *errors.AppError {
			if args.Get(1) == nil {
				return nil
			}
			return args.Get(1).(*errors.AppError)
		}()
	}
	return args.Get(0).(*res.WebRTCConnectionRes), func() *errors.AppError {
		if args.Get(1) == nil {
			return nil
		}
		return args.Get(1).(*errors.AppError)
	}()
}
