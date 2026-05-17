package tests

import (
	"context"
	"encoding/json"
	"net/http"
	"testing"
	"time"

	appErrors "github.com/gianghp123/SonaVoice/api/internal/core/errors"
	"github.com/gianghp123/SonaVoice/api/internal/core/enums"
	"github.com/gianghp123/SonaVoice/api/internal/modules/model-gateway/dtos"
	"github.com/gianghp123/SonaVoice/api/internal/modules/model-gateway/dtos/req"
	"github.com/gianghp123/SonaVoice/api/internal/modules/model-gateway/dtos/res"
	"github.com/gianghp123/SonaVoice/api/internal/modules/model-gateway/services"
	svcMocks "github.com/gianghp123/SonaVoice/api/internal/modules/model-gateway/services/mocks"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
)

func TestModelGatewayService_CreateSession(t *testing.T) {
	defaultConfig := &dtos.GlobalConfig{
		Config: dtos.ConfigPayload{
			Limits: dtos.LimitsConfig{
				Session: dtos.SessionLimitConfig{MaxSessionLockTTL: 60},
				User:    dtos.UserLimitConfig{DailyVoiceSeconds: 3600},
			},
		},
	}
	sessionRes := &res.SessionRes{ID: "s1", UserID: "user-1"}
	webrtcRes := &res.WebRTCConnectionRes{SessionID: "speech-s1"}
	cleanup := func(commit bool) {}

	tests := []struct {
		name      string
		setupMock func(
			configSvc *svcMocks.GlobalConfigService,
			sessionSvc *svcMocks.SessionService,
			speechSvc *svcMocks.SpeechProxyService,
			quoteSvc *svcMocks.QuoteService,
		)
		wantErr bool
		errCode int
	}{
		{
			name: "success",
			setupMock: func(configSvc *svcMocks.GlobalConfigService, sessionSvc *svcMocks.SessionService, speechSvc *svcMocks.SpeechProxyService, quoteSvc *svcMocks.QuoteService) {
				configSvc.On("Get", mock.Anything).Return(defaultConfig, (*appErrors.AppError)(nil))
				sessionSvc.On("FindActiveByUserID", mock.Anything, "user-1").Return(nil, (*appErrors.AppError)(nil))
				sessionSvc.On("FindStaleSessions", mock.Anything, "user-1", int64(60)).Return([]*res.SessionRes{}, (*appErrors.AppError)(nil))
				quoteSvc.On("AcquireSessionLock", mock.Anything, "user-1", 60*time.Second).Return("lock-val", nil)
				quoteSvc.On("ReleaseSessionLock", mock.Anything, "user-1", "lock-val").Return(nil)
				quoteSvc.On("ReserveQuota", mock.Anything, "user-1", int64(3600)).Return(int64(300), func(bool) {}, (*appErrors.AppError)(nil))
				sessionSvc.On("CreateSession", mock.Anything).Return(sessionRes, (*appErrors.AppError)(nil))
				sessionSvc.On("SetReservation", mock.Anything, "s1", int64(300), int64(3600)).Return((*appErrors.AppError)(nil))
				speechSvc.On("StartConnection", mock.Anything, mock.Anything).Return(webrtcRes, (*appErrors.AppError)(nil))
				sessionSvc.On("SetSpeechSessionID", mock.Anything, "s1", "speech-s1").Return((*appErrors.AppError)(nil))
			},
		},
		{
			name: "config service get error",
			setupMock: func(configSvc *svcMocks.GlobalConfigService, sessionSvc *svcMocks.SessionService, speechSvc *svcMocks.SpeechProxyService, quoteSvc *svcMocks.QuoteService) {
				sessionSvc.On("CreateSession", mock.Anything).Return(sessionRes, (*appErrors.AppError)(nil))
				configSvc.On("Get", mock.Anything).Return((*dtos.GlobalConfig)(nil), appErrors.Internal("config error"))
			},
			wantErr: true,
			errCode: http.StatusInternalServerError,
		},
		{
			name: "acquire session lock error",
			setupMock: func(configSvc *svcMocks.GlobalConfigService, sessionSvc *svcMocks.SessionService, speechSvc *svcMocks.SpeechProxyService, quoteSvc *svcMocks.QuoteService) {
				sessionSvc.On("CreateSession", mock.Anything).Return(sessionRes, (*appErrors.AppError)(nil))
				configSvc.On("Get", mock.Anything).Return(defaultConfig, (*appErrors.AppError)(nil))
				sessionSvc.On("FindActiveByUserID", mock.Anything, "user-1").Return(nil, (*appErrors.AppError)(nil))
				sessionSvc.On("FindStaleSessions", mock.Anything, "user-1", int64(60)).Return([]*res.SessionRes{}, (*appErrors.AppError)(nil))
				quoteSvc.On("AcquireSessionLock", mock.Anything, "user-1", 60*time.Second).Return("", appErrors.Internal("lock error"))
			},
			wantErr: true,
			errCode: http.StatusInternalServerError,
		},
		{
			name: "reserve quota error",
			setupMock: func(configSvc *svcMocks.GlobalConfigService, sessionSvc *svcMocks.SessionService, speechSvc *svcMocks.SpeechProxyService, quoteSvc *svcMocks.QuoteService) {
				sessionSvc.On("CreateSession", mock.Anything).Return(sessionRes, (*appErrors.AppError)(nil))
				configSvc.On("Get", mock.Anything).Return(defaultConfig, (*appErrors.AppError)(nil))
				sessionSvc.On("FindActiveByUserID", mock.Anything, "user-1").Return(nil, (*appErrors.AppError)(nil))
				sessionSvc.On("FindStaleSessions", mock.Anything, "user-1", int64(60)).Return([]*res.SessionRes{}, (*appErrors.AppError)(nil))
				quoteSvc.On("AcquireSessionLock", mock.Anything, "user-1", 60*time.Second).Return("lock-val", nil)
				quoteSvc.On("ReleaseSessionLock", mock.Anything, "user-1", "lock-val").Return(nil)
				quoteSvc.On("ReserveQuota", mock.Anything, "user-1", int64(3600)).Return(int64(0), (func(bool))(nil), appErrors.Forbidden("quota exceeded"))
			},
			wantErr: true,
			errCode: http.StatusForbidden,
		},
		{
			name: "session create error",
			setupMock: func(configSvc *svcMocks.GlobalConfigService, sessionSvc *svcMocks.SessionService, speechSvc *svcMocks.SpeechProxyService, quoteSvc *svcMocks.QuoteService) {
				configSvc.On("Get", mock.Anything).Return(defaultConfig, (*appErrors.AppError)(nil))
				sessionSvc.On("FindActiveByUserID", mock.Anything, "user-1").Return(nil, (*appErrors.AppError)(nil))
				sessionSvc.On("FindStaleSessions", mock.Anything, "user-1", int64(60)).Return([]*res.SessionRes{}, (*appErrors.AppError)(nil))
				quoteSvc.On("AcquireSessionLock", mock.Anything, "user-1", 60*time.Second).Return("lock-val", nil)
				quoteSvc.On("ReleaseSessionLock", mock.Anything, "user-1", "lock-val").Return(nil)
				quoteSvc.On("ReserveQuota", mock.Anything, "user-1", int64(3600)).Return(int64(300), cleanup, (*appErrors.AppError)(nil))
				sessionSvc.On("CreateSession", mock.Anything).Return((*res.SessionRes)(nil), appErrors.Internal("create error"))
			},
			wantErr: true,
			errCode: http.StatusInternalServerError,
		},
		{
			name: "speech proxy start error marks session failed and releases quota",
			setupMock: func(configSvc *svcMocks.GlobalConfigService, sessionSvc *svcMocks.SessionService, speechSvc *svcMocks.SpeechProxyService, quoteSvc *svcMocks.QuoteService) {
				configSvc.On("Get", mock.Anything).Return(defaultConfig, (*appErrors.AppError)(nil))
				sessionSvc.On("FindActiveByUserID", mock.Anything, "user-1").Return(nil, (*appErrors.AppError)(nil))
				sessionSvc.On("FindStaleSessions", mock.Anything, "user-1", int64(60)).Return([]*res.SessionRes{}, (*appErrors.AppError)(nil))
				quoteSvc.On("AcquireSessionLock", mock.Anything, "user-1", 60*time.Second).Return("lock-val", nil)
				quoteSvc.On("ReleaseSessionLock", mock.Anything, "user-1", "lock-val").Return(nil)
				quoteSvc.On("ReserveQuota", mock.Anything, "user-1", int64(3600)).Return(int64(300), cleanup, (*appErrors.AppError)(nil))
				sessionSvc.On("CreateSession", mock.Anything).Return(sessionRes, (*appErrors.AppError)(nil))
				sessionSvc.On("SetReservation", mock.Anything, "s1", int64(300), int64(3600)).Return((*appErrors.AppError)(nil))
				speechSvc.On("StartConnection", mock.Anything, mock.Anything).Return((*res.WebRTCConnectionRes)(nil), appErrors.Internal("speech error"))
				sessionSvc.On("MarkSessionFailed", mock.Anything, "s1").Return((*appErrors.AppError)(nil))
				quoteSvc.On("Release", mock.Anything, "user-1", int64(300), int64(0), int64(3600)).Return(nil)
				sessionSvc.On("MarkQuotaReleased", mock.Anything, "s1").Return((*appErrors.AppError)(nil))
			},
			wantErr: true,
			errCode: http.StatusInternalServerError,
		},
		{
			name: "set speech session id error releases quota",
			setupMock: func(configSvc *svcMocks.GlobalConfigService, sessionSvc *svcMocks.SessionService, speechSvc *svcMocks.SpeechProxyService, quoteSvc *svcMocks.QuoteService) {
				configSvc.On("Get", mock.Anything).Return(defaultConfig, (*appErrors.AppError)(nil))
				sessionSvc.On("FindActiveByUserID", mock.Anything, "user-1").Return(nil, (*appErrors.AppError)(nil))
				sessionSvc.On("FindStaleSessions", mock.Anything, "user-1", int64(60)).Return([]*res.SessionRes{}, (*appErrors.AppError)(nil))
				quoteSvc.On("AcquireSessionLock", mock.Anything, "user-1", 60*time.Second).Return("lock-val", nil)
				quoteSvc.On("ReleaseSessionLock", mock.Anything, "user-1", "lock-val").Return(nil)
				quoteSvc.On("ReserveQuota", mock.Anything, "user-1", int64(3600)).Return(int64(300), cleanup, (*appErrors.AppError)(nil))
				sessionSvc.On("CreateSession", mock.Anything).Return(sessionRes, (*appErrors.AppError)(nil))
				sessionSvc.On("SetReservation", mock.Anything, "s1", int64(300), int64(3600)).Return((*appErrors.AppError)(nil))
				speechSvc.On("StartConnection", mock.Anything, mock.Anything).Return(webrtcRes, (*appErrors.AppError)(nil))
				sessionSvc.On("SetSpeechSessionID", mock.Anything, "s1", "speech-s1").Return(appErrors.Internal("update error"))
				quoteSvc.On("Release", mock.Anything, "user-1", int64(300), int64(0), int64(3600)).Return(nil)
				sessionSvc.On("MarkQuotaReleased", mock.Anything, "s1").Return((*appErrors.AppError)(nil))
			},
			wantErr: true,
			errCode: http.StatusInternalServerError,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			configSvc := new(svcMocks.GlobalConfigService)
			sessionSvc := new(svcMocks.SessionService)
			speechSvc := new(svcMocks.SpeechProxyService)
			quoteSvc := new(svcMocks.QuoteService)

			tt.setupMock(configSvc, sessionSvc, speechSvc, quoteSvc)

			svc := services.NewModelGatewayService(configSvc, sessionSvc, speechSvc, quoteSvc)
			ctx := setupSessionCtx("user-1")

			result, appErr := svc.CreateSession(ctx)

			if tt.wantErr {
				require.NotNil(t, appErr)
				assert.Equal(t, tt.errCode, appErr.Code)
				return
			}
			require.Nil(t, appErr)
			assert.Equal(t, "speech-s1", result.WebRTCConnectionRes.SessionID)
		})
	}
}

