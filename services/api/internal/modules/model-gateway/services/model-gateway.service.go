package services

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	"github.com/gianghp123/SonaVoice/api/internal/core/enums"
	"github.com/gianghp123/SonaVoice/api/internal/core/errors"
	zapLogger "github.com/gianghp123/SonaVoice/api/internal/core/zap-logger"
	"github.com/gianghp123/SonaVoice/api/internal/database/models"
	repository_interfaces "github.com/gianghp123/SonaVoice/api/internal/database/repository-interfaces"
	httpclient "github.com/gianghp123/SonaVoice/api/internal/http-client"
	"github.com/gianghp123/SonaVoice/api/internal/modules/model-gateway/dtos"
	"github.com/gianghp123/SonaVoice/api/internal/modules/model-gateway/dtos/req"
	"github.com/gianghp123/SonaVoice/api/internal/modules/model-gateway/dtos/res"
	"github.com/gianghp123/SonaVoice/api/internal/utils"
)

type IModelGatewayService interface {
	CreateSession(ctx context.Context) (*res.SessionRes, *errors.AppError)
	StartConnection(ctx context.Context, reqBody *req.StartConnectionReq) (*res.WebRTCConnectionRes, *errors.AppError)
	CloseSession(ctx context.Context, reqBody *req.CloseSessionReq) *errors.AppError
	ProxyOffer(ctx context.Context, sessionId string, method string, body []byte) ([]byte, int, *errors.AppError)
}

type modelGatewayService struct {
	httpClient       httpclient.IHttpClient
	sessionRepo      repository_interfaces.ISessionRepository
	globalConfigRepo repository_interfaces.IGlobalConfigRepository
	quoteService     IQuoteService
}

func NewModelGatewayService(httpClient httpclient.IHttpClient, sessionRepo repository_interfaces.ISessionRepository, globalConfigRepo repository_interfaces.IGlobalConfigRepository, quoteService IQuoteService) IModelGatewayService {
	return &modelGatewayService{
		httpClient:       httpClient,
		sessionRepo:      sessionRepo,
		globalConfigRepo: globalConfigRepo,
		quoteService:     quoteService,
	}
}

func (s *modelGatewayService) CreateSession(
	ctx context.Context,
) (*res.SessionRes, *errors.AppError) {
	logger := zapLogger.S()

	requesterId := utils.GetCtx[string](ctx, enums.ContextKeyUserID)

	logger.Debugw("Creating session", "userId", requesterId)

	globalconfig, appErr := s.GetGlobalConfig(ctx)
	if appErr != nil {
		return nil, appErr
	}

	lockValue, err := s.quoteService.AcquireSessionLock(ctx, requesterId, time.Duration(globalconfig.Config.Limits.Session.MaxSessionLockTTL)*time.Second)

	if err != nil {
		return nil, errors.Internal("failed to acquire session lock")
	}

	defer s.quoteService.ReleaseSessionLock(ctx, requesterId, lockValue)

	session := &models.Session{
		UserID: requesterId,
		Status: enums.SessionStatusPending,
	}

	if err := s.sessionRepo.Create(ctx, session); err != nil {
		logger.Errorw("Failed to create session", "error", err)
		return nil, errors.MapRepoError(err)
	}

	return &res.SessionRes{
		ID:        session.ID,
		UserID:    session.UserID,
		Status:    string(session.Status),
		CreatedAt: session.CreatedAt,
	}, nil
}

