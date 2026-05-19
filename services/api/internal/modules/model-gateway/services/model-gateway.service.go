package services

import (
	"context"
	stdErrors "errors"
	"net/http"
	"time"

	"github.com/gianghp123/SonaVoice/api/internal/core/enums"
	"github.com/gianghp123/SonaVoice/api/internal/core/errors"
	zapLogger "github.com/gianghp123/SonaVoice/api/internal/core/zap-logger"
	"github.com/gianghp123/SonaVoice/api/internal/database/models"
	"github.com/gianghp123/SonaVoice/api/internal/database/transaction"
	"github.com/gianghp123/SonaVoice/api/internal/modules/model-gateway/domain"
	"github.com/gianghp123/SonaVoice/api/internal/modules/model-gateway/dtos/req"
	"github.com/gianghp123/SonaVoice/api/internal/modules/model-gateway/dtos/res"
	"github.com/gianghp123/SonaVoice/api/internal/utils"
	"gorm.io/gorm"
)

type IModelGatewayService interface {
	CreateSession(ctx context.Context) (*res.CreateSessionRes, *errors.AppError)
	StartConnection(ctx context.Context, sessionID string) (*res.WebRTCConnectionRes, *errors.AppError)
	CloseSession(ctx context.Context, reqBody *req.CloseSessionReq) *errors.AppError
	CancelSession(ctx context.Context, sessionID string) *errors.AppError
	ProxyOffer(ctx context.Context, speechSessionId string, method string, body []byte) ([]byte, int, *errors.AppError)
}

type modelGatewayService struct {
	configService      IGlobalConfigService
	sessionService     ISessionService
	speechService      ISpeechProxyService
	startConnectionSvc IStartConnectionService
	quotaService       IQuotaService
	uow                transaction.UnitOfWork
}

func NewModelGatewayService(
	configService IGlobalConfigService,
	sessionService ISessionService,
	speechService ISpeechProxyService,
	startConnectionSvc IStartConnectionService,
	quotaService IQuotaService,
	uow transaction.UnitOfWork,
) IModelGatewayService {
	return &modelGatewayService{
		configService:      configService,
		sessionService:     sessionService,
		speechService:      speechService,
		startConnectionSvc: startConnectionSvc,
		quotaService:       quotaService,
		uow:                uow,
	}
}

func (s *modelGatewayService) CreateSession(ctx context.Context) (*res.CreateSessionRes, *errors.AppError) {
	requesterID := utils.GetCtx[string](ctx, enums.ContextKeyUserID)

	model, appErr := s.configService.Get(ctx)
	if appErr != nil {
		return nil, appErr
	}
	if model == nil {
		return nil, errors.Internal("global config missing")
	}
	configPayload, err := domain.ParseGlobalConfig(model.Config)
	if err != nil {
		return nil, errors.Internal()
	}
	dailyLimit := int64(configPayload.Limits.User.DailyVoiceSeconds)

	remaining, err := s.quotaService.CheckRemaining(ctx, requesterID, dailyLimit)
	if err != nil {
		zapLogger.S().Errorw("Failed to check remaining quota", "userId", requesterID, "error", err)
		return nil, errors.Internal()
	}
	if remaining <= 0 {
		return nil, errors.Forbidden("quota exceeded")
	}

	if err := s.cancelStalePendingSession(ctx, requesterID); err != nil {
		zapLogger.S().Warnw("Failed to cancel stale pending session", "userId", requesterID, "error", err)
		return nil, errors.Internal()
	}

	var session *models.Session

	if err := s.uow.Do(ctx, func(ctx context.Context, p transaction.IProvider) error {
		sessionRepo := p.Session()

		quoteDate := utils.QuotaDate()

		session = &models.Session{
			UserID:      requesterID,
			Status:      enums.SessionStatusPending,
			QuotaDate:   &quoteDate,
			MaxDuration: remaining,
		}

		if err := sessionRepo.Create(ctx, session); err != nil {
			zapLogger.S().Errorw("Failed to create session", "error", err)
			return err
		}

		return nil
	}); err != nil {
		return nil, errors.MapRepoError(err)
	}

	return &res.CreateSessionRes{
		ID:                  session.ID,
		MaxDuration:         0,
		WebRTCConnectionRes: nil,
	}, nil
}

func (s *modelGatewayService) StartConnection(ctx context.Context, sessionID string) (*res.WebRTCConnectionRes, *errors.AppError) {
	requesterID := utils.GetCtx[string](ctx, enums.ContextKeyUserID)

	session, appErr := s.sessionService.Get(ctx, sessionID, requesterID)
	if appErr != nil {
		return nil, appErr
	}

	domainSession := domain.NewSessionFromModel(session)
	if appErr := domainSession.CanBeStarted(); appErr != nil {
		return nil, appErr
	}

	return s.startConnectionSvc.Start(ctx, session, requesterID)
}