func TestModelGatewayService_ProxyOffer(t *testing.T) {
	sessionRes := &res.SessionRes{ID: "s1", UserID: "user-1", Status: "active", ReservedAmount: 300, DailyQuota: 3600, QuotaReleased: false}
	responseBody := []byte(`{"ok":true}`)
	nonSuccessBody := []byte(`{"error":"gateway error"}`)

	tests := []struct {
		name      string
		sessionID string
		method    string
		body      []byte
		setupMock func(
			sessionSvc *svcMocks.SessionService,
			speechSvc *svcMocks.SpeechProxyService,
			quoteSvc *svcMocks.QuoteService,
		)
		wantErr    bool
		errCode    int
		wantStatus int
		wantBody   []byte
	}{
		{
			name:      "success with 200",
			sessionID: "speech-s1",
			method:    http.MethodPost,
			body:      []byte(`{}`),
			setupMock: func(sessionSvc *svcMocks.SessionService, speechSvc *svcMocks.SpeechProxyService, quoteSvc *svcMocks.QuoteService) {
				sessionSvc.On("GetSessionBySpeechSessionID", mock.Anything, "speech-s1").Return(sessionRes, (*appErrors.AppError)(nil))
				speechSvc.On("ProxyOffer", mock.Anything, "speech-s1", http.MethodPost, mock.Anything).Return(responseBody, http.StatusOK, (*appErrors.AppError)(nil))
				sessionSvc.On("MarkSessionActive", mock.Anything, "s1").Return((*appErrors.AppError)(nil))
			},
			wantStatus: http.StatusOK,
		},
		{
			name:      "empty session id",
			sessionID: "",
			method:    http.MethodPost,
			body:      []byte(`{}`),
			setupMock: func(sessionSvc *svcMocks.SessionService, speechSvc *svcMocks.SpeechProxyService, quoteSvc *svcMocks.QuoteService) {
			},
			wantErr: true,
			errCode: http.StatusBadRequest,
		},
		{
			name:      "get session by speech id error",
			sessionID: "speech-s1",
			method:    http.MethodPost,
			body:      []byte(`{}`),
			setupMock: func(sessionSvc *svcMocks.SessionService, speechSvc *svcMocks.SpeechProxyService, quoteSvc *svcMocks.QuoteService) {
				sessionSvc.On("GetSessionBySpeechSessionID", mock.Anything, "speech-s1").Return((*res.SessionRes)(nil), appErrors.Internal("not found"))
			},
			wantErr: true,
			errCode: http.StatusInternalServerError,
		},
		{
			name:      "speech proxy error releases quota",
			sessionID: "speech-s1",
			method:    http.MethodPost,
			body:      []byte(`{}`),
			setupMock: func(sessionSvc *svcMocks.SessionService, speechSvc *svcMocks.SpeechProxyService, quoteSvc *svcMocks.QuoteService) {
				sessionSvc.On("GetSessionBySpeechSessionID", mock.Anything, "speech-s1").Return(sessionRes, (*appErrors.AppError)(nil))
				speechSvc.On("ProxyOffer", mock.Anything, "speech-s1", http.MethodPost, mock.Anything).Return([]byte{}, 0, appErrors.Internal("proxy error"))
				sessionSvc.On("MarkSessionFailed", mock.Anything, "s1").Return((*appErrors.AppError)(nil))
				quoteSvc.On("Release", mock.Anything, "user-1", int64(300), int64(0), int64(3600)).Return(nil)
				sessionSvc.On("MarkQuotaReleased", mock.Anything, "s1").Return((*appErrors.AppError)(nil))
			},
			wantErr: true,
			errCode: http.StatusInternalServerError,
		},
		{
			name:      "non-2xx status releases quota",
			sessionID: "speech-s1",
			method:    http.MethodPost,
			body:      []byte(`{}`),
			setupMock: func(sessionSvc *svcMocks.SessionService, speechSvc *svcMocks.SpeechProxyService, quoteSvc *svcMocks.QuoteService) {
				sessionSvc.On("GetSessionBySpeechSessionID", mock.Anything, "speech-s1").Return(sessionRes, (*appErrors.AppError)(nil))
				speechSvc.On("ProxyOffer", mock.Anything, "speech-s1", http.MethodPost, mock.Anything).Return(nonSuccessBody, http.StatusBadGateway, (*appErrors.AppError)(nil))
				sessionSvc.On("MarkSessionFailed", mock.Anything, "s1").Return((*appErrors.AppError)(nil))
				quoteSvc.On("Release", mock.Anything, "user-1", int64(300), int64(0), int64(3600)).Return(nil)
				sessionSvc.On("MarkQuotaReleased", mock.Anything, "s1").Return((*appErrors.AppError)(nil))
			},
			wantStatus: http.StatusBadGateway,
			wantBody:    nonSuccessBody,
		},
		{
			name:      "mark active error after success",
			sessionID: "speech-s1",
			method:    http.MethodPost,
			body:      []byte(`{}`),
			setupMock: func(sessionSvc *svcMocks.SessionService, speechSvc *svcMocks.SpeechProxyService, quoteSvc *svcMocks.QuoteService) {
				sessionSvc.On("GetSessionBySpeechSessionID", mock.Anything, "speech-s1").Return(sessionRes, (*appErrors.AppError)(nil))
				speechSvc.On("ProxyOffer", mock.Anything, "speech-s1", http.MethodPost, mock.Anything).Return(responseBody, http.StatusOK, (*appErrors.AppError)(nil))
				sessionSvc.On("MarkSessionActive", mock.Anything, "s1").Return(appErrors.Internal("mark active error"))
			},
			wantErr: true,
			errCode: http.StatusInternalServerError,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			configSvc := new(svcMocks.GlobalConfigService)
			sessionSvc := new(svcMocks.SessionService)
			speechSvc := new(svcMocks.SpeechProxyService)
			quoteSvc := new(svcMocks.QuoteService)

			tt.setupMock(sessionSvc, speechSvc, quoteSvc)

			svc := services.NewModelGatewayService(configSvc, sessionSvc, speechSvc, quoteSvc)
			ctx := context.Background()

			respBody, statusCode, appErr := svc.ProxyOffer(ctx, tt.sessionID, tt.method, tt.body)

			if tt.wantErr {
				require.NotNil(t, appErr)
				assert.Equal(t, tt.errCode, appErr.Code)
				return
			}
			require.Nil(t, appErr)
			assert.Equal(t, tt.wantStatus, statusCode)
			if tt.wantBody != nil {
				assert.Equal(t, tt.wantBody, respBody)
			} else {
				assert.Equal(t, responseBody, respBody)
			}
		})
	}
}

