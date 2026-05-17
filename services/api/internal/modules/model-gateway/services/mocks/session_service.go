package mocks

import (
	"context"

	"github.com/gianghp123/SonaVoice/api/internal/core/enums"
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

func (m *SessionService) GetSessionInternal(ctx context.Context, sessionID string) (*res.SessionRes, *errors.AppError) {
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

func (m *SessionService) SetReservation(ctx context.Context, sessionID string, reservedAmount, dailyQuota int64) *errors.AppError {
	args := m.Called(ctx, sessionID, reservedAmount, dailyQuota)
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

func (m *SessionService) MarkQuotaReleased(ctx context.Context, sessionID string) *errors.AppError {
	args := m.Called(ctx, sessionID)
	if args.Get(0) == nil {
		return nil
	}
	return args.Get(0).(*errors.AppError)
}

func (m *SessionService) FindActiveByUserID(ctx context.Context, userID string) (*res.SessionRes, *errors.AppError) {
	args := m.Called(ctx, userID)
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

func (m *SessionService) FindResumableByUserID(ctx context.Context, userID string) ([]*res.SessionListItemRes, *errors.AppError) {
	args := m.Called(ctx, userID)
	if args.Get(0) == nil {
		return nil, func() *errors.AppError {
			if args.Get(1) == nil {
				return nil
			}
			return args.Get(1).(*errors.AppError)
		}()
	}
	return args.Get(0).([]*res.SessionListItemRes), func() *errors.AppError {
		if args.Get(1) == nil {
			return nil
		}
		return args.Get(1).(*errors.AppError)
	}()
}

func (m *SessionService) UpdateStatus(ctx context.Context, sessionID string, status enums.SessionStatus) *errors.AppError {
	args := m.Called(ctx, sessionID, status)
	if args.Get(0) == nil {
		return nil
	}
	return args.Get(0).(*errors.AppError)
}

func (m *SessionService) FindStaleSessions(ctx context.Context, userID string, pendingTimeoutSeconds int64) ([]*res.SessionRes, *errors.AppError) {
	args := m.Called(ctx, userID, pendingTimeoutSeconds)
	if args.Get(0) == nil {
		return nil, func() *errors.AppError {
			if args.Get(1) == nil {
				return nil
			}
			return args.Get(1).(*errors.AppError)
		}()
	}
	return args.Get(0).([]*res.SessionRes), func() *errors.AppError {
		if args.Get(1) == nil {
			return nil
		}
		return args.Get(1).(*errors.AppError)
	}()
}