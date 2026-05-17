package services

import (
	"context"
	"time"

	"github.com/gianghp123/SonaVoice/api/internal/core/enums"
	"github.com/gianghp123/SonaVoice/api/internal/core/errors"
	zapLogger "github.com/gianghp123/SonaVoice/api/internal/core/zap-logger"
	"github.com/gianghp123/SonaVoice/api/internal/database/models"
	repository_interfaces "github.com/gianghp123/SonaVoice/api/internal/database/repository-interfaces"
	"github.com/gianghp123/SonaVoice/api/internal/modules/model-gateway/dtos/res"
	"github.com/gianghp123/SonaVoice/api/internal/utils"
)

type ISessionService interface {
	CreateSession(ctx context.Context) (*res.SessionRes, *errors.AppError)
	GetSession(ctx context.Context, sessionID string) (*res.SessionRes, *errors.AppError)
	GetSessionBySpeechSessionID(ctx context.Context, speechSessionID string) (*res.SessionRes, *errors.AppError)
	SetSpeechSessionID(ctx context.Context, sessionID, speechSessionID string) *errors.AppError
	SetReservation(ctx context.Context, sessionID string, reservedAmount, dailyQuota int64) *errors.AppError
	MarkSessionFailed(ctx context.Context, sessionID string) *errors.AppError
	MarkSessionActive(ctx context.Context, sessionID string) *errors.AppError
	MarkSessionInactive(ctx context.Context, sessionID string) *errors.AppError
	MarkQuotaReleased(ctx context.Context, sessionID string) *errors.AppError
	CleanupStaleSessions(ctx context.Context, userID string) ([]*res.SessionRes, *errors.AppError)
}

type sessionService struct {
	sessionRepo repository_interfaces.ISessionRepository
}

func NewSessionService(sessionRepo repository_interfaces.ISessionRepository) ISessionService {
	return &sessionService{sessionRepo: sessionRepo}
}

func (s *sessionService) CreateSession(ctx context.Context) (*res.SessionRes, *errors.AppError) {
	logger := zapLogger.S()
	requesterId := utils.GetCtx[string](ctx, enums.ContextKeyUserID)
	session := &models.Session{
		UserID: requesterId,
		Status: enums.SessionStatusPending,
	}
	if err := s.sessionRepo.Create(ctx, session); err != nil {
		logger.Errorw("Failed to create session", "error", err)
		return nil, errors.Internal()
	}
	var dto res.SessionRes
	if err := utils.MapToDTO(session, &dto); err != nil {
		logger.Errorw("Failed to map session to dto", "error", err)
		return nil, errors.Internal()
	}
	return &dto, nil
}

func (s *sessionService) GetSession(ctx context.Context, sessionID string) (*res.SessionRes, *errors.AppError) {
	logger := zapLogger.S()
	session, err := s.sessionRepo.Get(ctx, sessionID)
	if err != nil {
		logger.Errorw("Failed to get session", "error", err)
		return nil, errors.Internal()
	}

	requesterId := utils.GetCtx[string](ctx, enums.ContextKeyUserID)

	if appErr := utils.EnforceOwnership(session.UserID, requesterId); appErr != nil {
		logger.Errorw("Failed to enforce ownership", "error", appErr)
		return nil, appErr
	}

	var dto res.SessionRes
	if err := utils.MapToDTO(session, &dto); err != nil {
		logger.Errorw("Failed to map session to dto", "error", err)
		return nil, errors.Internal()
	}
	return &dto, nil
}

func (s *sessionService) GetSessionBySpeechSessionID(ctx context.Context, speechSessionID string) (*res.SessionRes, *errors.AppError) {
	logger := zapLogger.S()
	session, err := s.sessionRepo.GetBySpeechSessionID(ctx, speechSessionID)
	if err != nil {
		logger.Errorw("Failed to get session by speech session id", "error", err)
		return nil, errors.Internal()
	}

	requesterId := utils.GetCtx[string](ctx, enums.ContextKeyUserID)
	if appErr := utils.EnforceOwnership(session.UserID, requesterId); appErr != nil {
		logger.Errorw("Failed to enforce ownership", "error", appErr)
		return nil, appErr
	}

	var dto res.SessionRes
	if err := utils.MapToDTO(session, &dto); err != nil {
		logger.Errorw("Failed to map session to dto", "error", err)
		return nil, errors.Internal()
	}
	return &dto, nil
}

