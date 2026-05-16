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
	"github.com/gianghp123/SonaVoice/api/internal/modules/model-gateway/dtos/req"
	"github.com/gianghp123/SonaVoice/api/internal/modules/model-gateway/dtos/res"
	"github.com/gianghp123/SonaVoice/api/internal/utils"
	"github.com/google/uuid"
)

type IModelGatewayService interface {
	CreateSession(ctx context.Context) (*res.SessionRes, *errors.AppError)
	StartConnection(ctx context.Context, reqBody *req.StartConnectionReq) (*res.WebRTCConnectionRes, *errors.AppError)
	ProxyOffer(ctx context.Context, sessionId string, method string, body []byte) ([]byte, int, *errors.AppError)
}

type modelGatewayService struct {
	httpClient   httpclient.IHttpClient
	sessionRepo  repository_interfaces.ISessionRepository
	quoteService IQuoteService
}

func NewModelGatewayService(httpClient httpclient.IHttpClient, sessionRepo repository_interfaces.ISessionRepository, quoteService IQuoteService) IModelGatewayService {
	return &modelGatewayService{
		httpClient:   httpClient,
		sessionRepo:  sessionRepo,
		quoteService: quoteService,
	}
}

func (s *modelGatewayService) CreateSession(
	ctx context.Context,
) (*res.SessionRes, *errors.AppError) {
	logger := zapLogger.S()

	requesterId := utils.GetCtx[string](ctx, enums.ContextKeyUserID)

	logger.Debugw("Creating session", "userId", requesterId)

	session := &models.Session{
		UserID: requesterId,
		Status: enums.SessionStatusPending,
	}

	if err := s.sessionRepo.Create(ctx, session); err != nil {
		logger.Errorw("Failed to create session", "error", err)
		return nil, errors.Internal("failed to create session")
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

	isGuest := requesterId == ""
	if isGuest {
		requesterId = "guest_" + uuid.NewString()
		sessionId = ""
	} else if sessionId != "" {
		existingSession, err := s.sessionRepo.Get(ctx, sessionId)
		if err != nil {
			logger.Errorw("Failed to get session", "error", err)
			return nil, errors.Internal("failed to get session")
		}

		if existingSession.UserID != requesterId {
			return nil, errors.Forbidden("session does not belong to requester")
		}

		session = existingSession
		sessionId = session.ID
	} else {
		return nil, errors.BadRequest("session_id is required")
	}

	body := map[string]interface{}{
		"enableDefaultIceServers": true,
		"body": map[string]interface{}{
			"user_id":    requesterId,
			"session_id": sessionId,
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
		logger.Warnw(
			"No app session found for speech session id; continuing as guest",
			"speechSessionId", sessionId,
			"error", err,
		)
		appSession = nil
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
