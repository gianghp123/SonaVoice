package services

import (
	"context"

	"github.com/gianghp123/SonaVoice/api/internal/core/errors"
	zapLogger "github.com/gianghp123/SonaVoice/api/internal/core/zap-logger"
	"github.com/gianghp123/SonaVoice/api/internal/database/models"
	"github.com/gianghp123/SonaVoice/api/internal/modules/model-gateway/dtos/req"
	"github.com/gianghp123/SonaVoice/api/internal/modules/model-gateway/dtos/res"
)

type ISessionStarterService interface {
	StartOrResume(ctx context.Context, session *models.Session, requesterID string, dailyQuota int) (*res.CreateSessionRes, *errors.AppError)
}

type sessionStarterService struct {
	quotaSvc   ISessionQuotaService
	sessionSvc ISessionService
	speechSvc  ISpeechProxyService
}

func NewSessionStarterService(quotaSvc ISessionQuotaService, sessionSvc ISessionService, speechSvc ISpeechProxyService) ISessionStarterService {
	return &sessionStarterService{
		quotaSvc:   quotaSvc,
		sessionSvc: sessionSvc,
		speechSvc:  speechSvc,
	}
}

func (s *sessionStarterService) StartOrResume(ctx context.Context, session *models.Session, requesterID string, dailyQuota int) (*res.CreateSessionRes, *errors.AppError) {
	logger := zapLogger.S()

	reservedAmount, appErr := s.quotaSvc.Reserve(ctx, requesterID, dailyQuota)
	if appErr != nil {
		return nil, appErr
	}

	if appErr := s.sessionSvc.SetReservation(ctx, session.ID, reservedAmount, int64(dailyQuota)); appErr != nil {
		_ = s.quotaSvc.ReleaseAll(ctx, requesterID, reservedAmount)
		return nil, appErr
	}

	connReq := &req.StartConnectionReq{
		EnableDefaultIceServers: true,
		Body: req.StartConnectionBody{
			UserID:      requesterID,
			SessionID:   session.ID,
			MaxDuration: reservedAmount,
		},
	}

	webrtcRes, appErr := s.speechSvc.StartConnection(ctx, connReq)
	if appErr != nil {
		logger.Errorw("Failed to connect to speech engine", "sessionId", session.ID, "error", appErr)
		_ = s.quotaSvc.ReleaseAll(ctx, requesterID, reservedAmount)
		_ = s.sessionSvc.MarkSessionFailed(ctx, session.ID)
		_ = s.sessionSvc.MarkQuotaReleased(ctx, session.ID)
		return nil, appErr
	}

	if appErr := s.sessionSvc.SetSpeechSessionID(ctx, session.ID, webrtcRes.SessionID); appErr != nil {
		_ = s.quotaSvc.ReleaseAll(ctx, requesterID, reservedAmount)
		_ = s.sessionSvc.MarkSessionFailed(ctx, session.ID)
		_ = s.sessionSvc.MarkQuotaReleased(ctx, session.ID)
		return nil, appErr
	}

	return &res.CreateSessionRes{
		ID:                  session.ID,
		MaxDuration:         reservedAmount,
		WebRTCConnectionRes: webrtcRes,
	}, nil
}
