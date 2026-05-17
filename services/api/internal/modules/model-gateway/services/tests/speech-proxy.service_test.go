package tests

import (
	"context"
	"encoding/json"
	"net/http"
	"os"
	"testing"

	appErrors "github.com/gianghp123/SonaVoice/api/internal/core/errors"
	httpclientmocks "github.com/gianghp123/SonaVoice/api/internal/http-client/mocks"
	"github.com/gianghp123/SonaVoice/api/internal/modules/model-gateway/dtos/req"
	"github.com/gianghp123/SonaVoice/api/internal/modules/model-gateway/dtos/res"
	"github.com/gianghp123/SonaVoice/api/internal/modules/model-gateway/services"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
)

func init() {
	os.Setenv("SPEECH_SERVICE_URL", "http://speech.local")
}

func TestSpeechProxyService_StartConnection(t *testing.T) {
	webrtcRes := res.WebRTCConnectionRes{
		SessionID: "speech-s1",
		IceConfig: &res.IceConfig{
			IceServers: []res.IceServer{
				{Urls: []string{"stun:stun.l.google.com:19302"}},
			},
		},
	}
	webrtcBytes, _ := json.Marshal(webrtcRes)

	connReq := &req.StartConnectionReq{
		EnableDefaultIceServers: true,
		Body: req.StartConnectionBody{
			UserID:    "u1",
			SessionID: "s1",
		},
	}

	tests := []struct {
		name      string
		setupMock func(mockHTTP *httpclientmocks.HttpClient)
		wantErr   bool
		errCode   int
		wantSID   string
	}{
		{
			name: "success",
			setupMock: func(mockHTTP *httpclientmocks.HttpClient) {
				mockHTTP.On("Do", mock.Anything, http.MethodPost, "http://speech.local/start", mock.Anything, mock.Anything).
					Return(webrtcBytes, http.StatusOK, (*appErrors.AppError)(nil))
			},
			wantSID: "speech-s1",
		},
		{
			name: "http do error",
			setupMock: func(mockHTTP *httpclientmocks.HttpClient) {
				mockHTTP.On("Do", mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything).
					Return([]byte{}, 0, appErrors.Internal("http error"))
			},
			wantErr: true,
			errCode: http.StatusInternalServerError,
		},
		{
			name: "non-2xx status",
			setupMock: func(mockHTTP *httpclientmocks.HttpClient) {
				mockHTTP.On("Do", mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything).
					Return([]byte("error"), http.StatusInternalServerError, (*appErrors.AppError)(nil))
			},
			wantErr: true,
			errCode: http.StatusInternalServerError,
		},
		{
			name: "unmarshal error",
			setupMock: func(mockHTTP *httpclientmocks.HttpClient) {
				mockHTTP.On("Do", mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything).
					Return([]byte("not json"), http.StatusOK, (*appErrors.AppError)(nil))
			},
			wantErr: true,
			errCode: http.StatusInternalServerError,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockHTTP := new(httpclientmocks.HttpClient)
			tt.setupMock(mockHTTP)

			svc := services.NewSpeechProxyService(mockHTTP)
			ctx := context.Background()

			result, appErr := svc.StartConnection(ctx, connReq)

			if tt.wantErr {
				require.NotNil(t, appErr)
				assert.Equal(t, tt.errCode, appErr.Code)
				return
			}
			require.Nil(t, appErr)
			assert.Equal(t, tt.wantSID, result.SessionID)
			mockHTTP.AssertExpectations(t)
		})
	}
}

func TestSpeechProxyService_ProxyOffer(t *testing.T) {
	responseBody := []byte(`{"status":"ok"}`)

	tests := []struct {
		name            string
		speechSessionID string
		method          string
		body            []byte
		setupMock       func(mockHTTP *httpclientmocks.HttpClient)
		wantErr         bool
		wantStatus      int
		wantBody        []byte
	}{
		{
			name:            "success",
			speechSessionID: "speech-1",
			method:          http.MethodPost,
			body:            []byte(`{"offer":true}`),
			setupMock: func(mockHTTP *httpclientmocks.HttpClient) {
				mockHTTP.On("Do", mock.Anything, http.MethodPost, "http://speech.local/sessions/speech-1/api/offer", mock.Anything, mock.Anything).
					Return(responseBody, http.StatusOK, (*appErrors.AppError)(nil))
			},
			wantStatus: http.StatusOK,
		},
		{
			name:            "http do error",
			speechSessionID: "speech-1",
			method:          http.MethodPost,
			body:            []byte(`{}`),
			setupMock: func(mockHTTP *httpclientmocks.HttpClient) {
				mockHTTP.On("Do", mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything).
					Return([]byte{}, 0, appErrors.Internal("http error"))
			},
			wantErr: true,
		},
		{
			name:            "non-2xx status returns body no error",
			speechSessionID: "speech-1",
			method:          http.MethodPost,
			body:            []byte(`{}`),
			setupMock: func(mockHTTP *httpclientmocks.HttpClient) {
				mockHTTP.On("Do", mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything).
					Return([]byte("gateway error"), http.StatusBadGateway, (*appErrors.AppError)(nil))
			},
			wantStatus: http.StatusBadGateway,
			wantBody:   []byte("gateway error"),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockHTTP := new(httpclientmocks.HttpClient)
			tt.setupMock(mockHTTP)

			svc := services.NewSpeechProxyService(mockHTTP)
			ctx := context.Background()

			respBody, statusCode, appErr := svc.ProxyOffer(ctx, tt.speechSessionID, tt.method, tt.body)

			if tt.wantErr {
				require.NotNil(t, appErr)
				return
			}
			require.Nil(t, appErr)
			assert.Equal(t, tt.wantStatus, statusCode)
			if tt.wantBody != nil {
				assert.Equal(t, tt.wantBody, respBody)
			} else {
				assert.Equal(t, responseBody, respBody)
			}
			mockHTTP.AssertExpectations(t)
		})
	}
}