func (s *modelGatewayService) ProxyOffer(ctx context.Context, speechSessionId string, method string, body []byte) ([]byte, int, *errors.AppError) {
	logger := zapLogger.S()
	logger.Debug("Proxying offer to speech engine")

	if speechSessionId == "" {
		return nil, 0, errors.BadRequest("missing session id")
	}

	requesterID := utils.GetCtx[string](ctx, enums.ContextKeyUserID)

	session, appErr := s.sessionService.GetBySpeechSessionID(ctx, speechSessionId, requesterID)
	if appErr != nil {
		logger.Errorw("Failed to get app session by speech session id", "speechSessionId", speechSessionId, "error", appErr)
		return nil, 0, appErr
	}

	if method == http.MethodPatch {
		speechCtx, cancel := context.WithTimeout(ctx, 10*time.Second)
		defer cancel()

		return s.speechService.ProxyOffer(speechCtx, speechSessionId, method, body)
	}

	var responseBody []byte
	var statusCode int
	var proxyErr *errors.AppError

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

		speechCtx, cancel := context.WithTimeout(ctx, 10*time.Second)
		defer cancel()
		responseBody, statusCode, appErr = s.speechService.ProxyOffer(speechCtx, speechSessionId, method, body)
		if appErr != nil {
			logger.Errorw("Speech engine error during offer", "sessionId", sess.ID, "speechSessionId", speechSessionId, "error", appErr)
			if err := sessionRepo.SetSessionFailed(ctx, sess.ID); err != nil {
				logger.Errorw("Failed to mark session failed", "sessionId", sess.ID, "error", err)
				return err
			}
			proxyErr = appErr
			return nil
		}

		if err := sessionRepo.SetSessionActive(ctx, sess.ID, utils.NowUTC()); err != nil {
			logger.Errorw("Failed to mark session active", "sessionId", sess.ID, "error", err)
			if err := sessionRepo.SetSessionFailed(ctx, sess.ID); err != nil {
				logger.Errorw("Failed to mark session failed after activation error", "sessionId", sess.ID, "error", err)
				return err
			}
			proxyErr = errors.Internal("failed to activate session")
			return nil
		}

		return nil
	})

	if err != nil {
		logger.Errorw("ProxyOffer transaction failed", "speechSessionId", speechSessionId, "error", err)
		return nil, 0, errors.MapRepoError(err)
	}

	if proxyErr != nil {
		return responseBody, statusCode, proxyErr
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

		session, err := sessionRepo.GetForUpdate(ctx, sessionId)
		if err != nil {
			return err
		}

		domainSession := domain.NewSessionFromModel(session)
		if appErr := domainSession.CanBeClosed(); appErr != nil {
			return appErr
		}

		actualUsage := min(int64(reqBody.ActualUsage), domainSession.MaxDuration)

		if actualUsage > 0 {
			quotaRepo := p.UserQuota()
			if err := quotaRepo.Deduct(ctx, session.UserID, "voice", *session.QuotaDate, actualUsage); err != nil {
				logger.Errorw("Failed to deduct quota", "sessionId", sessionId, "error", err)
				return err
			}
		}

		if err := sessionRepo.SetActualUsage(ctx, session.ID, actualUsage); err != nil {
			logger.Errorw("Failed to set actual usage", "sessionId", sessionId, "error", err)
			return err
		}

		return sessionRepo.SetSessionInactive(ctx, session.ID, utils.NowUTC())
	})

	if err != nil {
		return errors.MapRepoError(err)
	}
	return nil
}

func (s *modelGatewayService) CancelSession(ctx context.Context, sessionID string) *errors.AppError {
	logger := zapLogger.S()

	if sessionID == "" {
		return errors.BadRequest("sessionId is required")
	}

	requesterID := utils.GetCtx[string](ctx, enums.ContextKeyUserID)

	err := s.uow.Do(ctx, func(ctx context.Context, p transaction.IProvider) error {
		sessionRepo := p.Session()

		session, err := sessionRepo.GetForUpdate(ctx, sessionID)
		if err != nil {
			return err
		}

		if appErr := utils.EnforceOwnership(session.UserID, requesterID); appErr != nil {
			return appErr
		}

		domainSession := domain.NewSessionFromModel(session)
		if appErr := domainSession.CanBeCancelled(); appErr != nil {
			return appErr
		}

		return sessionRepo.SetSessionInactive(ctx, session.ID, utils.NowUTC())
	})

	if err != nil {
		logger.Errorw("Failed to cancel session", "sessionId", sessionID, "error", err)
		return errors.MapRepoError(err)
	}
	return nil
}

func (s *modelGatewayService) cancelStalePendingSession(ctx context.Context, userID string) error {
	return s.uow.Do(ctx, func(ctx context.Context, p transaction.IProvider) error {
		sessionRepo := p.Session()

		staleSession, err := sessionRepo.GetPendingByUserIDForUpdate(ctx, userID)
		if err != nil {
			if stdErrors.Is(err, gorm.ErrRecordNotFound) {
				return nil
			}
			return err
		}

		if !utils.IsStale(staleSession.CreatedAt, 2*time.Minute) {
			return nil
		}

		return sessionRepo.SetSessionFailed(ctx, staleSession.ID)
	})
}
