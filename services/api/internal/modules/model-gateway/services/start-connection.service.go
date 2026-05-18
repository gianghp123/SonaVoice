package services

import (
	"context"
	"time"

	"github.com/gianghp123/SonaVoice/api/internal/core/errors"
	zapLogger "github.com/gianghp123/SonaVoice/api/internal/core/zap-logger"
	"github.com/gianghp123/SonaVoice/api/internal/database/models"
	"github.com/gianghp123/SonaVoice/api/internal/database/transaction"
	"github.com/gianghp123/SonaVoice/api/internal/modules/model-gateway/domain"
	"github.com/gianghp123/SonaVoice/api/internal/modules/model-gateway/dtos/req"
	"github.com/gianghp123/SonaVoice/api/internal/modules/model-gateway/dtos/res"
	"github.com/gianghp123/SonaVoice/api/internal/utils"
)

type IStartConnectionService interface {
	Start(ctx context.Context, session *models.Session, userID string, dailyQuota int) (*res.WebRTCConnectionRes, *errors.AppError)
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

func (s *startConnectionService) Start(ctx context.Context, session *models.Session, userID string, dailyQuota int) (*res.WebRTCConnectionRes, *errors.AppError) {
	logger := zapLogger.S()

	logger.Debugw("starting connection", "sessionId", session.ID, "userId", userID, "dailyQuota", dailyQuota)

	var reservedAmount int64
	var quotaDate time.Time
	var webrtcRes *res.WebRTCConnectionRes

	err := s.uow.Do(ctx, func(ctx context.Context, p transaction.IProvider) error {
		quotaRepo := p.UserQuota()
		sessionRepo := p.Session()

		sess, err := sessionRepo.GetForUpdate(ctx, session.ID)
		if err != nil {
			return err
		}
		domainSession := domain.NewSessionFromModel(sess)
		if appErr := domainSession.CanBeStarted(); appErr != nil {
			return appErr
		}

		quotaDate = utils.QuotaDate()
		reservedAmount, err = quotaRepo.Reserve(ctx, userID, "voice", quotaDate, int64(dailyQuota))
		if err != nil {
			return err
		}
		if reservedAmount <= 0 {
			logger.Errorw("Reserved amount is less than or equal to 0", "sessionId", session.ID, "userId", userID, "dailyQuota", dailyQuota)
			if dailyQuota <= 0 {
				return errors.Internal("voice service is not configured")
			}
			return errors.Forbidden("quota exceeded")
		}

		if err := sessionRepo.SetQuotaDate(ctx, session.ID, quotaDate); err != nil {
			return err
		}
		if err := sessionRepo.SetReservedAmount(ctx, session.ID, reservedAmount); err != nil {
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
		if appErr, ok := err.(*errors.AppError); ok {
			return nil, appErr
		}
		logger.Errorw("Failed to start connection", "sessionId", session.ID, "error", err)
		return nil, errors.Internal()
	}

	return webrtcRes, nil
}


