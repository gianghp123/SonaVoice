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
	StartConnection(ctx context.Context) (*res.CreateSessionRes, *errors.AppError)
	CloseSession(ctx context.Context, reqBody *req.CloseSessionReq) *errors.AppError
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

func (s *modelGatewayService) StartConnection(ctx context.Context) (*res.CreateSessionRes, *errors.AppError) {
	logger := zapLogger.S()
	requesterId := utils.GetCtx[string](ctx, enums.ContextKeyUserID)
	logger.Debugw("Starting connect to speech engine", "userId", requesterId)

	globalconfig, appErr := s.configService.Get(ctx)
	if appErr != nil {
		return nil, appErr
	}

	lockValue, err := s.quoteService.AcquireSessionLock(ctx, requesterId, time.Duration(globalconfig.Config.Limits.Session.MaxSessionLockTTL)*time.Second)
	if err != nil {
		return nil, errors.Internal("failed to acquire session lock")
	}
	defer s.quoteService.ReleaseSessionLock(ctx, requesterId, lockValue)

	dailyQuota := globalconfig.Config.Limits.User.DailyVoiceSeconds

	reservedAmount, cleanup, appErr := s.quoteService.ReserveQuota(ctx, requesterId, int64(dailyQuota))
	if appErr != nil {
		return nil, appErr
	}
	defer cleanup(false)

	session, appErr := s.sessionService.CreateSession(ctx)
	if appErr != nil {
		return nil, appErr
	}

	maxDuration := reservedAmount

	body := map[string]interface{}{
		"enableDefaultIceServers": true,
		"body": map[string]interface{}{
			"user_id":      requesterId,
			"session_id":   session.ID,
			"max_duration": maxDuration,
		},
	}

	result, appErr := s.speechProxyService.StartConnection(ctx, body)
	if appErr != nil {
		if err := s.sessionService.MarkSessionFailed(ctx, session.ID); err != nil {
			logger.Errorw("Failed to mark session as failed", "error", err)
		}
		return nil, appErr
	}

	if err := s.sessionService.SetSpeechSessionID(ctx, session.ID, result.SessionID); err != nil {
		return nil, err
	}

	cleanup(true)

	var response res.CreateSessionRes

	response.ID = session.ID
	response.WebRTCConnectionRes = result
	return &response, nil
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
		return nil, 0, appErr
	}

	if statusCode < http.StatusOK || statusCode >= http.StatusMultipleChoices {
		if err := s.sessionService.MarkSessionFailed(ctx, session.ID); err != nil {
			logger.Errorw("Failed to mark session as failed", "error", err)
		}
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

	logger.Debugw("Closing session", "SessionId", sessionId, "actualUsage", reqBody.ActualUsage, "maxDuration", reqBody.MaxDuration)

	if sessionId == "" {
		return errors.BadRequest("sessionId is required")
	}
	if reqBody.MaxDuration <= 0 {
		return errors.BadRequest("maxDuration must be greater than 0")
	}
	if reqBody.ActualUsage < 0 {
		return errors.BadRequest("actualUsage cannot be negative")
	}

	session, appErr := s.sessionService.GetSession(ctx, sessionId)
	if appErr != nil {
		logger.Errorw("Failed to get session", "speechSessionId", sessionId, "error", appErr)
		return appErr
	}

	if session.Status != enums.SessionStatusActive {
		logger.Errorw("Session is not active", "speechSessionId", sessionId, "status", session.Status)
		return errors.BadRequest("session is not active")
	}

	globalConfig, appErr := s.configService.Get(ctx)
	if appErr != nil {
		return appErr
	}

	dailyQuota := globalConfig.Config.Limits.User.DailyVoiceSeconds

	reservedAmount := int64(reqBody.MaxDuration)
	actualUsage := int64(reqBody.ActualUsage)

	if actualUsage > reservedAmount {
		logger.Warnw("Actual usage clamped to reserved amount", "speechSessionId", sessionId, "actualUsage", actualUsage, "reservedAmount", reservedAmount)
		actualUsage = reservedAmount
	}

	releaseCtx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	if err := s.quoteService.Release(releaseCtx, session.UserID, reservedAmount, actualUsage, int64(dailyQuota)); err != nil {
		logger.Errorw("Failed to release quota", "speechSessionId", sessionId, "userId", session.UserID, "reservedAmount", reservedAmount, "actualUsage", actualUsage, "error", err)
		return errors.Internal("failed to release quota")
	}

	if err := s.sessionService.MarkSessionInactive(ctx, sessionId); err != nil {
		logger.Errorw("Failed to mark session inactive", "speechSessionId", sessionId, "error", err)
		return err
	}

	return nil
}
