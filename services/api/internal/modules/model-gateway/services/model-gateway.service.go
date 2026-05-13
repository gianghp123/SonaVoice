package services

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"

	"github.com/gianghp123/SonaVoice/api/internal/core/errors"
	zapLogger "github.com/gianghp123/SonaVoice/api/internal/core/zap-logger"
	"github.com/gianghp123/SonaVoice/api/internal/modules/model-gateway/dtos/res"
	"github.com/gianghp123/SonaVoice/api/internal/utils"
	"github.com/google/uuid"
)

type IModelGatewayService interface {
	GetConnnection(ctx context.Context) (*res.WebRTCConnectionRes, *errors.AppError)
	ProxyOffer(ctx context.Context, sessionId string, method string, body []byte) ([]byte, int, *errors.AppError)
}

type modelGatewayService struct {
}

func NewModelGatewayService() IModelGatewayService {
	return &modelGatewayService{}
}

func (s *modelGatewayService) GetConnnection(ctx context.Context) (*res.WebRTCConnectionRes, *errors.AppError) {
	logger := zapLogger.S()
	logger.Debug("Starting connect to speech engine")
	requesterId := utils.GetCtx[string](ctx, "user_id")
	sessionId := "test"

	if requesterId == "" {
		requesterId = "guest_" + uuid.NewString()
		sessionId = ""
	}

	logger.Debugw(
		"Requester from context",
		"requesterId", requesterId,
		"requesterIdLength", len(requesterId),
	)

	body := map[string]interface{}{
		"enableDefaultIceServers": true,
		"body": map[string]interface{}{
			"user_id":    requesterId,
			"session_id": sessionId,
		},
	}

	jsonBody, err := json.Marshal(body)
	if err != nil {
		logger.Error(err)
		return nil, errors.Internal()
	}

	req, err := http.NewRequest(
		http.MethodPost,
		fmt.Sprintf("%s/start", utils.GetEnv("SPEECH_SERVICE_URL", "")),
		bytes.NewBuffer(jsonBody),
	)
	if err != nil {
		logger.Errorw("Failed to create request to speech service", "error", err)
		return nil, errors.Internal()
	}

	req.Header.Set("Content-Type", "application/json")

	client := &http.Client{}

	resp, err := client.Do(req)
	if err != nil {
		logger.Errorw("Failed to connect to speech service", "error", err)
		return nil, errors.Internal()
	}
	defer resp.Body.Close()

	responseBody, err := io.ReadAll(resp.Body)
	if err != nil {
		logger.Errorw("Failed to read response from speech service", "error", err)
		return nil, errors.Internal()
	}

	var result res.WebRTCConnectionRes
	if err := json.Unmarshal(responseBody, &result); err != nil {
		return nil, errors.Internal("failed to parse speech service response")
	}

	return &result, nil
}

func (s *modelGatewayService) ProxyOffer(ctx context.Context, sessionId string, method string, body []byte) ([]byte, int, *errors.AppError) {
	logger := zapLogger.S()
	logger.Debug("Proxying offer to speech engine")

	req, err := http.NewRequest(
		method,
		fmt.Sprintf("%s/sessions/%s/api/offer", utils.GetEnv("SPEECH_SERVICE_URL", ""), sessionId),
		bytes.NewBuffer(body),
	)
	if err != nil {
		logger.Errorw("Failed to create offer request to speech service", "error", err)
		return nil, 0, errors.Internal()
	}

	req.Header.Set("Content-Type", "application/json")

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		logger.Errorw("Failed to proxy offer to speech service", "error", err)
		return nil, 0, errors.Internal()
	}
	defer resp.Body.Close()

	responseBody, err := io.ReadAll(resp.Body)
	if err != nil {
		logger.Errorw("Failed to read offer response from speech service", "error", err)
		return nil, 0, errors.Internal()
	}

	return responseBody, resp.StatusCode, nil
}
