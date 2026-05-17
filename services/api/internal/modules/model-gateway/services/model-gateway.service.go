package services

import (
	"context"
	"net/http"
	"time"

	"github.com/gianghp123/SonaVoice/api/internal/core/enums"
	"github.com/gianghp123/SonaVoice/api/internal/core/errors"
	zapLogger "github.com/gianghp123/SonaVoice/api/internal/core/zap-logger"
	"github.com/gianghp123/SonaVoice/api/internal/modules/model-gateway/dtos/req"
	"github.com/gianghp123/SonaVoice/api/internal/modules/model-gateway/dtos/res"
	"github.com/gianghp123/SonaVoice/api/internal/utils"
)

type IModelGatewayService interface {
	CreateSession(ctx context.Context) (*res.CreateSessionRes, *errors.AppError)
	ResumeSession(ctx context.Context, sessionID string) (*res.CreateSessionRes, *errors.AppError)
	CloseSession(ctx context.Context, reqBody *req.CloseSessionReq) *errors.AppError
	ListSessions(ctx context.Context) ([]*res.SessionListItemRes, *errors.AppError)
	ProxyOffer(ctx context.Context, sessionId string, method string, body []byte) ([]byte, int, *errors.AppError)
}

type modelGatewayService struct {
	configService      IGlobalConfigService
	sessionService     ISessionService
	speechProxyService ISpeechProxyService
	quoteService       IQuoteService
}

func NewModelGatewayService(
	configService IGlobalConfigService,
	sessionService ISessionService,
	speechProxyService ISpeechProxyService,
	quoteService IQuoteService,
) IModelGatewayService {
	return &modelGatewayService{
		configService:      configService,
		sessionService:     sessionService,
		speechProxyService: speechProxyService,
		quoteService:       quoteService,
	}
}