func TestModelGatewayService_CloseSession(t *testing.T) {
	activeSession := &res.SessionRes{ID: "s1", Status: "active", UserID: "user-1", ReservedAmount: 300, DailyQuota: 3600, QuotaReleased: false}

	tests := []struct {
		name   string
		req    *req.CloseSessionReq
		setupMock func(
			sessionSvc *svcMocks.SessionService,
			quoteSvc *svcMocks.QuoteService,
		)
		wantErr bool
		errCode int
	}{
		{
			name: "success",
			req:  &req.CloseSessionReq{SessionID: "s1", ActualUsage: 60},
			setupMock: func(sessionSvc *svcMocks.SessionService, quoteSvc *svcMocks.QuoteService) {
				sessionSvc.On("GetSessionInternal", mock.Anything, "s1").Return(activeSession, (*appErrors.AppError)(nil))
				quoteSvc.On("Release", mock.Anything, "user-1", int64(300), int64(60), int64(3600)).Return(nil)
				sessionSvc.On("MarkQuotaReleased", mock.Anything, "s1").Return((*appErrors.AppError)(nil))
				sessionSvc.On("MarkSessionInactive", mock.Anything, "s1").Return((*appErrors.AppError)(nil))
			},
		},
		{
			name: "nil req body",
			req:  nil,
			setupMock: func(sessionSvc *svcMocks.SessionService, quoteSvc *svcMocks.QuoteService) {
			},
			wantErr: true,
			errCode: http.StatusBadRequest,
		},
		{
			name: "empty session id",
			req:  &req.CloseSessionReq{SessionID: "", ActualUsage: 60},
			setupMock: func(sessionSvc *svcMocks.SessionService, quoteSvc *svcMocks.QuoteService) {
			},
			wantErr: true,
			errCode: http.StatusBadRequest,
		},
		{
			name: "negative actualUsage",
			req:  &req.CloseSessionReq{SessionID: "s1", ActualUsage: -1},
			setupMock: func(sessionSvc *svcMocks.SessionService, quoteSvc *svcMocks.QuoteService) {
			},
			wantErr: true,
			errCode: http.StatusBadRequest,
		},
		{
			name: "session get error",
			req:  &req.CloseSessionReq{SessionID: "s1", ActualUsage: 60},
			setupMock: func(sessionSvc *svcMocks.SessionService, quoteSvc *svcMocks.QuoteService) {
				sessionSvc.On("GetSessionInternal", mock.Anything, "s1").Return((*res.SessionRes)(nil), appErrors.Internal("not found"))
			},
			wantErr: true,
			errCode: http.StatusInternalServerError,
		},
		{
			name: "session already inactive",
			req:  &req.CloseSessionReq{SessionID: "s1", ActualUsage: 60},
			setupMock: func(sessionSvc *svcMocks.SessionService, quoteSvc *svcMocks.QuoteService) {
				sessionSvc.On("GetSessionInternal", mock.Anything, "s1").Return(&res.SessionRes{ID: "s1", Status: "inactive", UserID: "user-1", ReservedAmount: 300, DailyQuota: 3600, QuotaReleased: true}, (*appErrors.AppError)(nil))
			},
			wantErr: true,
			errCode: http.StatusBadRequest,
		},
		{
			name: "pending session succeeds with actualUsage 0",
			req:  &req.CloseSessionReq{SessionID: "s1", ActualUsage: 0},
			setupMock: func(sessionSvc *svcMocks.SessionService, quoteSvc *svcMocks.QuoteService) {
				sessionSvc.On("GetSessionInternal", mock.Anything, "s1").Return(&res.SessionRes{ID: "s1", Status: "pending", UserID: "user-1", ReservedAmount: 300, DailyQuota: 3600, QuotaReleased: false}, (*appErrors.AppError)(nil))
				quoteSvc.On("Release", mock.Anything, "user-1", int64(300), int64(0), int64(3600)).Return(nil)
				sessionSvc.On("MarkQuotaReleased", mock.Anything, "s1").Return((*appErrors.AppError)(nil))
				sessionSvc.On("MarkSessionInactive", mock.Anything, "s1").Return((*appErrors.AppError)(nil))
			},
		},
		{
			name: "failed session succeeds with actualUsage 0",
			req:  &req.CloseSessionReq{SessionID: "s1", ActualUsage: 0},
			setupMock: func(sessionSvc *svcMocks.SessionService, quoteSvc *svcMocks.QuoteService) {
				sessionSvc.On("GetSessionInternal", mock.Anything, "s1").Return(&res.SessionRes{ID: "s1", Status: "failed", UserID: "user-1", ReservedAmount: 300, DailyQuota: 3600, QuotaReleased: false}, (*appErrors.AppError)(nil))
				quoteSvc.On("Release", mock.Anything, "user-1", int64(300), int64(0), int64(3600)).Return(nil)
				sessionSvc.On("MarkQuotaReleased", mock.Anything, "s1").Return((*appErrors.AppError)(nil))
				sessionSvc.On("MarkSessionInactive", mock.Anything, "s1").Return((*appErrors.AppError)(nil))
			},
		},
		{
			name: "quota already released skips release",
			req:  &req.CloseSessionReq{SessionID: "s1", ActualUsage: 60},
			setupMock: func(sessionSvc *svcMocks.SessionService, quoteSvc *svcMocks.QuoteService) {
				sessionSvc.On("GetSessionInternal", mock.Anything, "s1").Return(&res.SessionRes{ID: "s1", Status: "active", UserID: "user-1", ReservedAmount: 300, DailyQuota: 3600, QuotaReleased: true}, (*appErrors.AppError)(nil))
				sessionSvc.On("MarkSessionInactive", mock.Anything, "s1").Return((*appErrors.AppError)(nil))
			},
		},
		{
			name: "actualUsage exceeds reservedAmount gets clamped",
			req:  &req.CloseSessionReq{SessionID: "s1", ActualUsage: 500},
			setupMock: func(sessionSvc *svcMocks.SessionService, quoteSvc *svcMocks.QuoteService) {
				sessionSvc.On("GetSessionInternal", mock.Anything, "s1").Return(activeSession, (*appErrors.AppError)(nil))
				quoteSvc.On("Release", mock.Anything, "user-1", int64(300), int64(300), int64(3600)).Return(nil)
				sessionSvc.On("MarkQuotaReleased", mock.Anything, "s1").Return((*appErrors.AppError)(nil))
				sessionSvc.On("MarkSessionInactive", mock.Anything, "s1").Return((*appErrors.AppError)(nil))
			},
		},
		{
			name: "quote release error",
			req:  &req.CloseSessionReq{SessionID: "s1", ActualUsage: 60},
			setupMock: func(sessionSvc *svcMocks.SessionService, quoteSvc *svcMocks.QuoteService) {
				sessionSvc.On("GetSessionInternal", mock.Anything, "s1").Return(activeSession, (*appErrors.AppError)(nil))
				quoteSvc.On("Release", mock.Anything, "user-1", int64(300), int64(60), int64(3600)).Return(appErrors.Internal("release error"))
			},
			wantErr: true,
			errCode: http.StatusInternalServerError,
		},
		{
			name: "mark session inactive error",
			req:  &req.CloseSessionReq{SessionID: "s1", ActualUsage: 60},
			setupMock: func(sessionSvc *svcMocks.SessionService, quoteSvc *svcMocks.QuoteService) {
				sessionSvc.On("GetSessionInternal", mock.Anything, "s1").Return(activeSession, (*appErrors.AppError)(nil))
				quoteSvc.On("Release", mock.Anything, "user-1", int64(300), int64(60), int64(3600)).Return(nil)
				sessionSvc.On("MarkQuotaReleased", mock.Anything, "s1").Return((*appErrors.AppError)(nil))
				sessionSvc.On("MarkSessionInactive", mock.Anything, "s1").Return(appErrors.Internal("inactive error"))
			},
			wantErr: true,
			errCode: http.StatusInternalServerError,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			configSvc := new(svcMocks.GlobalConfigService)
			sessionSvc := new(svcMocks.SessionService)
			speechSvc := new(svcMocks.SpeechProxyService)
			quoteSvc := new(svcMocks.QuoteService)

			tt.setupMock(sessionSvc, quoteSvc)

			svc := services.NewModelGatewayService(configSvc, sessionSvc, speechSvc, quoteSvc)
			ctx := context.Background()

			appErr := svc.CloseSession(ctx, tt.req)

			if tt.wantErr {
				require.NotNil(t, appErr)
				assert.Equal(t, tt.errCode, appErr.Code)
				return
			}
			require.Nil(t, appErr)
		})
	}
}

