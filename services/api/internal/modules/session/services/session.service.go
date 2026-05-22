package services

import (
	"context"

	"github.com/gianghp123/SonaVoice/api/internal/core/enums"
	"github.com/gianghp123/SonaVoice/api/internal/core/errors"
	"github.com/gianghp123/SonaVoice/api/internal/core/response"
	zapLogger "github.com/gianghp123/SonaVoice/api/internal/core/zap-logger"
	"github.com/gianghp123/SonaVoice/api/internal/database"
	"github.com/gianghp123/SonaVoice/api/internal/database/models"
	repository_interfaces "github.com/gianghp123/SonaVoice/api/internal/database/repository-interfaces"
	"github.com/gianghp123/SonaVoice/api/internal/modules/session/dtos/req"
	"github.com/gianghp123/SonaVoice/api/internal/utils"
)

type ISessionService interface {
	Create(ctx context.Context, userID string) (*models.Session, *errors.AppError)
	Get(ctx context.Context, sessionID string) (*models.Session, *errors.AppError)
	GetBySpeechSessionID(ctx context.Context, speechSessionID string) (*models.Session, *errors.AppError)
	List(ctx context.Context, q req.SessionListQuery) (*response.PaginatedResult[*models.Session], *errors.AppError)
	MarkSessionActive(ctx context.Context, sessionID string) *errors.AppError
	MarkSessionFailed(ctx context.Context, sessionID string) *errors.AppError
}

type sessionService struct {
	sessionRepo repository_interfaces.ISessionRepository
}

func NewSessionService(sessionRepo repository_interfaces.ISessionRepository) ISessionService {
	return &sessionService{sessionRepo: sessionRepo}
}

func (s *sessionService) List(ctx context.Context, q req.SessionListQuery) (*response.PaginatedResult[*models.Session], *errors.AppError) {
	logger := zapLogger.S()

	dbQuery := database.NewQuery().
		SetPage(q.Page).
		SetLimit(q.Limit).
		SetOrderBy("created_at DESC")

	if q.UserID != nil {
		dbQuery.SetFilter("user_id", *q.UserID)
	}

	if q.Status != nil {
		dbQuery.SetFilter("status", *q.Status)
	}

	result, err := s.sessionRepo.List(ctx, dbQuery)
	if err != nil {
		logger.Errorw("Failed to list sessions", "error", err)
		return nil, errors.MapRepoError(err)
	}

	return result, nil
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

func (s *sessionService) Get(ctx context.Context, sessionID string) (*models.Session, *errors.AppError) {
	logger := zapLogger.S()
	session, err := s.sessionRepo.Get(ctx, sessionID)
	if err != nil {
		logger.Errorw("Failed to get session", "error", err)
		return nil, errors.MapRepoError(err)
	}
	return session, nil
}

func (s *sessionService) GetBySpeechSessionID(ctx context.Context, speechSessionID string) (*models.Session, *errors.AppError) {
	logger := zapLogger.S()
	session, err := s.sessionRepo.GetBySpeechSessionID(ctx, speechSessionID)
	if err != nil {
		logger.Errorw("Failed to get session by speech session id", "error", err)
		return nil, errors.MapRepoError(err)
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
