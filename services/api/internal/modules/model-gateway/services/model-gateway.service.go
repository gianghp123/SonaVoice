package services

import (
	"context"

	"github.com/gianghp123/SonaVoice/api/internal/core/enums"
	"github.com/gianghp123/SonaVoice/api/internal/core/errors"
	zapLogger "github.com/gianghp123/SonaVoice/api/internal/core/zap-logger"
	"github.com/gianghp123/SonaVoice/api/internal/database/models"
	"github.com/gianghp123/SonaVoice/api/internal/database/transaction"
	"github.com/gianghp123/SonaVoice/api/internal/modules/model-gateway/domain"
	"github.com/gianghp123/SonaVoice/api/internal/modules/model-gateway/dtos/req"
	"github.com/gianghp123/SonaVoice/api/internal/modules/model-gateway/dtos/res"
	"github.com/gianghp123/SonaVoice/api/internal/utils"
)

type IModelGatewayService interface {
	CreateSession(ctx context.Context) (*res.CreateSessionRes, *errors.AppError)
	ResumeSession(ctx context.Context, sessionID string) (*res.CreateSessionRes, *errors.AppError)
	CloseSession(ctx context.Context, reqBody *req.CloseSessionReq) *errors.AppError
	ProxyOffer(ctx context.Context, sessionId string, method string, body []byte) ([]byte, int, *errors.AppError)
}

type modelGatewayService struct {
	configService  IGlobalConfigService
	sessionService ISessionService
	speechService  ISpeechProxyService
	quotaService   ISessionQuotaService
	janitorService ISessionJanitorService
	starterService ISessionStarterService
	uow            transaction.UnitOfWork
}

func NewModelGatewayService(
	configService IGlobalConfigService,
	sessionService ISessionService,
	speechService ISpeechProxyService,
	quotaService ISessionQuotaService,
	janitorService ISessionJanitorService,
	starterService ISessionStarterService,
	uow transaction.UnitOfWork,
) IModelGatewayService {
	return &modelGatewayService{
		configService:  configService,
		sessionService: sessionService,
		speechService:  speechService,
		quotaService:   quotaService,
		janitorService: janitorService,
		starterService: starterService,
		uow:            uow,
	}
}

func (s *modelGatewayService) CreateSession(ctx context.Context) (*res.CreateSessionRes, *errors.AppError) {
	logger := zapLogger.S()
	requesterID := utils.GetCtx[string](ctx, enums.ContextKeyUserID)
	logger.Debugw("Creating session", "userId", requesterID)

	model, appErr := s.configService.Get(ctx)
	if appErr != nil {
		return nil, appErr
	}
	configPayload, err := domain.ParseGlobalConfig(model.Config)
	if err != nil {
		return nil, errors.Internal()
	}

	var session *models.Session
	err = s.uow.Do(ctx, func(ctx context.Context, p transaction.IProvider) error {
		sessionRepo := p.Session()

		if err := sessionRepo.AcquireLock(ctx, requesterID); err != nil {
			return err
		}

		if err := s.janitorService.CleanupStaleSessions(ctx, p, requesterID, int64(configPayload.Limits.Session.MaxSessionLockTTL), int64(configPayload.Limits.User.DailyVoiceSeconds)); err != nil {
			return err
		}

		active, err := sessionRepo.FindActiveByUserID(ctx, requesterID)
		if err != nil {
			return err
		}
		if active != nil {
			return errors.Conflict("close current session before starting a new one")
		}

		session = &models.Session{
			UserID: requesterID,
			Status: enums.SessionStatusPending,
		}
		if err := sessionRepo.Create(ctx, session); err != nil {
			return err
		}

		return nil
	})
	if err != nil {
		if appErr, ok := err.(*errors.AppError); ok {
			return nil, appErr
		}
		return nil, errors.Internal()
	}

	return s.starterService.StartOrResume(ctx, session, requesterID, configPayload.Limits.User.DailyVoiceSeconds)
}