func (s *modelGatewayService) StartConnection(
	ctx context.Context,
	reqBody *req.StartConnectionReq,
) (*res.WebRTCConnectionRes, *errors.AppError) {
	logger := zapLogger.S()
	logger.Debug("Starting connect to speech engine")

	requesterId := utils.GetCtx[string](ctx, enums.ContextKeyUserID)

	var sessionId string
	var session *models.Session

	if reqBody != nil {
		sessionId = reqBody.SessionId
	}

	if sessionId == "" {
		return nil, errors.BadRequest("session_id is required")
	}

	existingSession, err := s.sessionRepo.Get(ctx, sessionId)
	if err != nil {
		logger.Errorw("Failed to get session", "error", err)
		return nil, errors.Internal("failed to get session")
	}

	if existingSession.UserID != requesterId {
		return nil, errors.Forbidden("session does not belong to requester")
	}

	session = existingSession

	globalconfig, appErr := s.GetGlobalConfig(ctx)
	if appErr != nil {
		return nil, appErr
	}

	dailyQuota := globalconfig.Config.Limits.User.DailyVoiceSeconds

	reservedAmount, err := s.quoteService.ReserveAllRemaining(ctx, requesterId, int64(dailyQuota))
	if err != nil {
		return nil, errors.Internal()
	}

	if reservedAmount <= 0 {
		return nil, errors.Forbidden("quota exceeded")
	}

	quotaCommitted := false

	defer func() {
		if quotaCommitted {
			return
		}

		releaseCtx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
		defer cancel()

		if err := s.quoteService.Release(
			releaseCtx,
			requesterId,
			reservedAmount,
			0, // actualUsage = 0 because session did not successfully start
			int64(dailyQuota),
		); err != nil {
			logger.Errorw(
				"Failed to rollback reserved quota",
				"userId", requesterId,
				"reservedAmount", reservedAmount,
				"error", err,
			)
		}
	}()

	maxDuration := reservedAmount

	body := map[string]interface{}{
		"enableDefaultIceServers": true,
		"body": map[string]interface{}{
			"user_id":     requesterId,
			"session_id":  sessionId,
			"maxDuration": maxDuration,
		},
	}

	failSession := func(reason string, fields ...interface{}) {
		logger.Errorw(reason, fields...)

		if session == nil {
			return
		}

		session.Status = enums.SessionStatusFailed

		if err := s.sessionRepo.Update(ctx, session); err != nil {
			logger.Errorw("Failed to update session to failed", "error", err)
		}
	}

	responseBody, statusCode, appErr := s.httpClient.Do(
		ctx,
		http.MethodPost,
		fmt.Sprintf("%s/start", utils.GetEnv("SPEECH_SERVICE_URL", "")),
		map[string]string{
			"Content-Type": "application/json",
		},
		body,
	)
	if appErr != nil {
		failSession("Failed to connect to speech service", "error", appErr)
		return nil, appErr
	}

	if statusCode < http.StatusOK || statusCode >= http.StatusMultipleChoices {
		failSession(
			"Speech service returned non-success status",
			"statusCode", statusCode,
			"responseBody", string(responseBody),
		)

		return nil, errors.Internal("speech service failed")
	}

	var result res.WebRTCConnectionRes
	if err := json.Unmarshal(responseBody, &result); err != nil {
		failSession(
			"Failed to parse speech service response",
			"error", err,
			"responseBody", string(responseBody),
		)

		return nil, errors.Internal("failed to parse speech service response")
	}

	if session != nil {
		session.SpeechSessionID = result.SessionID
		if err := s.sessionRepo.Update(ctx, session); err != nil {
			logger.Errorw("Failed to update session to active", "error", err)
			return nil, errors.Internal("failed to update speechSessionId")
		}
	}

	quotaCommitted = true

	return &result, nil
}

func (s *modelGatewayService) ProxyOffer(
	ctx context.Context,
	sessionId string,
	method string,
	body []byte,
) ([]byte, int, *errors.AppError) {
	logger := zapLogger.S()
	logger.Debug("Proxying offer to speech engine")

	if sessionId == "" {
		return nil, 0, errors.BadRequest("missing session id")
	}

	appSession, err := s.sessionRepo.GetBySpeechSessionID(ctx, sessionId)
	if err != nil {
		logger.Errorw(
			"Failed to get app session by speech session id",
			"speechSessionId", sessionId,
			"error", err,
		)
		return nil, 0, errors.Internal("failed to get session")
	}

	failSession := func(reason string, fields ...interface{}) {
		logger.Errorw(reason, fields...)

		if appSession == nil {
			return
		}

		appSession.Status = enums.SessionStatusFailed

		if err := s.sessionRepo.Update(ctx, appSession); err != nil {
			logger.Errorw(
				"Failed to update session to failed",
				"sessionId", appSession.ID,
				"speechSessionId", sessionId,
				"error", err,
			)
		}
	}

	responseBody, statusCode, appErr := s.httpClient.Do(
		ctx,
		method,
		fmt.Sprintf(
			"%s/sessions/%s/api/offer",
			utils.GetEnv("SPEECH_SERVICE_URL", ""),
			sessionId,
		),
		map[string]string{
			"Content-Type": "application/json",
		},
		json.RawMessage(body),
	)
	if appErr != nil {
		failSession(
			"Failed to proxy offer to speech service",
			"speechSessionId", sessionId,
			"error", appErr,
		)

		return nil, 0, appErr
	}

	if statusCode < http.StatusOK || statusCode >= http.StatusMultipleChoices {
		failSession(
			"Speech service returned non-success status while proxying offer",
			"speechSessionId", sessionId,
			"statusCode", statusCode,
			"responseBody", string(responseBody),
		)

		return responseBody, statusCode, nil
	}

	if appSession != nil {
		appSession.StartedAt = time.Now()
		appSession.Status = enums.SessionStatusActive

		if err := s.sessionRepo.Update(ctx, appSession); err != nil {
			logger.Errorw(
				"Failed to update session to active",
				"sessionId", appSession.ID,
				"speechSessionId", sessionId,
				"error", err,
			)
			return nil, 0, errors.Internal("failed to update session to active")
		}
	}

	return responseBody, statusCode, nil
}