func (s *sessionService) SetSpeechSessionID(ctx context.Context, sessionID, speechSessionID string) *errors.AppError {
	logger := zapLogger.S()
	session, err := s.sessionRepo.Get(ctx, sessionID)
	if err != nil {
		logger.Errorw("Failed to get session", "error", err)
		return errors.Internal()
	}
	session.SpeechSessionID = speechSessionID
	if err := s.sessionRepo.Update(ctx, session); err != nil {
		logger.Errorw("Failed to update speechSessionId", "error", err)
		return errors.Internal()
	}
	return nil
}

func (s *sessionService) SetReservation(ctx context.Context, sessionID string, reservedAmount, dailyQuota int64) *errors.AppError {
	logger := zapLogger.S()
	session, err := s.sessionRepo.Get(ctx, sessionID)
	if err != nil {
		logger.Errorw("Failed to get session", "error", err)
		return errors.Internal()
	}
	session.ReservedAmount = reservedAmount
	session.DailyQuota = dailyQuota
	if err := s.sessionRepo.Update(ctx, session); err != nil {
		logger.Errorw("Failed to update session reservation", "error", err)
		return errors.Internal()
	}
	return nil
}

func (s *sessionService) MarkSessionFailed(ctx context.Context, sessionID string) *errors.AppError {
	logger := zapLogger.S()
	session, err := s.sessionRepo.Get(ctx, sessionID)
	if err != nil {
		logger.Errorw("Failed to get session", "error", err)
		return errors.Internal()
	}
	session.Status = enums.SessionStatusFailed
	if err := s.sessionRepo.Update(ctx, session); err != nil {
		logger.Errorw("Failed to update session to failed", "error", err)
		return errors.Internal()
	}
	return nil
}

func (s *sessionService) MarkSessionActive(ctx context.Context, sessionID string) *errors.AppError {
	logger := zapLogger.S()
	session, err := s.sessionRepo.Get(ctx, sessionID)
	if err != nil {
		logger.Errorw("Failed to get session", "error", err)
		return errors.Internal()
	}
	session.StartedAt = time.Now()
	session.Status = enums.SessionStatusActive
	if err := s.sessionRepo.Update(ctx, session); err != nil {
		logger.Errorw("Failed to update session to active", "error", err)
		return errors.Internal("failed to update session to active")
	}
	return nil
}

func (s *sessionService) MarkSessionInactive(ctx context.Context, sessionID string) *errors.AppError {
	logger := zapLogger.S()
	session, err := s.sessionRepo.Get(ctx, sessionID)
	if err != nil {
		logger.Errorw("Failed to get session", "error", err)
		return errors.Internal()
	}
	session.Status = enums.SessionStatusInactive
	if err := s.sessionRepo.Update(ctx, session); err != nil {
		logger.Errorw("Failed to update session", "error", err)
		return errors.Internal()
	}
	return nil
}

func (s *sessionService) MarkQuotaReleased(ctx context.Context, sessionID string) *errors.AppError {
	logger := zapLogger.S()
	session, err := s.sessionRepo.Get(ctx, sessionID)
	if err != nil {
		logger.Errorw("Failed to get session", "error", err)
		return errors.Internal()
	}
	session.QuotaReleased = true
	if err := s.sessionRepo.Update(ctx, session); err != nil {
		logger.Errorw("Failed to mark quota released", "error", err)
		return errors.Internal()
	}
	return nil
}

func (s *sessionService) CleanupStaleSessions(ctx context.Context, userID string) ([]*res.SessionRes, *errors.AppError) {
	logger := zapLogger.S()
	sessions, err := s.sessionRepo.FindStaleByUserID(ctx, userID)
	if err != nil {
		logger.Errorw("Failed to find stale sessions", "error", err)
		return nil, errors.Internal()
	}

	var results []*res.SessionRes
	for _, session := range sessions {
		var dto res.SessionRes
		if err := utils.MapToDTO(session, &dto); err != nil {
			logger.Errorw("Failed to map session to dto", "sessionId", session.ID, "error", err)
			continue
		}
		results = append(results, &dto)
	}
	return results, nil
}