func (s *modelGatewayService) releaseQuotaIfNotReleased(ctx context.Context, session *res.SessionRes, actualUsage int64) *errors.AppError {
	logger := zapLogger.S()
	logger.Debugw("Releasing quota", "sessionId", session.ID, "actualUsage", actualUsage)
	if session.QuotaReleased {
		logger.Debugw("Quota already released, skipping", "sessionId", session.ID)
		return nil
	}
	releaseCtx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()
	if err := s.quoteService.Release(releaseCtx, session.UserID, session.ReservedAmount, actualUsage, session.DailyQuota); err != nil {
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
		if releaseErr := s.releaseQuotaIfNotReleased(ctx, ss, 0); releaseErr != nil {
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

func (s *modelGatewayService) connectToSpeech(ctx context.Context, session *res.SessionRes, requesterID string) (*res.WebRTCConnectionRes, *errors.AppError) {
	body := map[string]interface{}{
		"enableDefaultIceServers": true,
		"body": map[string]interface{}{
			"user_id":      requesterID,
			"session_id":   session.ID,
			"max_duration": session.ReservedAmount,
		},
	}

	result, appErr := s.speechProxyService.StartConnection(ctx, body)
	if appErr != nil {
		_ = s.sessionService.MarkSessionFailed(ctx, session.ID)
		return nil, appErr
	}

	if appErr := s.sessionService.SetSpeechSessionID(ctx, session.ID, result.SessionID); appErr != nil {
		return nil, appErr
	}

	return result, nil
}

func (s *modelGatewayService) startOrResume(ctx context.Context, session *res.SessionRes) (*res.CreateSessionRes, *errors.AppError) {
	requesterID := utils.GetCtx[string](ctx, enums.ContextKeyUserID)

	globalconfig, appErr := s.configService.Get(ctx)
	if appErr != nil {
		return nil, appErr
	}

	if appErr := s.ensureNoActiveSession(ctx, requesterID); appErr != nil {
		return nil, appErr
	}

	if appErr := s.cleanupStaleSessions(ctx, requesterID, int64(globalconfig.Config.Limits.Session.MaxSessionLockTTL)); appErr != nil {
		return nil, appErr
	}

	lockValue, appErr := s.quoteService.AcquireSessionLock(ctx, requesterID, time.Duration(globalconfig.Config.Limits.Session.MaxSessionLockTTL)*time.Second)
	if appErr != nil {
		return nil, appErr
	}
	defer s.quoteService.ReleaseSessionLock(ctx, requesterID, lockValue)

	dailyQuota := globalconfig.Config.Limits.User.DailyVoiceSeconds
	reservedAmount, cleanup, appErr := s.quoteService.ReserveQuota(ctx, requesterID, int64(dailyQuota))
	if appErr != nil {
		return nil, appErr
	}
	defer cleanup(false)

	if appErr := s.sessionService.SetReservation(ctx, session.ID, reservedAmount, int64(dailyQuota)); appErr != nil {
		return nil, appErr
	}

	session.ReservedAmount = reservedAmount
	session.DailyQuota = int64(dailyQuota)

	webrtcRes, appErr := s.connectToSpeech(ctx, session, requesterID)
	if appErr != nil {
		_ = s.releaseQuotaIfNotReleased(ctx, session, 0)
		return nil, appErr
	}

	cleanup(true)

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

	session, appErr := s.sessionService.CreateSession(ctx)
	if appErr != nil {
		return nil, appErr
	}

	return s.startOrResume(ctx, session)
}

func (s *modelGatewayService) ResumeSession(ctx context.Context, sessionID string) (*res.CreateSessionRes, *errors.AppError) {
	logger := zapLogger.S()
	requesterID := utils.GetCtx[string](ctx, enums.ContextKeyUserID)
	logger.Debugw("Resuming session", "userId", requesterID, "sessionId", sessionID)

	session, appErr := s.sessionService.GetSessionInternal(ctx, sessionID)
	if appErr != nil {
		return nil, appErr
	}
	if session.UserID != requesterID {
		return nil, errors.Forbidden()
	}
	if session.Status != enums.SessionStatusInactive {
		return nil, errors.BadRequest("session is not resumable")
	}

	if appErr := s.sessionService.UpdateStatus(ctx, session.ID, enums.SessionStatusPending); appErr != nil {
		return nil, appErr
	}
	session.Status = enums.SessionStatusPending

	return s.startOrResume(ctx, session)
}

func (s *modelGatewayService) ListSessions(ctx context.Context) ([]*res.SessionListItemRes, *errors.AppError) {
	requesterID := utils.GetCtx[string](ctx, enums.ContextKeyUserID)
	return s.sessionService.FindResumableByUserID(ctx, requesterID)
}

func (s *modelGatewayService) ProxyOffer(ctx context.Context, sessionId string, method string, body []byte) ([]byte, int, *errors.AppError) {
	logger := zapLogger.S()
	logger.Debug("Proxying offer to speech engine")

	if sessionId == "" {
		return nil, 0, errors.BadRequest("missing session id")
	}

	session, appErr := s.sessionService.GetSessionBySpeechSessionID(ctx, sessionId)
	if appErr != nil {
		logger.Errorw("Failed to get app session by speech session id", "speechSessionId", sessionId, "error", appErr)
		return nil, 0, appErr
	}

	responseBody, statusCode, appErr := s.speechProxyService.ProxyOffer(ctx, sessionId, method, body)
	if appErr != nil {
		if err := s.sessionService.MarkSessionFailed(ctx, session.ID); err != nil {
			logger.Errorw("Failed to mark session as failed", "error", err)
		}
		_ = s.releaseQuotaIfNotReleased(ctx, session, 0)
		return nil, 0, appErr
	}

	if statusCode < http.StatusOK || statusCode >= http.StatusMultipleChoices {
		if err := s.sessionService.MarkSessionFailed(ctx, session.ID); err != nil {
			logger.Errorw("Failed to mark session as failed", "error", err)
		}
		_ = s.releaseQuotaIfNotReleased(ctx, session, 0)
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

	session, appErr := s.sessionService.GetSessionInternal(ctx, sessionId)
	if appErr != nil {
		logger.Errorw("Failed to get session", "sessionId", sessionId, "error", appErr)
		return appErr
	}

	if session.Status == enums.SessionStatusInactive {
		logger.Errorw("Session is already inactive", "sessionId", sessionId, "status", session.Status)
		return errors.BadRequest("session is already inactive")
	}

	if session.QuotaReleased {
		logger.Debugw("Quota already released, just marking inactive", "sessionId", sessionId)
		if err := s.sessionService.MarkSessionInactive(ctx, sessionId); err != nil {
			logger.Errorw("Failed to mark session inactive", "sessionId", sessionId, "error", err)
			return err
		}
		return nil
	}

	actualUsage := int64(reqBody.ActualUsage)
	if actualUsage > session.ReservedAmount {
		logger.Warnw("Actual usage clamped to reserved amount", "sessionId", sessionId, "actualUsage", actualUsage, "reservedAmount", session.ReservedAmount)
		actualUsage = session.ReservedAmount
	}

	if appErr := s.releaseQuotaIfNotReleased(ctx, session, actualUsage); appErr != nil {
		return appErr
	}

	if err := s.sessionService.MarkSessionInactive(ctx, sessionId); err != nil {
		logger.Errorw("Failed to mark session inactive", "sessionId", sessionId, "error", err)
		return err
	}

	return nil
}
