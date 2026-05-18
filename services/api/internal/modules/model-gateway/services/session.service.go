package services

import (
	"context"

	"github.com/gianghp123/SonaVoice/api/internal/core/enums"
	"github.com/gianghp123/SonaVoice/api/internal/core/errors"
	zapLogger "github.com/gianghp123/SonaVoice/api/internal/core/zap-logger"
	"github.com/gianghp123/SonaVoice/api/internal/database/models"
	repository_interfaces "github.com/gianghp123/SonaVoice/api/internal/database/repository-interfaces"
	"github.com/gianghp123/SonaVoice/api/internal/utils"
)

type ISessionService interface {
	Create(ctx context.Context, userID string) (*models.Session, *errors.AppError)
	Get(ctx context.Context, sessionID, requesterID string) (*models.Session, *errors.AppError)
	GetBySpeechSessionID(ctx context.Context, speechSessionID, requesterID string) (*models.Session, *errors.AppError)
	MarkSessionActive(ctx context.Context, sessionID string) *errors.AppError
	MarkSessionFailed(ctx context.Context, sessionID string) *errors.AppError
}

type sessionService struct {
	sessionRepo repository_interfaces.ISessionRepository
}

func NewSessionService(sessionRepo repository_interfaces.ISessionRepository) ISessionService {
	return &sessionService{sessionRepo: sessionRepo}
}

func (s *sessionService) Create(ctx context.Context, userID string) (*models.Session, *errors.AppError) {
	logger := zapLogger.S()
	session := &models.Session{
		UserID: userID,
		Status: enums.SessionStatusPending,
	}
	if err := s.sessionRepo.Create(ctx, session); err != nil {
		logger.Errorw("Failed to create session", "error", err)
		return nil, errors.MapRepoError(err)
	}
	return session, nil
}

func (s *sessionService) Get(ctx context.Context, sessionID, requesterID string) (*models.Session, *errors.AppError) {
	logger := zapLogger.S()
	session, err := s.sessionRepo.Get(ctx, sessionID)
	if err != nil {
		logger.Errorw("Failed to get session", "error", err)
		return nil, errors.MapRepoError(err)
	}

	if appErr := utils.EnforceOwnership(session.UserID, requesterID); appErr != nil {
		logger.Errorw("Failed to enforce ownership", "error", appErr)
		return nil, appErr
	}

	return session, nil
}

func (s *sessionService) GetBySpeechSessionID(ctx context.Context, speechSessionID, requesterID string) (*models.Session, *errors.AppError) {
	logger := zapLogger.S()
	session, err := s.sessionRepo.GetBySpeechSessionID(ctx, speechSessionID)
	if err != nil {
		logger.Errorw("Failed to get session by speech session id", "error", err)
		return nil, errors.MapRepoError(err)
	}

	if appErr := utils.EnforceOwnership(session.UserID, requesterID); appErr != nil {
		logger.Errorw("Failed to enforce ownership", "error", appErr)
		return nil, appErr
	}

	return session, nil
}

func (s *sessionService) MarkSessionActive(ctx context.Context, sessionID string) *errors.AppError {
	logger := zapLogger.S()
	if err := s.sessionRepo.SetSessionActive(ctx, sessionID, utils.NowUTC()); err != nil {
		logger.Errorw("Failed to mark session active", "error", err)
		return errors.Internal("failed to mark session active")
	}
	return nil
}

func (s *sessionService) MarkSessionFailed(ctx context.Context, sessionID string) *errors.AppError {
	logger := zapLogger.S()
	if err := s.sessionRepo.SetSessionFailed(ctx, sessionID); err != nil {
		logger.Errorw("Failed to mark session failed", "error", err)
		return errors.Internal("failed to mark session failed")
	}
	return nil
}
