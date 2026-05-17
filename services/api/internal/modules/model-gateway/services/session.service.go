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
	GetSessionInternal(ctx context.Context, sessionID string) (*res.SessionRes, *errors.AppError)
	GetSessionBySpeechSessionID(ctx context.Context, speechSessionID string) (*res.SessionRes, *errors.AppError)
	SetSpeechSessionID(ctx context.Context, sessionID, speechSessionID string) *errors.AppError
	SetReservation(ctx context.Context, sessionID string, reservedAmount, dailyQuota int64) *errors.AppError
	MarkSessionFailed(ctx context.Context, sessionID string) *errors.AppError
	MarkSessionActive(ctx context.Context, sessionID string) *errors.AppError
	MarkSessionInactive(ctx context.Context, sessionID string) *errors.AppError
	MarkQuotaReleased(ctx context.Context, sessionID string) *errors.AppError
	FindStaleSessions(ctx context.Context, userID string, pendingTimeoutSeconds int64) ([]*res.SessionRes, *errors.AppError)
	FindActiveByUserID(ctx context.Context, userID string) (*res.SessionRes, *errors.AppError)
	FindResumableByUserID(ctx context.Context, userID string) ([]*res.SessionListItemRes, *errors.AppError)
	UpdateStatus(ctx context.Context, sessionID string, status enums.SessionStatus) *errors.AppError
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

func (s *sessionService) GetSessionInternal(ctx context.Context, sessionID string) (*res.SessionRes, *errors.AppError) {
	logger := zapLogger.S()
	session, err := s.sessionRepo.Get(ctx, sessionID)
	if err != nil {
		logger.Errorw("Failed to get session", "error", err)
		return nil, errors.Internal()
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
	if err := s.sessionRepo.UpdateSpeechSessionID(ctx, sessionID, speechSessionID); err != nil {
		logger.Errorw("Failed to update speechSessionId", "error", err)
		return errors.Internal()
	}
	return nil
}

func (s *sessionService) SetReservation(ctx context.Context, sessionID string, reservedAmount, dailyQuota int64) *errors.AppError {
	logger := zapLogger.S()
	if err := s.sessionRepo.UpdateReservation(ctx, sessionID, reservedAmount, dailyQuota); err != nil {
		logger.Errorw("Failed to update session reservation", "error", err)
		return errors.Internal()
	}
	return nil
}

func (s *sessionService) MarkSessionFailed(ctx context.Context, sessionID string) *errors.AppError {
	logger := zapLogger.S()
	logger.Debugw("Marking session as failed", "sessionId", sessionID)
	if err := s.sessionRepo.UpdateStatus(ctx, sessionID, enums.SessionStatusFailed); err != nil {
		logger.Errorw("Failed to update session to failed", "error", err)
		return errors.Internal()
	}
	return nil
}

func (s *sessionService) MarkSessionActive(ctx context.Context, sessionID string) *errors.AppError {
	logger := zapLogger.S()
	if err := s.sessionRepo.UpdateActiveSession(ctx, sessionID, time.Now()); err != nil {
		logger.Errorw("Failed to update session to active", "error", err)
		return errors.Internal("failed to update session to active")
	}
	return nil
}

func (s *sessionService) MarkSessionInactive(ctx context.Context, sessionID string) *errors.AppError {
	logger := zapLogger.S()
	if err := s.sessionRepo.UpdateStatus(ctx, sessionID, enums.SessionStatusInactive); err != nil {
		logger.Errorw("Failed to update session", "error", err)
		return errors.Internal()
	}
	return nil
}

func (s *sessionService) MarkQuotaReleased(ctx context.Context, sessionID string) *errors.AppError {
	logger := zapLogger.S()
	if err := s.sessionRepo.UpdateQuotaReleased(ctx, sessionID); err != nil {
		logger.Errorw("Failed to mark quota released", "error", err)
		return errors.Internal()
	}
	return nil
}

func (s *sessionService) FindStaleSessions(ctx context.Context, userID string, pendingTimeoutSeconds int64) ([]*res.SessionRes, *errors.AppError) {
	logger := zapLogger.S()
	logger.Debugw("Finding up stale sessions", "userId", userID, "pendingTimeoutSeconds", pendingTimeoutSeconds)
	sessions, err := s.sessionRepo.FindStaleByUserID(ctx, userID, pendingTimeoutSeconds)
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
	logger.Debugw("Found up stale sessions", "count", len(results))
	return results, nil
}

func (s *sessionService) FindActiveByUserID(ctx context.Context, userID string) (*res.SessionRes, *errors.AppError) {
	logger := zapLogger.S()
	session, err := s.sessionRepo.FindActiveByUserID(ctx, userID)
	if err != nil {
		logger.Errorw("Failed to find active session by user ID", "userId", userID, "error", err)
		return nil, errors.Internal()
	}
	if session == nil {
		return nil, nil
	}
	var dto res.SessionRes
	if err := utils.MapToDTO(session, &dto); err != nil {
		logger.Errorw("Failed to map session to dto", "error", err)
		return nil, errors.Internal()
	}
	return &dto, nil
}

func (s *sessionService) FindResumableByUserID(ctx context.Context, userID string) ([]*res.SessionListItemRes, *errors.AppError) {
	logger := zapLogger.S()
	sessions, err := s.sessionRepo.FindResumableByUserID(ctx, userID)
	if err != nil {
		logger.Errorw("Failed to find resumable sessions", "userId", userID, "error", err)
		return nil, errors.Internal()
	}
	var results []*res.SessionListItemRes
	for _, session := range sessions {
		var dto res.SessionListItemRes
		if err := utils.MapToDTO(session, &dto); err != nil {
			logger.Errorw("Failed to map session to dto", "sessionId", session.ID, "error", err)
			continue
		}
		results = append(results, &dto)
	}
	return results, nil
}

func (s *sessionService) UpdateStatus(ctx context.Context, sessionID string, status enums.SessionStatus) *errors.AppError {
	logger := zapLogger.S()
	if err := s.sessionRepo.UpdateStatus(ctx, sessionID, status); err != nil {
		logger.Errorw("Failed to update session status", "sessionId", sessionID, "status", status, "error", err)
		return errors.Internal()
	}
	return nil
}
