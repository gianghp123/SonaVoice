package services

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/gianghp123/SonaVoice/api/internal/core/errors"
	zapLogger "github.com/gianghp123/SonaVoice/api/internal/core/zap-logger"
	httpclient "github.com/gianghp123/SonaVoice/api/internal/http-client"
	"github.com/gianghp123/SonaVoice/api/internal/modules/model-gateway/dtos/res"
	"github.com/gianghp123/SonaVoice/api/internal/utils"
)

type ISpeechProxyService interface {
	StartConnection(ctx context.Context, body map[string]interface{}) (*res.WebRTCConnectionRes, *errors.AppError)
	ProxyOffer(ctx context.Context, speechSessionID, method string, body []byte) ([]byte, int, *errors.AppError)
}

type speechProxyService struct {
	httpClient httpclient.IHttpClient
}

func NewSpeechProxyService(httpClient httpclient.IHttpClient) ISpeechProxyService {
	return &speechProxyService{httpClient: httpClient}
}

func (s *speechProxyService) StartConnection(ctx context.Context, body map[string]interface{}) (*res.WebRTCConnectionRes, *errors.AppError) {
	logger := zapLogger.S()

	responseBody, statusCode, appErr := s.httpClient.Do(
		ctx,
		http.MethodPost,
		fmt.Sprintf("%s/start", utils.GetEnv("SPEECH_SERVICE_URL", "")),
		map[string]string{"Content-Type": "application/json"},
		body,
	)
	if appErr != nil {
		return nil, appErr
	}

	if statusCode < http.StatusOK || statusCode >= http.StatusMultipleChoices {
		logger.Errorw("Speech service returned non-success status", "statusCode", statusCode, "responseBody", string(responseBody))
		return nil, errors.Internal("speech service failed")
	}

	var result res.WebRTCConnectionRes
	if err := json.Unmarshal(responseBody, &result); err != nil {
		logger.Errorw("Failed to parse speech service response", "error", err, "responseBody", string(responseBody))
		return nil, errors.Internal("failed to parse speech service response")
	}

	return &result, nil
}

func (s *speechProxyService) ProxyOffer(ctx context.Context, speechSessionID, method string, body []byte) ([]byte, int, *errors.AppError) {
	logger := zapLogger.S()

	responseBody, statusCode, appErr := s.httpClient.Do(
		ctx,
		method,
		fmt.Sprintf("%s/sessions/%s/api/offer", utils.GetEnv("SPEECH_SERVICE_URL", ""), speechSessionID),
		map[string]string{"Content-Type": "application/json"},
		json.RawMessage(body),
	)
	if appErr != nil {
		return nil, 0, appErr
	}

	if statusCode < http.StatusOK || statusCode >= http.StatusMultipleChoices {
		logger.Errorw("Speech service returned non-success status while proxying offer", "speechSessionId", speechSessionID, "statusCode", statusCode, "responseBody", string(responseBody))
		return responseBody, statusCode, nil
	}

	return responseBody, statusCode, nil
}