func TestModelGatewayService_CloseSession_ActualUsageClampedJSON(t *testing.T) {
	activeSession := &res.SessionRes{ID: "s1", Status: "active", UserID: "user-1", ReservedAmount: 200, DailyQuota: 3600, QuotaReleased: false}

	sessionSvc := new(svcMocks.SessionService)
	speechSvc := new(svcMocks.SpeechProxyService)
	configSvc := new(svcMocks.GlobalConfigService)
	quoteSvc := new(svcMocks.QuoteService)

	sessionSvc.On("GetSessionInternal", mock.Anything, "s1").Return(activeSession, (*appErrors.AppError)(nil))
	quoteSvc.On("Release", mock.Anything, "user-1", int64(200), int64(200), int64(3600)).Return(nil)
	sessionSvc.On("MarkQuotaReleased", mock.Anything, "s1").Return((*appErrors.AppError)(nil))
	sessionSvc.On("MarkSessionInactive", mock.Anything, "s1").Return((*appErrors.AppError)(nil))

	svc := services.NewModelGatewayService(configSvc, sessionSvc, speechSvc, quoteSvc)
	ctx := context.Background()

	appErr := svc.CloseSession(ctx, &req.CloseSessionReq{SessionID: "s1", ActualUsage: 500})
	require.Nil(t, appErr)
}

