package mocks

import (
	"context"

	"github.com/gianghp123/SonaVoice/api/internal/core/errors"
	"github.com/gianghp123/SonaVoice/api/internal/modules/model-gateway/dtos/res"
	"github.com/stretchr/testify/mock"
)

type SessionService struct {
	mock.Mock
}

func (m *SessionService) CreateSession(ctx context.Context) (*res.SessionRes, *errors.AppError) {
	args := m.Called(ctx)
	if args.Get(0) == nil {
		return nil, func() *errors.AppError {
			if args.Get(1) == nil {
				return nil
			}
			return args.Get(1).(*errors.AppError)
		}()
	}
	return args.Get(0).(*res.SessionRes), func() *errors.AppError {
		if args.Get(1) == nil {
			return nil
		}
		return args.Get(1).(*errors.AppError)
	}()
}

func (m *SessionService) GetSession(ctx context.Context, sessionID string) (*res.SessionRes, *errors.AppError) {
	args := m.Called(ctx, sessionID)
	if args.Get(0) == nil {
		return nil, func() *errors.AppError {
			if args.Get(1) == nil {
				return nil
			}
			return args.Get(1).(*errors.AppError)
		}()
	}
	return args.Get(0).(*res.SessionRes), func() *errors.AppError {
		if args.Get(1) == nil {
			return nil
		}
		return args.Get(1).(*errors.AppError)
	}()
}

func (m *SessionService) GetSessionBySpeechSessionID(ctx context.Context, speechSessionID string) (*res.SessionRes, *errors.AppError) {
	args := m.Called(ctx, speechSessionID)
	if args.Get(0) == nil {
		return nil, func() *errors.AppError {
			if args.Get(1) == nil {
				return nil
			}
			return args.Get(1).(*errors.AppError)
		}()
	}
	return args.Get(0).(*res.SessionRes), func() *errors.AppError {
		if args.Get(1) == nil {
			return nil
		}
		return args.Get(1).(*errors.AppError)
	}()
}

func (m *SessionService) SetSpeechSessionID(ctx context.Context, sessionID, speechSessionID string) *errors.AppError {
	args := m.Called(ctx, sessionID, speechSessionID)
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

func (m *SessionService) MarkSessionActive(ctx context.Context, sessionID string) *errors.AppError {
	args := m.Called(ctx, sessionID)
	if args.Get(0) == nil {
		return nil
	}
	return args.Get(0).(*errors.AppError)
}

func (m *SessionService) MarkSessionInactive(ctx context.Context, sessionID string) *errors.AppError {
	args := m.Called(ctx, sessionID)
	if args.Get(0) == nil {
		return nil
	}
	return args.Get(0).(*errors.AppError)
}
