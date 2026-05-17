package services

import (
	"context"
	"net/http"
	"time"

	"github.com/gianghp123/SonaVoice/api/internal/core/compensations"
	"github.com/gianghp123/SonaVoice/api/internal/core/enums"
	"github.com/gianghp123/SonaVoice/api/internal/core/errors"
	"github.com/gianghp123/SonaVoice/api/internal/core/quota"
	zapLogger "github.com/gianghp123/SonaVoice/api/internal/core/zap-logger"
	"github.com/gianghp123/SonaVoice/api/internal/database/models"
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
	configService      IGlobalConfigService
	sessionService     ISessionService
	speechProxyService ISpeechProxyService
	quotaService       quota.IQuotaService
	lockService        ISessionLockService
}

func NewModelGatewayService(
	configService IGlobalConfigService,
	sessionService ISessionService,
	speechProxyService ISpeechProxyService,
	quotaService quota.IQuotaService,
	lockService ISessionLockService,
) IModelGatewayService {
	return &modelGatewayService{
		configService:      configService,
		sessionService:     sessionService,
		speechProxyService: speechProxyService,
		quotaService:       quotaService,
		lockService:        lockService,
	}
}

func (s *modelGatewayService) releaseQuotaIfNotReleased(ctx context.Context, session *domain.Session, actualUsage int64) *errors.AppError {
	logger := zapLogger.S()
	logger.Debugw("Releasing quota", "sessionId", session.ID, "actualUsage", actualUsage)
	if !session.ShouldReleaseQuota() {
		logger.Debugw("Quota already released, skipping", "sessionId", session.ID)
		return nil
	}
	releaseCtx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()
	if err := s.quotaService.Release(releaseCtx, session.UserID, quota.QuotaConfig{Key: "voice", DailyLimit: session.DailyQuota}, session.ReservedAmount, actualUsage); err != nil {
		logger.Errorw("Failed to release quota", "sessionId", session.ID, "userId", session.UserID, "error", err)
		return errors.Internal("failed to release quota")
	}
	if appErr := s.sessionService.MarkQuotaReleased(ctx, session.ID); appErr != nil {
		logger.Errorw("Failed to mark quota released", "sessionId", session.ID, "error", appErr)
		return appErr
	}
	return nil
}

func (s *modelGatewayService) ensureNoActiveSession(ctx context.Context, userID string) *errors.AppError {
	active, appErr := s.sessionService.FindActiveByUserID(ctx, userID)
	if appErr != nil {
		return appErr
	}
	if active != nil {
		return errors.Conflict("close current session before starting a new one")
	}
	return nil
}

func (s *modelGatewayService) cleanupStaleSessions(ctx context.Context, userID string, pendingTimeoutSeconds int64) *errors.AppError {
	logger := zapLogger.S()
	staleSessions, appErr := s.sessionService.FindStaleSessions(ctx, userID, pendingTimeoutSeconds)
	if appErr != nil {
		logger.Errorw("Failed to find stale sessions", "error", appErr)
		return appErr
	}
	for _, ss := range staleSessions {
		domainSession := domain.NewSessionFromModel(ss)
		if releaseErr := s.releaseQuotaIfNotReleased(ctx, domainSession, 0); releaseErr != nil {
			logger.Errorw("Failed to release quota for stale session", "sessionId", ss.ID, "error", releaseErr)
			return releaseErr
		}
		if markErr := s.sessionService.MarkSessionInactive(ctx, ss.ID); markErr != nil {
			logger.Errorw("Failed to mark stale session inactive", "sessionId", ss.ID, "error", markErr)
			return markErr
		}
	}
	return nil
}

func (s *modelGatewayService) connectToSpeech(ctx context.Context, sessionID string, reservedAmount int64, requesterID string) (*res.WebRTCConnectionRes, *errors.AppError) {
	connReq := &req.StartConnectionReq{
		EnableDefaultIceServers: true,
		Body: req.StartConnectionBody{
			UserID:      requesterID,
			SessionID:   sessionID,
			MaxDuration: reservedAmount,
		},
	}

	result, appErr := s.speechProxyService.StartConnection(ctx, connReq)
	if appErr != nil {
		_ = s.sessionService.MarkSessionFailed(ctx, sessionID)
		return nil, appErr
	}

	if appErr := s.sessionService.SetSpeechSessionID(ctx, sessionID, result.SessionID); appErr != nil {
		return nil, appErr
	}

	return result, nil
}

