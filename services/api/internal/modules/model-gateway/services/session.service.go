package services

import (
	"context"
	"time"

	"github.com/gianghp123/SonaVoice/api/internal/core/enums"
	"github.com/gianghp123/SonaVoice/api/internal/core/errors"
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
	MarkSessionFailed(ctx context.Context, sessionID string) *errors.AppError
	MarkSessionActive(ctx context.Context, sessionID string) *errors.AppError
	MarkSessionInactive(ctx context.Context, sessionID string) *errors.AppError
}

type sessionService struct {
	sessionRepo repository_interfaces.ISessionRepository
}

func NewSessionService(sessionRepo repository_interfaces.ISessionRepository) ISessionService {
	return &sessionService{sessionRepo: sessionRepo}
}

func (s *sessionService) CreateSession(ctx context.Context) (*res.SessionRes, *errors.AppError) {
	requesterId := utils.GetCtx[string](ctx, enums.ContextKeyUserID)
	session := &models.Session{
		UserID: requesterId,
		Status: enums.SessionStatusPending,
	}
	if err := s.sessionRepo.Create(ctx, session); err != nil {
		return nil, errors.MapRepoError(err)
	}
	var dto res.SessionRes
	if err := utils.MapToDTO(session, &dto); err != nil {
		return nil, errors.Internal("failed to map session to dto")
	}
	return &dto, nil
}

func (s *sessionService) GetSession(ctx context.Context, sessionID string) (*res.SessionRes, *errors.AppError) {
	session, err := s.sessionRepo.Get(ctx, sessionID)
	if err != nil {
		return nil, errors.Internal("failed to get session")
	}
	var dto res.SessionRes
	if err := utils.MapToDTO(session, &dto); err != nil {
		return nil, errors.Internal("failed to map session to dto")
	}
	return &dto, nil
}

func (s *sessionService) GetSessionBySpeechSessionID(ctx context.Context, speechSessionID string) (*res.SessionRes, *errors.AppError) {
	session, err := s.sessionRepo.GetBySpeechSessionID(ctx, speechSessionID)
	if err != nil {
		return nil, errors.Internal("failed to get session")
	}
	var dto res.SessionRes
	if err := utils.MapToDTO(session, &dto); err != nil {
		return nil, errors.Internal("failed to map session to dto")
	}
	return &dto, nil
}

func (s *sessionService) SetSpeechSessionID(ctx context.Context, sessionID, speechSessionID string) *errors.AppError {
	session, err := s.sessionRepo.Get(ctx, sessionID)
	if err != nil {
		return errors.Internal("failed to get session")
	}
	session.SpeechSessionID = speechSessionID
	if err := s.sessionRepo.Update(ctx, session); err != nil {
		return errors.Internal("failed to update speechSessionId")
	}
	return nil
}

func (s *sessionService) MarkSessionFailed(ctx context.Context, sessionID string) *errors.AppError {
	session, err := s.sessionRepo.Get(ctx, sessionID)
	if err != nil {
		return errors.Internal("failed to get session")
	}
	session.Status = enums.SessionStatusFailed
	if err := s.sessionRepo.Update(ctx, session); err != nil {
		return errors.Internal("failed to update session to failed")
	}
	return nil
}

func (s *sessionService) MarkSessionActive(ctx context.Context, sessionID string) *errors.AppError {
	session, err := s.sessionRepo.Get(ctx, sessionID)
	if err != nil {
		return errors.Internal("failed to get session")
	}
	session.StartedAt = time.Now()
	session.Status = enums.SessionStatusActive
	if err := s.sessionRepo.Update(ctx, session); err != nil {
		return errors.Internal("failed to update session to active")
	}
	return nil
}

func (s *sessionService) MarkSessionInactive(ctx context.Context, sessionID string) *errors.AppError {
	session, err := s.sessionRepo.Get(ctx, sessionID)
	if err != nil {
		return errors.Internal("failed to get session")
	}
	session.Status = enums.SessionStatusInactive
	if err := s.sessionRepo.Update(ctx, session); err != nil {
		return errors.Internal("failed to update session")
	}
	return nil
}