func TestModelGatewayService_CreateSession_PanicOnCleanup(t *testing.T) {
	defaultConfig := &dtos.GlobalConfig{
		Config: dtos.ConfigPayload{
			Limits: dtos.LimitsConfig{
				Session: dtos.SessionLimitConfig{MaxSessionLockTTL: 60},
				User:    dtos.UserLimitConfig{DailyVoiceSeconds: 3600},
			},
		},
	}
	sessionRes := &res.SessionRes{ID: "s1", UserID: "user-1"}

	configSvc := new(svcMocks.GlobalConfigService)
	sessionSvc := new(svcMocks.SessionService)
	speechSvc := new(svcMocks.SpeechProxyService)
	quoteSvc := new(svcMocks.QuoteService)

	configSvc.On("Get", mock.Anything).Return(defaultConfig, (*appErrors.AppError)(nil))
	sessionSvc.On("FindActiveByUserID", mock.Anything, "user-1").Return(nil, (*appErrors.AppError)(nil))
	sessionSvc.On("FindStaleSessions", mock.Anything, "user-1", int64(60)).Return([]*res.SessionRes{}, (*appErrors.AppError)(nil))
	quoteSvc.On("AcquireSessionLock", mock.Anything, "user-1", 60*time.Second).Return("lock-val", nil)
	quoteSvc.On("ReleaseSessionLock", mock.Anything, "user-1", "lock-val").Return(nil)
	quoteSvc.On("ReserveQuota", mock.Anything, "user-1", int64(3600)).Return(int64(300), func(bool) {}, (*appErrors.AppError)(nil))
	sessionSvc.On("CreateSession", mock.Anything).Return(sessionRes, (*appErrors.AppError)(nil))
	sessionSvc.On("SetReservation", mock.Anything, "s1", int64(300), int64(3600)).Return((*appErrors.AppError)(nil))
	speechSvc.On("StartConnection", mock.Anything, mock.Anything).Return((*res.WebRTCConnectionRes)(nil), appErrors.Internal("speech error"))
	sessionSvc.On("MarkSessionFailed", mock.Anything, "s1").Return((*appErrors.AppError)(nil))
	quoteSvc.On("Release", mock.Anything, "user-1", int64(300), int64(0), int64(3600)).Return(nil)
	sessionSvc.On("MarkQuotaReleased", mock.Anything, "s1").Return((*appErrors.AppError)(nil))

	svc := services.NewModelGatewayService(configSvc, sessionSvc, speechSvc, quoteSvc)
	ctx := setupSessionCtx("user-1")

	result, appErr := svc.CreateSession(ctx)
	require.NotNil(t, appErr)
	require.Nil(t, result)
}

func TestModelGatewayService_CreateSession_CleanupOnSetReservationError(t *testing.T) {
	defaultConfig := &dtos.GlobalConfig{
		Config: dtos.ConfigPayload{
			Limits: dtos.LimitsConfig{
				Session: dtos.SessionLimitConfig{MaxSessionLockTTL: 60},
				User:    dtos.UserLimitConfig{DailyVoiceSeconds: 3600},
			},
		},
	}
	sessionRes := &res.SessionRes{ID: "s1", UserID: "user-1"}

	var cleanupCalled bool
	cleanup := func(commit bool) {
		cleanupCalled = true
		assert.False(t, commit)
	}

	configSvc := new(svcMocks.GlobalConfigService)
	sessionSvc := new(svcMocks.SessionService)
	speechSvc := new(svcMocks.SpeechProxyService)
	quoteSvc := new(svcMocks.QuoteService)

	sessionSvc.On("CreateSession", mock.Anything).Return(sessionRes, (*appErrors.AppError)(nil))
	configSvc.On("Get", mock.Anything).Return(defaultConfig, (*appErrors.AppError)(nil))
	sessionSvc.On("FindActiveByUserID", mock.Anything, "user-1").Return(nil, (*appErrors.AppError)(nil))
	sessionSvc.On("FindStaleSessions", mock.Anything, "user-1", int64(60)).Return([]*res.SessionRes{}, (*appErrors.AppError)(nil))
	quoteSvc.On("AcquireSessionLock", mock.Anything, "user-1", 60*time.Second).Return("lock-val", nil)
	quoteSvc.On("ReleaseSessionLock", mock.Anything, "user-1", "lock-val").Return(nil)
	quoteSvc.On("ReserveQuota", mock.Anything, "user-1", int64(3600)).Return(int64(300), cleanup, (*appErrors.AppError)(nil))
	sessionSvc.On("SetReservation", mock.Anything, "s1", int64(300), int64(3600)).Return(appErrors.Internal("set reservation error"))

	svc := services.NewModelGatewayService(configSvc, sessionSvc, speechSvc, quoteSvc)
	ctx := setupSessionCtx("user-1")

	result, appErr := svc.CreateSession(ctx)
	require.NotNil(t, appErr)
	require.Nil(t, result)
	assert.True(t, cleanupCalled)
}