func (s *modelGatewayService) startOrResume(ctx context.Context, session *models.Session) (*res.CreateSessionRes, *errors.AppError) {
	requesterID := utils.GetCtx[string](ctx, enums.ContextKeyUserID)

	model, appErr := s.configService.Get(ctx)
	if appErr != nil {
		return nil, appErr
	}
	configPayload, err := domain.ParseGlobalConfig(model.Config)
	if err != nil {
		return nil, errors.Internal()
	}

	var comp compensations.Compensations
	defer comp.Run()

	lockValue, appErr := s.lockService.Acquire(ctx, requesterID, time.Duration(configPayload.Limits.Session.MaxSessionLockTTL)*time.Second)
	if appErr != nil {
		return nil, appErr
	}
	comp.Push(func() { _ = s.lockService.Release(ctx, requesterID, lockValue) })

	if appErr := s.ensureNoActiveSession(ctx, requesterID); appErr != nil {
		return nil, appErr
	}

	if appErr := s.cleanupStaleSessions(ctx, requesterID, int64(configPayload.Limits.Session.MaxSessionLockTTL)); appErr != nil {
		return nil, appErr
	}

	dailyQuota := configPayload.Limits.User.DailyVoiceSeconds
	quotaCfg := quota.QuotaConfig{Key: "voice", DailyLimit: int64(dailyQuota)}
	reservedAmount, err := s.quotaService.ReserveAll(ctx, requesterID, quotaCfg)
	if err != nil {
		logger := zapLogger.S()
		logger.Errorw("Failed to reserve quota", "userId", requesterID, "error", err)
		return nil, errors.Internal()
	}
	if reservedAmount <= 0 {
		return nil, errors.Forbidden("quota exceeded")
	}
	comp.Push(func() {
		releaseCtx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
		defer cancel()
		if err := s.quotaService.Release(releaseCtx, requesterID, quotaCfg, reservedAmount, 0); err != nil {
			zapLogger.S().Errorw("Failed to rollback reserved quota", "userId", requesterID, "reservedAmount", reservedAmount, "error", err)
		}
	})

	if appErr := s.sessionService.SetReservation(ctx, session.ID, reservedAmount, int64(dailyQuota)); appErr != nil {
		return nil, appErr
	}

	session.ReservedAmount = reservedAmount
	session.DailyQuota = int64(dailyQuota)

	webrtcRes, appErr := s.connectToSpeech(ctx, session.ID, reservedAmount, requesterID)
	if appErr != nil {
		_ = s.releaseQuotaIfNotReleased(ctx, domain.NewSessionFromModel(session), 0)
		comp.Pop()
		return nil, appErr
	}

	comp.Pop()

	return &res.CreateSessionRes{
		ID:                  session.ID,
		MaxDuration:         reservedAmount,
		WebRTCConnectionRes: webrtcRes,
	}, nil
}

func (s *modelGatewayService) CreateSession(ctx context.Context) (*res.CreateSessionRes, *errors.AppError) {
	logger := zapLogger.S()
	requesterID := utils.GetCtx[string](ctx, enums.ContextKeyUserID)
	logger.Debugw("Creating session", "userId", requesterID)

	session, appErr := s.sessionService.Create(ctx, requesterID)
	if appErr != nil {
		return nil, appErr
	}

	return s.startOrResume(ctx, session)
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

	if appErr := s.sessionService.UpdateStatus(ctx, session.ID, enums.SessionStatusPending); appErr != nil {
		return nil, appErr
	}
	session.Status = enums.SessionStatusPending

	return s.startOrResume(ctx, session)
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

	responseBody, statusCode, appErr := s.speechProxyService.ProxyOffer(ctx, sessionId, method, body)
	if appErr != nil {
		if err := s.sessionService.MarkSessionFailed(ctx, session.ID); err != nil {
			logger.Errorw("Failed to mark session as failed", "error", err)
		}
		_ = s.releaseQuotaIfNotReleased(ctx, domain.NewSessionFromModel(session), 0)
		return nil, 0, appErr
	}

	if statusCode < http.StatusOK || statusCode >= http.StatusMultipleChoices {
		if err := s.sessionService.MarkSessionFailed(ctx, session.ID); err != nil {
			logger.Errorw("Failed to mark session as failed", "error", err)
		}
		_ = s.releaseQuotaIfNotReleased(ctx, domain.NewSessionFromModel(session), 0)
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

	session, appErr := s.sessionService.GetInternal(ctx, sessionId)
	if appErr != nil {
		logger.Errorw("Failed to get session", "sessionId", sessionId, "error", appErr)
		return appErr
	}

	domainSession := domain.NewSessionFromModel(session)
	if appErr := domainSession.CanBeClosed(); appErr != nil {
		logger.Errorw("Session is already inactive", "sessionId", sessionId, "status", session.Status)
		return appErr
	}

	if !domainSession.ShouldReleaseQuota() {
		logger.Debugw("Quota already released, just marking inactive", "sessionId", sessionId)
		if err := s.sessionService.MarkSessionInactive(ctx, sessionId); err != nil {
			logger.Errorw("Failed to mark session inactive", "sessionId", sessionId, "error", err)
			return err
		}
		return nil
	}

	actualUsage := domainSession.ClampActualUsage(int64(reqBody.ActualUsage))

	if appErr := s.releaseQuotaIfNotReleased(ctx, domainSession, actualUsage); appErr != nil {
		return appErr
	}

	if err := s.sessionService.MarkSessionInactive(ctx, sessionId); err != nil {
		logger.Errorw("Failed to mark session inactive", "sessionId", sessionId, "error", err)
		return err
	}

	return nil
}