func (s *modelGatewayService) ResumeSession(ctx context.Context, sessionID string) (*res.CreateSessionRes, *errors.AppError) {
	logger := zapLogger.S()
	requesterID := utils.GetCtx[string](ctx, enums.ContextKeyUserID)
	logger.Debugw("Resuming session", "userId", requesterID, "sessionId", sessionID)

	session, appErr := s.sessionService.GetInternal(ctx, sessionID)
	if appErr != nil {
		return nil, appErr
	}

	domainSession := domain.NewSessionFromModel(session)
	if appErr := domainSession.CanBeResumedBy(requesterID); appErr != nil {
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

	err = s.uow.Do(ctx, func(ctx context.Context, p transaction.IProvider) error {
		sessionRepo := p.Session()

		if err := sessionRepo.AcquireLock(ctx, requesterID); err != nil {
			return err
		}

		if err := s.janitorService.CleanupStaleSessions(ctx, p, requesterID, int64(configPayload.Limits.Session.MaxSessionLockTTL), int64(configPayload.Limits.User.DailyVoiceSeconds)); err != nil {
			return err
		}

		active, err := sessionRepo.FindActiveByUserID(ctx, requesterID)
		if err != nil {
			return err
		}
		if active != nil && active.ID != sessionID {
			return errors.Conflict("close current session before starting a new one")
		}

		if err := sessionRepo.UpdateStatus(ctx, session.ID, enums.SessionStatusPending); err != nil {
			return err
		}
		return nil
	})
	if err != nil {
		if appErr, ok := err.(*errors.AppError); ok {
			return nil, appErr
		}
		return nil, errors.Internal()
	}

	session.Status = enums.SessionStatusPending

	return s.starterService.StartOrResume(ctx, session, requesterID, configPayload.Limits.User.DailyVoiceSeconds)
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
		if err := s.sessionService.MarkSessionFailed(ctx, session.ID); err != nil {
			logger.Errorw("Failed to mark session as failed", "error", err)
		}
		_ = s.quotaService.ReleaseAll(ctx, session.UserID, session.ReservedAmount)
		_ = s.sessionService.MarkQuotaReleased(ctx, session.ID)
		return nil, 0, appErr
	}

	if statusCode < 200 || statusCode >= 300 {
		if err := s.sessionService.MarkSessionFailed(ctx, session.ID); err != nil {
			logger.Errorw("Failed to mark session as failed", "error", err)
		}
		_ = s.quotaService.ReleaseAll(ctx, session.UserID, session.ReservedAmount)
		_ = s.sessionService.MarkQuotaReleased(ctx, session.ID)
		return responseBody, statusCode, nil
	}

	if err := s.sessionService.MarkSessionActive(ctx, session.ID); err != nil {
		return nil, 0, err
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

		session, err := sessionRepo.Get(ctx, sessionId)
		if err != nil {
			return err
		}

		domainSession := domain.NewSessionFromModel(session)
		if appErr := domainSession.CanBeClosed(); appErr != nil {
			return appErr
		}

		if domainSession.ShouldReleaseQuota() {
			actualUsage := domainSession.ClampActualUsage(int64(reqBody.ActualUsage))
			reservedAmount := domainSession.ReservedAmount
			dailyQuota := domainSession.DailyQuota
			if reservedAmount <= 0 || dailyQuota <= 0 {
				configModel, err := p.GlobalConfig().Get(ctx)
				if err != nil {
					return err
				}
				configPayload, err := domain.ParseGlobalConfig(configModel.Config)
				if err != nil {
					return err
				}
				dailyQuota = int64(configPayload.Limits.User.DailyVoiceSeconds)
				if reservedAmount <= 0 {
					reservedAmount = dailyQuota
				}
			}
			quotaDate := today()
			unused := reservedAmount - actualUsage
			if unused < 0 {
				unused = 0
			}
			if err := quotaRepo.Release(ctx, session.UserID, "voice", quotaDate, unused); err != nil {
				return err
			}
			if err := sessionRepo.UpdateQuotaReleased(ctx, session.ID); err != nil {
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
