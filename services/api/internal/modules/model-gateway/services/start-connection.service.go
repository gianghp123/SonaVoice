package services

import (
	"context"
	"time"

	"github.com/gianghp123/SonaVoice/api/internal/core/enums"
	"github.com/gianghp123/SonaVoice/api/internal/core/errors"
	zapLogger "github.com/gianghp123/SonaVoice/api/internal/core/zap-logger"
	"github.com/gianghp123/SonaVoice/api/internal/database/models"
	"github.com/gianghp123/SonaVoice/api/internal/database/transaction"
	"github.com/gianghp123/SonaVoice/api/internal/modules/model-gateway/dtos/req"
	"github.com/gianghp123/SonaVoice/api/internal/modules/model-gateway/dtos/res"
)

type IStartConnectionService interface {
	Start(ctx context.Context, session *models.Session, userID string, dailyQuota int) (*res.CreateSessionRes, *errors.AppError)
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

func (s *startConnectionService) Start(ctx context.Context, session *models.Session, userID string, dailyQuota int) (*res.CreateSessionRes, *errors.AppError) {
	logger := zapLogger.S()

	var reservedAmount int64
	var quotaDate time.Time
	var webrtcRes *res.WebRTCConnectionRes

	err := s.uow.Do(ctx, func(ctx context.Context, p transaction.IProvider) error {
		quotaRepo := p.UserQuota()
		sessionRepo := p.Session()

		quotaDate = time.Now().Truncate(24 * time.Hour)
		var err error
		reservedAmount, err = quotaRepo.Reserve(ctx, userID, "voice", quotaDate, int64(dailyQuota))
		if err != nil {
			return err
		}
		if reservedAmount <= 0 {
			return errors.Forbidden("quota exceeded")
		}

		if err := sessionRepo.SetQuotaDate(ctx, session.ID, quotaDate); err != nil {
			return err
		}

		connReq := &req.StartConnectionReq{
			EnableDefaultIceServers: true,
			Body: req.StartConnectionBody{
				UserID:      userID,
				SessionID:   session.ID,
				MaxDuration: reservedAmount,
			},
		}

		var appErr *errors.AppError
		webrtcRes, appErr = s.speechSvc.StartConnection(ctx, connReq)
		if appErr != nil {
			return appErr
		}

		if err := sessionRepo.UpdateSpeechSessionID(ctx, session.ID, webrtcRes.SessionID); err != nil {
			return err
		}

		return nil
	})

	if err != nil {
		if appErr, ok := err.(*errors.AppError); ok {
			return nil, appErr
		}
		if reservedAmount > 0 {
			s.cleanupFailedSession(ctx, session.ID, userID, reservedAmount, quotaDate)
		}
		logger.Errorw("Failed to start connection", "sessionId", session.ID, "error", err)
		return nil, errors.Internal()
	}

	return &res.CreateSessionRes{
		ID:                  session.ID,
		MaxDuration:         reservedAmount,
		WebRTCConnectionRes: webrtcRes,
	}, nil
}

func (s *startConnectionService) cleanupFailedSession(ctx context.Context, sessionID string, userID string, reservedAmount int64, quotaDate time.Time) {
	logger := zapLogger.S()
	err := s.uow.Do(ctx, func(ctx context.Context, p transaction.IProvider) error {
		quotaRepo := p.UserQuota()
		sessionRepo := p.Session()

		if err := quotaRepo.Release(ctx, userID, "voice", quotaDate, reservedAmount); err != nil {
			return err
		}
		if err := sessionRepo.UpdateStatus(ctx, sessionID, enums.SessionStatusFailed); err != nil {
			return err
		}
		return nil
	})
	if err != nil {
		logger.Errorw("Failed to cleanup failed session", "sessionId", sessionID, "error", err)
	}
}