func TestModelGatewayService_ProxyOffer_NonJSONBody(t *testing.T) {
	sessionRes := &res.SessionRes{ID: "s1", UserID: "user-1", Status: "active", ReservedAmount: 300, DailyQuota: 3600, QuotaReleased: false}
	rawBody := []byte(`raw body`)

	sessionSvc := new(svcMocks.SessionService)
	speechSvc := new(svcMocks.SpeechProxyService)
	configSvc := new(svcMocks.GlobalConfigService)
	quoteSvc := new(svcMocks.QuoteService)

	sessionSvc.On("GetSessionBySpeechSessionID", mock.Anything, "speech-s1").Return(sessionRes, (*appErrors.AppError)(nil))
	speechSvc.On("ProxyOffer", mock.Anything, "speech-s1", http.MethodPost, rawBody).Return([]byte("resp"), http.StatusOK, (*appErrors.AppError)(nil))
	sessionSvc.On("MarkSessionActive", mock.Anything, "s1").Return((*appErrors.AppError)(nil))

	svc := services.NewModelGatewayService(configSvc, sessionSvc, speechSvc, quoteSvc)
	ctx := context.Background()

	respBody, statusCode, appErr := svc.ProxyOffer(ctx, "speech-s1", http.MethodPost, rawBody)
	require.Nil(t, appErr)
	assert.Equal(t, http.StatusOK, statusCode)
	assert.Equal(t, []byte("resp"), respBody)
}

func TestModelGatewayService_ProxyOffer_NonSuccessWithFailedMark(t *testing.T) {
	sessionRes := &res.SessionRes{ID: "s1", UserID: "user-1", Status: "active", ReservedAmount: 300, DailyQuota: 3600, QuotaReleased: false}
	sessionSvc := new(svcMocks.SessionService)
	speechSvc := new(svcMocks.SpeechProxyService)
	configSvc := new(svcMocks.GlobalConfigService)
	quoteSvc := new(svcMocks.QuoteService)

	sessionSvc.On("GetSessionBySpeechSessionID", mock.Anything, "speech-s1").Return(sessionRes, (*appErrors.AppError)(nil))
	speechSvc.On("ProxyOffer", mock.Anything, "speech-s1", http.MethodPost, mock.Anything).Return([]byte("err"), http.StatusBadRequest, (*appErrors.AppError)(nil))
	sessionSvc.On("MarkSessionFailed", mock.Anything, "s1").Return((*appErrors.AppError)(nil))
	quoteSvc.On("Release", mock.Anything, "user-1", int64(300), int64(0), int64(3600)).Return(nil)
	sessionSvc.On("MarkQuotaReleased", mock.Anything, "s1").Return((*appErrors.AppError)(nil))

	svc := services.NewModelGatewayService(configSvc, sessionSvc, speechSvc, quoteSvc)
	ctx := context.Background()

	respBody, statusCode, appErr := svc.ProxyOffer(ctx, "speech-s1", http.MethodPost, []byte(`{}`))
	require.Nil(t, appErr)
	assert.Equal(t, http.StatusBadRequest, statusCode)
	assert.Equal(t, []byte("err"), respBody)
}

func TestModelGatewayService_ProxyOffer_2xxStatusCodes(t *testing.T) {
	sessionRes := &res.SessionRes{ID: "s1", Status: "active", ReservedAmount: 300, DailyQuota: 3600, QuotaReleased: false}

	tests := []struct {
		name       string
		statusCode int
	}{
		{"201 Created", http.StatusCreated},
		{"202 Accepted", http.StatusAccepted},
		{"204 No Content", http.StatusNoContent},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			sessionSvc := new(svcMocks.SessionService)
			speechSvc := new(svcMocks.SpeechProxyService)
			configSvc := new(svcMocks.GlobalConfigService)
			quoteSvc := new(svcMocks.QuoteService)

			sessionSvc.On("GetSessionBySpeechSessionID", mock.Anything, "speech-s1").Return(sessionRes, (*appErrors.AppError)(nil))
			speechSvc.On("ProxyOffer", mock.Anything, "speech-s1", http.MethodPost, mock.Anything).Return([]byte("body"), tt.statusCode, (*appErrors.AppError)(nil))
			sessionSvc.On("MarkSessionActive", mock.Anything, "s1").Return((*appErrors.AppError)(nil))

			svc := services.NewModelGatewayService(configSvc, sessionSvc, speechSvc, quoteSvc)
			ctx := context.Background()

			respBody, statusCode, appErr := svc.ProxyOffer(ctx, "speech-s1", http.MethodPost, []byte(`{}`))
			require.Nil(t, appErr)
			assert.Equal(t, tt.statusCode, statusCode)
			assert.Equal(t, []byte("body"), respBody)
		})
	}
}

