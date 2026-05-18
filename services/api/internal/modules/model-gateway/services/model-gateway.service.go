package services

import (
	"context"
	stdErrors "errors"
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

	return s.startConnectionSvc.Start(ctx, session, requesterID, configPayload.Limits.User.DailyVoiceSeconds)
}

func (s *modelGatewayService) markSessionFailedAndReleaseQuota(ctx context.Context, p transaction.IProvider, session *models.Session) error {
	sessionRepo := p.Session()
	quotaRepo := p.UserQuota()

	if session.QuotaDate != nil && session.ReservedAmount > 0 {
		if err := quotaRepo.Release(ctx, session.UserID, "voice", *session.QuotaDate, session.ReservedAmount); err != nil {
			return err
		}
	}

	return sessionRepo.SetSessionFailed(ctx, session.ID)
}

func (s *modelGatewayService) ProxyOffer(ctx context.Context, sessionId string, method string, body []byte) ([]byte, int, *errors.AppError) {
	logger := zapLogger.S()
	logger.Debug("Proxying offer to speech engine")

	if sessionId == "" {
		return nil, 0, errors.BadRequest("missing session id")
	}

	requesterID := utils.GetCtx[string](ctx, enums.ContextKeyUserID)

	// Pre-validation: get session by speech session ID (outside tx, cheap)
	session, appErr := s.sessionService.GetBySpeechSessionID(ctx, sessionId, requesterID)
	if appErr != nil {
		logger.Errorw("Failed to get app session by speech session id", "speechSessionId", sessionId, "error", appErr)
		return nil, 0, appErr
	}

	var responseBody []byte
	var statusCode int
	var proxyErr *errors.AppError

	err := s.uow.Do(ctx, func(ctx context.Context, p transaction.IProvider) error {
		sessionRepo := p.Session()

		// 1. Lock the row for the duration of the offer
		sess, err := sessionRepo.GetForUpdate(ctx, session.ID)
		if err != nil {
			return err
		}

		// 2. Validate state machine: must be PENDING to transition
		domainSession := domain.NewSessionFromModel(sess)
		if appErr := domainSession.CanBeStarted(); appErr != nil {
			return appErr
		}

		// 3. Forward to speech engine (lock held during HTTP call)
		responseBody, statusCode, appErr = s.speechService.ProxyOffer(ctx, sessionId, method, body)
		if appErr != nil {
			// 4a. Speech engine error → mark FAILED, release quota, capture error
			if failErr := s.markSessionFailedAndReleaseQuota(ctx, p, sess); failErr != nil {
				logger.Errorw("Failed to mark session failed after speech error", "sessionId", sess.ID, "error", failErr)
				return failErr
			}
			proxyErr = appErr
			return nil
		}

		if statusCode < 200 || statusCode >= 300 {
			// 4b. Non-2xx from speech engine → same as error
			if failErr := s.markSessionFailedAndReleaseQuota(ctx, p, sess); failErr != nil {
				logger.Errorw("Failed to mark session failed after non-2xx", "sessionId", sess.ID, "statusCode", statusCode, "error", failErr)
				return failErr
			}
			return nil // non-2xx is not a Go error; caller handles status code
		}

		// 5. Success → mark ACTIVE (status precondition ensures no race)
		if err := sessionRepo.SetSessionActive(ctx, sess.ID, time.Now().UTC()); err != nil {
			logger.Errorw("Failed to mark session active", "sessionId", sess.ID, "error", err)
			// If activation fails, roll back by marking failed + releasing quota
			if failErr := s.markSessionFailedAndReleaseQuota(ctx, p, sess); failErr != nil {
				logger.Errorw("Failed to mark session failed after activation error", "sessionId", sess.ID, "error", failErr)
				return failErr
			}
			proxyErr = errors.Internal("failed to activate session")
			return nil
		}

		return nil
	})

	if err != nil {
		if appErr, ok := err.(*errors.AppError); ok {
			return nil, 0, appErr
		}
		logger.Errorw("ProxyOffer transaction failed", "sessionId", sessionId, "error", err)
		return nil, 0, errors.Internal()
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
			if unused > 0 {
				if err := quotaRepo.Release(ctx, session.UserID, "voice", *session.QuotaDate, unused); err != nil {
					return err
				}
			}
		}

		return sessionRepo.SetSessionInactive(ctx, session.ID, time.Now().UTC())
	})

	if err != nil {
		if appErr, ok := err.(*errors.AppError); ok {
			return appErr
		}
		return errors.Internal()
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
		quotaRepo := p.UserQuota()

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

		if domainSession.WantsQuotaRelease() {
			if err := quotaRepo.Release(ctx, session.UserID, "voice", *session.QuotaDate, session.ReservedAmount); err != nil {
				return err
			}
		}

		return sessionRepo.SetSessionInactive(ctx, session.ID, time.Now().UTC())
	})

	if err != nil {
		if appErr, ok := err.(*errors.AppError); ok {
			return appErr
		}
		logger.Errorw("Failed to cancel session", "sessionId", sessionID, "error", err)
		return errors.Internal()
	}
	return nil
}

func (s *modelGatewayService) cancelStalePendingSession(ctx context.Context, userID string) error {
	return s.uow.Do(ctx, func(ctx context.Context, p transaction.IProvider) error {
		sessionRepo := p.Session()
		quotaRepo := p.UserQuota()

		staleSession, err := sessionRepo.GetPendingByUserIDForUpdate(ctx, userID)
		if err != nil {
			if stdErrors.Is(err, gorm.ErrRecordNotFound) {
				return nil // Not found = nothing to do
			}
			return err // Real DB error — will be logged by caller
		}

		staleThreshold := time.Now().UTC().Add(-2 * time.Minute)
		if staleSession.CreatedAt.After(staleThreshold) {
			return nil // Not stale enough
		}

		if staleSession.QuotaDate != nil && staleSession.ReservedAmount > 0 {
			if err := quotaRepo.Release(ctx, staleSession.UserID, "voice", *staleSession.QuotaDate, staleSession.ReservedAmount); err != nil {
				return err
			}
		}

		return sessionRepo.SetSessionInactive(ctx, staleSession.ID, time.Now().UTC())
	})
}

func (s *modelGatewayService) CreateSession(ctx context.Context) (*res.CreateSessionRes, *errors.AppError) {
	requesterID := utils.GetCtx[string](ctx, enums.ContextKeyUserID)

	// Opportunistic stale PENDING cleanup (serverless — no background jobs)
	if err := s.cancelStalePendingSession(ctx, requesterID); err != nil {
		zapLogger.S().Warnw("Failed to cancel stale pending session", "userId", requesterID, "error", err)
	}

	session, appErr := s.sessionService.Create(ctx, requesterID)
	if appErr != nil {
		return nil, appErr
	}

	return &res.CreateSessionRes{
		ID:                  session.ID,
		MaxDuration:         0,
		WebRTCConnectionRes: nil,
	}, nil
}