func (s *modelGatewayService) CloseSession(
	ctx context.Context,
	reqBody *req.CloseSessionReq,
) *errors.AppError {
	logger := zapLogger.S()

	if reqBody == nil {
		return errors.BadRequest("request body is required")
	}

	sessionId := reqBody.SessionID

	logger.Debugw(
		"Closing session",
		"SessionId", sessionId,
		"actualUsage", reqBody.ActualUsage,
		"maxDuration", reqBody.MaxDuration,
	)

	if sessionId == "" {
		return errors.BadRequest("sessionId is required")
	}

	if reqBody.MaxDuration <= 0 {
		return errors.BadRequest("maxDuration must be greater than 0")
	}

	if reqBody.ActualUsage < 0 {
		return errors.BadRequest("actualUsage cannot be negative")
	}

	session, err := s.sessionRepo.Get(ctx, sessionId)
	if err != nil {
		logger.Errorw(
			"Failed to get session",
			"speechSessionId", sessionId,
			"error", err,
		)
		return errors.Internal("failed to get session")
	}

	if session.Status != enums.SessionStatusActive {
		logger.Errorw(
			"Session is not active",
			"speechSessionId", sessionId,
			"status", session.Status,
		)
		return errors.BadRequest("session is not active")
	}

	globalConfig, appErr := s.GetGlobalConfig(ctx)
	if appErr != nil {
		return appErr
	}

	dailyQuota := globalConfig.Config.Limits.User.DailyVoiceSeconds

	reservedAmount := int64(reqBody.MaxDuration)
	actualUsage := int64(reqBody.ActualUsage)

	if actualUsage > reservedAmount {
		logger.Warnw(
			"Actual usage is greater than reserved amount; clamping to reserved amount",
			"speechSessionId", sessionId,
			"actualUsage", actualUsage,
			"reservedAmount", reservedAmount,
		)

		actualUsage = reservedAmount
	}

	releaseCtx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	if err := s.quoteService.Release(
		releaseCtx,
		session.UserID,
		reservedAmount,
		actualUsage,
		int64(dailyQuota),
	); err != nil {
		logger.Errorw(
			"Failed to release quota",
			"speechSessionId", sessionId,
			"userId", session.UserID,
			"reservedAmount", reservedAmount,
			"actualUsage", actualUsage,
			"error", err,
		)

		return errors.Internal("failed to release quota")
	}

	session.Status = enums.SessionStatusInactive

	if err := s.sessionRepo.Update(ctx, session); err != nil {
		logger.Errorw(
			"Failed to update session to inactive",
			"speechSessionId", sessionId,
			"error", err,
		)

		return errors.Internal("failed to update session")
	}

	return nil
}

func (s *modelGatewayService) GetGlobalConfig(ctx context.Context) (*dtos.GlobalConfig, *errors.AppError) {
	logger := zapLogger.S()
	globalconfig, err := s.globalConfigRepo.Get(ctx)

	if err != nil {
		logger.Errorw("Failed to get global config", "error", err)
		return nil, errors.Internal("failed to get global config")
	}

	var globalConfigDTO dtos.GlobalConfig
	err = utils.MapToDTO(globalconfig, &globalConfigDTO)

	if err != nil {
		logger.Errorw("Failed to map global config to dto", "error", err)
		return nil, errors.Internal("failed to map global config to dto")
	}

	return &globalConfigDTO, nil
}