func TestModelGatewayService_ProxyOffer_Non200StatusCodes(t *testing.T) {
	sessionRes := &res.SessionRes{ID: "s1", UserID: "user-1", Status: "active", ReservedAmount: 300, DailyQuota: 3600, QuotaReleased: false}

	tests := []struct {
		name       string
		statusCode int
	}{
		{"300 Multiple Choices", http.StatusMultipleChoices},
		{"301 Moved Permanently", http.StatusMovedPermanently},
		{"400 Bad Request", http.StatusBadRequest},
		{"502 Bad Gateway", http.StatusBadGateway},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			sessionSvc := new(svcMocks.SessionService)
			speechSvc := new(svcMocks.SpeechProxyService)
			configSvc := new(svcMocks.GlobalConfigService)
			quoteSvc := new(svcMocks.QuoteService)

			sessionSvc.On("GetSessionBySpeechSessionID", mock.Anything, "speech-s1").Return(sessionRes, (*appErrors.AppError)(nil))
			speechSvc.On("ProxyOffer", mock.Anything, "speech-s1", http.MethodPost, mock.Anything).Return([]byte("body"), tt.statusCode, (*appErrors.AppError)(nil))
			sessionSvc.On("MarkSessionFailed", mock.Anything, "s1").Return((*appErrors.AppError)(nil))
			quoteSvc.On("Release", mock.Anything, "user-1", int64(300), int64(0), int64(3600)).Return(nil)
			sessionSvc.On("MarkQuotaReleased", mock.Anything, "s1").Return((*appErrors.AppError)(nil))

			svc := services.NewModelGatewayService(configSvc, sessionSvc, speechSvc, quoteSvc)
			ctx := context.Background()

			respBody, statusCode, appErr := svc.ProxyOffer(ctx, "speech-s1", http.MethodPost, []byte(`{}`))
			require.Nil(t, appErr)
			assert.Equal(t, tt.statusCode, statusCode)
			assert.Equal(t, []byte("body"), respBody)
		})
	}
}

func TestModelGatewayService_CloseSession_PendingSessionSucceeds(t *testing.T) {
	pendingSession := &res.SessionRes{ID: "s1", Status: "pending", UserID: "user-1", ReservedAmount: 300, DailyQuota: 3600, QuotaReleased: false}

	sessionSvc := new(svcMocks.SessionService)
	configSvc := new(svcMocks.GlobalConfigService)
	speechSvc := new(svcMocks.SpeechProxyService)
	quoteSvc := new(svcMocks.QuoteService)

	sessionSvc.On("GetSessionInternal", mock.Anything, "s1").Return(pendingSession, (*appErrors.AppError)(nil))
	quoteSvc.On("Release", mock.Anything, "user-1", int64(300), int64(0), int64(3600)).Return(nil)
	sessionSvc.On("MarkQuotaReleased", mock.Anything, "s1").Return((*appErrors.AppError)(nil))
	sessionSvc.On("MarkSessionInactive", mock.Anything, "s1").Return((*appErrors.AppError)(nil))

	svc := services.NewModelGatewayService(configSvc, sessionSvc, speechSvc, quoteSvc)
	ctx := context.Background()

	appErr := svc.CloseSession(ctx, &req.CloseSessionReq{SessionID: "s1", ActualUsage: 0})
	require.Nil(t, appErr)
}

func TestModelGatewayService_CloseSession_AlreadyReleased(t *testing.T) {
	releasedSession := &res.SessionRes{ID: "s1", Status: "active", UserID: "user-1", ReservedAmount: 300, DailyQuota: 3600, QuotaReleased: true}

	sessionSvc := new(svcMocks.SessionService)
	configSvc := new(svcMocks.GlobalConfigService)
	speechSvc := new(svcMocks.SpeechProxyService)
	quoteSvc := new(svcMocks.QuoteService)

	sessionSvc.On("GetSessionInternal", mock.Anything, "s1").Return(releasedSession, (*appErrors.AppError)(nil))
	sessionSvc.On("MarkSessionInactive", mock.Anything, "s1").Return((*appErrors.AppError)(nil))

	svc := services.NewModelGatewayService(configSvc, sessionSvc, speechSvc, quoteSvc)
	ctx := context.Background()

	appErr := svc.CloseSession(ctx, &req.CloseSessionReq{SessionID: "s1", ActualUsage: 60})
	require.Nil(t, appErr)
}

func TestModelGatewayService_ProxyOffer_ContextPropagation(t *testing.T) {
	sessionRes := &res.SessionRes{ID: "s1", Status: "active", ReservedAmount: 300, DailyQuota: 3600, QuotaReleased: false}
	type ctxKey string
	testKey := ctxKey("test-key")
	ctx := context.WithValue(context.Background(), testKey, "test-value")

	sessionSvc := new(svcMocks.SessionService)
	speechSvc := new(svcMocks.SpeechProxyService)
	configSvc := new(svcMocks.GlobalConfigService)
	quoteSvc := new(svcMocks.QuoteService)

	sessionSvc.On("GetSessionBySpeechSessionID", mock.Anything, "speech-s1").Return(sessionRes, (*appErrors.AppError)(nil))
	speechSvc.On("ProxyOffer", mock.Anything, "speech-s1", http.MethodPost, mock.Anything).Return([]byte("ok"), http.StatusOK, (*appErrors.AppError)(nil))
	sessionSvc.On("MarkSessionActive", mock.Anything, "s1").Return((*appErrors.AppError)(nil))

	svc := services.NewModelGatewayService(configSvc, sessionSvc, speechSvc, quoteSvc)

	respBody, statusCode, appErr := svc.ProxyOffer(ctx, "speech-s1", http.MethodPost, []byte(`{}`))
	require.Nil(t, appErr)
	assert.Equal(t, http.StatusOK, statusCode)
	assert.Equal(t, []byte("ok"), respBody)
}

