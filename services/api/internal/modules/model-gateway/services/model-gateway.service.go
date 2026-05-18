package services

import (
	"context"

	"github.com/gianghp123/SonaVoice/api/internal/core/enums"
	"github.com/gianghp123/SonaVoice/api/internal/core/errors"
	zapLogger "github.com/gianghp123/SonaVoice/api/internal/core/zap-logger"
	"github.com/gianghp123/SonaVoice/api/internal/database/transaction"
	"github.com/gianghp123/SonaVoice/api/internal/modules/model-gateway/domain"
	"github.com/gianghp123/SonaVoice/api/internal/modules/model-gateway/dtos/req"
	"github.com/gianghp123/SonaVoice/api/internal/modules/model-gateway/dtos/res"
	"github.com/gianghp123/SonaVoice/api/internal/utils"
)

type IModelGatewayService interface {
	CreateSession(ctx context.Context) (*res.CreateSessionRes, *errors.AppError)
	StartConnection(ctx context.Context, sessionID string) (*res.CreateSessionRes, *errors.AppError)
	CloseSession(ctx context.Context, reqBody *req.CloseSessionReq) *errors.AppError
	ProxyOffer(ctx context.Context, sessionId string, method string, body []byte) ([]byte, int, *errors.AppError)
}

type modelGatewayService struct {
	configService      IGlobalConfigService
	sessionService     ISessionService
	speechService      ISpeechProxyService
	startConnectionSvc IStartConnectionService
	uow                transaction.UnitOfWork
}

func NewModelGatewayService(
	configService IGlobalConfigService,
	sessionService ISessionService,
	speechService ISpeechProxyService,
	startConnectionSvc IStartConnectionService,
	uow transaction.UnitOfWork,
) IModelGatewayService {
	return &modelGatewayService{
		configService:      configService,
		sessionService:     sessionService,
		speechService:      speechService,
		startConnectionSvc: startConnectionSvc,
		uow:                uow,
	}
}

func (s *modelGatewayService) CreateSession(ctx context.Context) (*res.CreateSessionRes, *errors.AppError) {
	requesterID := utils.GetCtx[string](ctx, enums.ContextKeyUserID)

	model, appErr := s.configService.Get(ctx)
	if appErr != nil {
		return nil, appErr
	}
	configPayload, err := domain.ParseGlobalConfig(model.Config)
	if err != nil {
		return nil, errors.Internal()
	}

	session, appErr := s.sessionService.Create(ctx, requesterID)
	if appErr != nil {
		return nil, appErr
	}

	return s.startConnectionSvc.Start(ctx, session, requesterID, configPayload.Limits.User.DailyVoiceSeconds)
}

func (s *modelGatewayService) StartConnection(ctx context.Context, sessionID string) (*res.CreateSessionRes, *errors.AppError) {
	requesterID := utils.GetCtx[string](ctx, enums.ContextKeyUserID)

	session, appErr := s.sessionService.Get(ctx, sessionID, requesterID)
	if appErr != nil {
		return nil, appErr
	}

	domainSession := domain.NewSessionFromModel(session)
	if appErr := domainSession.CanBeStarted(); appErr != nil {
		return nil, appErr
	}

	model, appErr := s.configService.Get(ctx)
	if appErr != nil {
		return nil, appErr
	}
	configPayload, err := domain.ParseGlobalConfig(model.Config)
	if err != nil {
		return nil, errors.Internal()
	}

	return s.startConnectionSvc.Start(ctx, session, requesterID, configPayload.Limits.User.DailyVoiceSeconds)
}

func (s *modelGatewayService) ProxyOffer(ctx context.Context, sessionId string, method string, body []byte) ([]byte, int, *errors.AppError) {
	logger := zapLogger.S()
	logger.Debug("Proxying offer to speech engine")

	if sessionId == "" {
		return nil, 0, errors.BadRequest("missing session id")
	}

	requesterID := utils.GetCtx[string](ctx, enums.ContextKeyUserID)

	session, appErr := s.sessionService.GetBySpeechSessionID(ctx, sessionId, requesterID)
	if appErr != nil {
		logger.Errorw("Failed to get app session by speech session id", "speechSessionId", sessionId, "error", appErr)
		return nil, 0, appErr
	}

	responseBody, statusCode, appErr := s.speechService.ProxyOffer(ctx, sessionId, method, body)
	if appErr != nil {
		return nil, 0, appErr
	}

	if statusCode >= 200 && statusCode < 300 && session.Status == enums.SessionStatusPending {
		if err := s.sessionService.MarkSessionActive(ctx, session.ID); err != nil {
			logger.Errorw("Failed to mark session active", "error", err)
		}
	}

	return responseBody, statusCode, nil
}

func (s *modelGatewayService) CloseSession(ctx context.Context, reqBody *req.CloseSessionReq) *errors.AppError {
	logger := zapLogger.S()

	if reqBody == nil {
		return errors.BadRequest("request body is required")
	}

	sessionId := reqBody.SessionID

	logger.Debugw("Closing session", "sessionId", sessionId, "actualUsage", reqBody.ActualUsage)

	if sessionId == "" {
		return errors.BadRequest("sessionId is required")
	}
	if reqBody.ActualUsage < 0 {
		return errors.BadRequest("actualUsage cannot be negative")
	}

	err := s.uow.Do(ctx, func(ctx context.Context, p transaction.IProvider) error {
		sessionRepo := p.Session()
		quotaRepo := p.UserQuota()

		session, err := sessionRepo.GetForUpdate(ctx, sessionId)
		if err != nil {
			return err
		}

		domainSession := domain.NewSessionFromModel(session)
		if appErr := domainSession.CanBeClosed(); appErr != nil {
			return appErr
		}

		if domainSession.WantsQuotaRelease() {
			actualUsage := domainSession.ClampActualUsage(int64(reqBody.ActualUsage))
			reservedAmount := domainSession.ReservedAmount
			unused := reservedAmount - actualUsage
			if unused < 0 {
				unused = 0
			}
			if err := quotaRepo.Release(ctx, session.UserID, "voice", *session.QuotaDate, unused); err != nil {
				return err
			}
		}

		return sessionRepo.UpdateStatus(ctx, session.ID, enums.SessionStatusInactive)
	})

	if err != nil {
		if appErr, ok := err.(*errors.AppError); ok {
			return appErr
		}
		return errors.Internal()
	}
	return nil
}
