package services

import (
	"context"
	"time"

	"github.com/gianghp123/SonaVoice/api/internal/core/errors"
	zapLogger "github.com/gianghp123/SonaVoice/api/internal/core/zap-logger"
	"github.com/gianghp123/SonaVoice/api/internal/database/models"
	"github.com/gianghp123/SonaVoice/api/internal/database/transaction"
	"github.com/gianghp123/SonaVoice/api/internal/modules/session/domain"
	"github.com/gianghp123/SonaVoice/api/internal/modules/session/dtos/req"
	"github.com/gianghp123/SonaVoice/api/internal/modules/session/dtos/res"
)

type IStartConnectionService interface {
	Start(ctx context.Context, session *models.Session, userID string) (*res.WebRTCConnectionRes, *errors.AppError)
}

type startConnectionService struct {
	speechSvc ISpeechProxyService
	uow       transaction.UnitOfWork
}

func NewStartConnectionService(speechSvc ISpeechProxyService, uow transaction.UnitOfWork) IStartConnectionService {
	return &startConnectionService{
		speechSvc: speechSvc,
		uow:       uow,
	}
}

func (s *startConnectionService) Start(ctx context.Context, session *models.Session, userID string) (*res.WebRTCConnectionRes, *errors.AppError) {
	logger := zapLogger.S()

	logger.Debugw("starting connection", "sessionId", session.ID, "userId", userID)

	connReq := &req.StartConnectionReq{
		EnableDefaultIceServers: true,
		Body: req.StartConnectionBody{
			UserID:      userID,
			SessionID:   session.ID,
			MaxDuration: session.MaxDuration,
		},
	}

	var webrtcRes *res.WebRTCConnectionRes

	err := s.uow.Do(ctx, func(ctx context.Context, p transaction.IProvider) error {
		sessionRepo := p.Session()

		sess, err := sessionRepo.GetForUpdate(ctx, session.ID)
		if err != nil {
			return err
		}
		domainSession := domain.NewSessionFromModel(sess)
		if appErr := domainSession.CanBeStarted(); appErr != nil {
			return appErr
		}

		var appErr *errors.AppError
		speechCtx, cancel := context.WithTimeout(ctx, 15*time.Second)
		defer cancel()
		webrtcRes, appErr = s.speechSvc.StartConnection(speechCtx, connReq)
		if appErr != nil {
			return appErr
		}

		if err := sessionRepo.UpdateSpeechSessionID(ctx, session.ID, webrtcRes.SessionID); err != nil {
			return err
		}

		return nil
	})

	if err != nil {
		logger.Errorw("Failed to start connection", "sessionId", session.ID, "error", err)
		return nil, errors.MapRepoError(err)
	}

	return webrtcRes, nil
}