func TestModelGatewayService_CreateSession_JSONSerializable(t *testing.T) {
	defaultConfig := &dtos.GlobalConfig{
		Config: dtos.ConfigPayload{
			Limits: dtos.LimitsConfig{
				Session: dtos.SessionLimitConfig{MaxSessionLockTTL: 60},
				User:    dtos.UserLimitConfig{DailyVoiceSeconds: 3600},
			},
		},
	}
	sessionRes := &res.SessionRes{ID: "s1", UserID: "user-1"}
	webrtcRes := &res.WebRTCConnectionRes{
		SessionID: "speech-s1",
		IceConfig: &res.IceConfig{
			IceServers: []res.IceServer{
				{Urls: []string{"stun:stun.l.google.com:19302"}},
			},
		},
	}

	configSvc := new(svcMocks.GlobalConfigService)
	sessionSvc := new(svcMocks.SessionService)
	speechSvc := new(svcMocks.SpeechProxyService)
	quoteSvc := new(svcMocks.QuoteService)

	configSvc.On("Get", mock.Anything).Return(defaultConfig, (*appErrors.AppError)(nil))
	sessionSvc.On("FindActiveByUserID", mock.Anything, "user-1").Return(nil, (*appErrors.AppError)(nil))
	sessionSvc.On("FindStaleSessions", mock.Anything, "user-1", int64(60)).Return([]*res.SessionRes{}, (*appErrors.AppError)(nil))
	quoteSvc.On("AcquireSessionLock", mock.Anything, "user-1", 60*time.Second).Return("lock-val", nil)
	quoteSvc.On("ReleaseSessionLock", mock.Anything, "user-1", "lock-val").Return(nil)
	quoteSvc.On("ReserveQuota", mock.Anything, "user-1", int64(3600)).Return(int64(100), func(bool) {}, (*appErrors.AppError)(nil))
	sessionSvc.On("CreateSession", mock.Anything).Return(sessionRes, (*appErrors.AppError)(nil))
	sessionSvc.On("SetReservation", mock.Anything, "s1", int64(100), int64(3600)).Return((*appErrors.AppError)(nil))
	speechSvc.On("StartConnection", mock.Anything, mock.Anything).Return(webrtcRes, (*appErrors.AppError)(nil))
	sessionSvc.On("SetSpeechSessionID", mock.Anything, "s1", "speech-s1").Return((*appErrors.AppError)(nil))

	svc := services.NewModelGatewayService(configSvc, sessionSvc, speechSvc, quoteSvc)
	ctx := setupSessionCtx("user-1")

	result, appErr := svc.CreateSession(ctx)
	require.Nil(t, appErr)

	bytes, err := json.Marshal(result)
	require.NoError(t, err)
	assert.Contains(t, string(bytes), "speech-s1")
	assert.Contains(t, string(bytes), "stun:stun.l.google.com:19302")
}

func TestModelGatewayService_CreateSession_ConflictWhenActiveSession(t *testing.T) {
	configSvc := new(svcMocks.GlobalConfigService)
	sessionSvc := new(svcMocks.SessionService)
	speechSvc := new(svcMocks.SpeechProxyService)
	quoteSvc := new(svcMocks.QuoteService)

	sessionSvc.On("CreateSession", mock.Anything).Return(&res.SessionRes{ID: "s1", UserID: "user-1"}, (*appErrors.AppError)(nil))
	configSvc.On("Get", mock.Anything).Return(&dtos.GlobalConfig{}, (*appErrors.AppError)(nil))
	sessionSvc.On("FindActiveByUserID", mock.Anything, "user-1").Return(&res.SessionRes{ID: "existing-active", UserID: "user-1", Status: "active"}, (*appErrors.AppError)(nil))

	svc := services.NewModelGatewayService(configSvc, sessionSvc, speechSvc, quoteSvc)
	ctx := setupSessionCtx("user-1")

	_, appErr := svc.CreateSession(ctx)
	require.NotNil(t, appErr)
	assert.Equal(t, http.StatusConflict, appErr.Code)
}

func TestModelGatewayService_ResumeSession_NotOwner(t *testing.T) {
	configSvc := new(svcMocks.GlobalConfigService)
	sessionSvc := new(svcMocks.SessionService)
	speechSvc := new(svcMocks.SpeechProxyService)
	quoteSvc := new(svcMocks.QuoteService)

	sessionSvc.On("GetSessionInternal", mock.Anything, "session-1").Return(&res.SessionRes{ID: "session-1", UserID: "user-2", Status: "inactive"}, (*appErrors.AppError)(nil))

	svc := services.NewModelGatewayService(configSvc, sessionSvc, speechSvc, quoteSvc)
	ctx := setupSessionCtx("user-1")

	_, appErr := svc.ResumeSession(ctx, "session-1")
	require.NotNil(t, appErr)
	assert.Equal(t, http.StatusForbidden, appErr.Code)
}

func TestModelGatewayService_ResumeSession_NotResumable(t *testing.T) {
	configSvc := new(svcMocks.GlobalConfigService)
	sessionSvc := new(svcMocks.SessionService)
	speechSvc := new(svcMocks.SpeechProxyService)
	quoteSvc := new(svcMocks.QuoteService)

	sessionSvc.On("GetSessionInternal", mock.Anything, "session-1").Return(&res.SessionRes{ID: "session-1", UserID: "user-1", Status: "active"}, (*appErrors.AppError)(nil))

	svc := services.NewModelGatewayService(configSvc, sessionSvc, speechSvc, quoteSvc)
	ctx := setupSessionCtx("user-1")

	_, appErr := svc.ResumeSession(ctx, "session-1")
	require.NotNil(t, appErr)
	assert.Equal(t, http.StatusBadRequest, appErr.Code)
}

func TestModelGatewayService_ResumeSession_ConflictWhenActiveSession(t *testing.T) {
	configSvc := new(svcMocks.GlobalConfigService)
	sessionSvc := new(svcMocks.SessionService)
	speechSvc := new(svcMocks.SpeechProxyService)
	quoteSvc := new(svcMocks.QuoteService)

	sessionSvc.On("GetSessionInternal", mock.Anything, "session-1").Return(&res.SessionRes{ID: "session-1", UserID: "user-1", Status: "inactive"}, (*appErrors.AppError)(nil))
	sessionSvc.On("UpdateStatus", mock.Anything, "session-1", enums.SessionStatusPending).Return((*appErrors.AppError)(nil))
	configSvc.On("Get", mock.Anything).Return(&dtos.GlobalConfig{}, (*appErrors.AppError)(nil))
	sessionSvc.On("FindActiveByUserID", mock.Anything, "user-1").Return(&res.SessionRes{ID: "session-2", UserID: "user-1", Status: "active"}, (*appErrors.AppError)(nil))

	svc := services.NewModelGatewayService(configSvc, sessionSvc, speechSvc, quoteSvc)
	ctx := setupSessionCtx("user-1")

	_, appErr := svc.ResumeSession(ctx, "session-1")
	require.NotNil(t, appErr)
	assert.Equal(t, http.StatusConflict, appErr.Code)
}

func TestModelGatewayService_ListSessions(t *testing.T) {
	configSvc := new(svcMocks.GlobalConfigService)
	sessionSvc := new(svcMocks.SessionService)
	speechSvc := new(svcMocks.SpeechProxyService)
	quoteSvc := new(svcMocks.QuoteService)

	expectedList := []*res.SessionListItemRes{
		{ID: "s1", Status: "inactive"},
		{ID: "s2", Status: "inactive"},
	}
	sessionSvc.On("FindResumableByUserID", mock.Anything, "user-1").Return(expectedList, (*appErrors.AppError)(nil))

	svc := services.NewModelGatewayService(configSvc, sessionSvc, speechSvc, quoteSvc)
	ctx := setupSessionCtx("user-1")

	result, appErr := svc.ListSessions(ctx)
	require.Nil(t, appErr)
	assert.Len(t, result, 2)
}
