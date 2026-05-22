package mocks

import (
	"context"

	"github.com/gianghp123/SonaVoice/api/internal/core/errors"
	"github.com/gianghp123/SonaVoice/api/internal/core/response"
	"github.com/gianghp123/SonaVoice/api/internal/database/models"
	"github.com/gianghp123/SonaVoice/api/internal/modules/session/dtos/req"
	"github.com/stretchr/testify/mock"
)

type SessionService struct {
	mock.Mock
}

func (m *SessionService) Create(ctx context.Context, userID string) (*models.Session, *errors.AppError) {
	args := m.Called(ctx, userID)
	if args.Get(0) == nil {
		return nil, func() *errors.AppError {
			if args.Get(1) == nil {
				return nil
			}
			return args.Get(1).(*errors.AppError)
		}()
	}
	return args.Get(0).(*models.Session), func() *errors.AppError {
		if args.Get(1) == nil {
			return nil
		}
		return args.Get(1).(*errors.AppError)
	}()
}

func (m *SessionService) Get(ctx context.Context, sessionID string) (*models.Session, *errors.AppError) {
	args := m.Called(ctx, sessionID)
	if args.Get(0) == nil {
		return nil, func() *errors.AppError {
			if args.Get(1) == nil {
				return nil
			}
			return args.Get(1).(*errors.AppError)
		}()
	}
	return args.Get(0).(*models.Session), func() *errors.AppError {
		if args.Get(1) == nil {
			return nil
		}
		return args.Get(1).(*errors.AppError)
	}()
}

func (m *SessionService) GetBySpeechSessionID(ctx context.Context, speechSessionID string) (*models.Session, *errors.AppError) {
	args := m.Called(ctx, speechSessionID)
	if args.Get(0) == nil {
		return nil, func() *errors.AppError {
			if args.Get(1) == nil {
				return nil
			}
			return args.Get(1).(*errors.AppError)
		}()
	}
	return args.Get(0).(*models.Session), func() *errors.AppError {
		if args.Get(1) == nil {
			return nil
		}
		return args.Get(1).(*errors.AppError)
	}()
}

func (m *SessionService) List(ctx context.Context, q req.SessionListQuery) (*response.PaginatedResult[*models.Session], *errors.AppError) {
	args := m.Called(ctx, q)
	if args.Get(0) == nil {
		return nil, func() *errors.AppError {
			if args.Get(1) == nil {
				return nil
			}
			return args.Get(1).(*errors.AppError)
		}()
	}
	return args.Get(0).(*response.PaginatedResult[*models.Session]), func() *errors.AppError {
		if args.Get(1) == nil {
			return nil
		}
		return args.Get(1).(*errors.AppError)
	}()
}

func (m *SessionService) MarkSessionActive(ctx context.Context, sessionID string) *errors.AppError {
	args := m.Called(ctx, sessionID)
	if args.Get(0) == nil {
		return nil
	}
	return args.Get(0).(*errors.AppError)
}

func (m *SessionService) MarkSessionFailed(ctx context.Context, sessionID string) *errors.AppError {
	args := m.Called(ctx, sessionID)
	if args.Get(0) == nil {
		return nil
	}
	return args.Get(0).(*errors.AppError)
}
