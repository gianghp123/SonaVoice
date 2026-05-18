package tests

import (
	"context"
	"encoding/json"
	"net/http"
	"testing"
	"time"

	appErrors "github.com/gianghp123/SonaVoice/api/internal/core/errors"
	"github.com/gianghp123/SonaVoice/api/internal/core/enums"
	"github.com/gianghp123/SonaVoice/api/internal/database/models"
	repoMocks "github.com/gianghp123/SonaVoice/api/internal/database/repository-interfaces/mocks"
	"github.com/gianghp123/SonaVoice/api/internal/modules/model-gateway/dtos/req"
	"github.com/gianghp123/SonaVoice/api/internal/modules/model-gateway/dtos/res"
	"github.com/gianghp123/SonaVoice/api/internal/modules/model-gateway/services"
	svcMocks "github.com/gianghp123/SonaVoice/api/internal/modules/model-gateway/services/mocks"
	"gorm.io/datatypes"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
)

func setupSessionCtx(userID string) context.Context {
	ctx := context.Background()
	ctx = context.WithValue(ctx, enums.ContextKeyUserID, userID)
	return ctx
}

func TestModelGatewayService_CreateSession(t *testing.T) {
	defaultConfig := &models.GlobalConfig{
		Config: datatypes.JSON(`{"limits":{"session":{"max_session_lockTTL":60},"user":{"daily_voice_seconds":3600}}}`),
	}
	webrtcRes := &res.WebRTCConnectionRes{SessionID: "speech-s1"}

	tests := []struct {
		name      string
		setupMock func(
			configSvc *svcMocks.GlobalConfigService,
			sessionSvc *svcMocks.SessionService,
			speechSvc *svcMocks.SpeechProxyService,
			startConnSvc *svcMocks.StartConnectionService,
		)
		wantErr bool
		errCode int
	}{
		{
			name: "success",
			setupMock: func(configSvc *svcMocks.GlobalConfigService, sessionSvc *svcMocks.SessionService, speechSvc *svcMocks.SpeechProxyService, startConnSvc *svcMocks.StartConnectionService) {
				configSvc.On("Get", mock.Anything).Return(defaultConfig, (*appErrors.AppError)(nil))
				sessionSvc.On("Create", mock.Anything, "user-1").Return(&models.Session{BaseModel: models.BaseModel{ID: "s1"}, UserID: "user-1", Status: enums.SessionStatusPending}, (*appErrors.AppError)(nil))
				startConnSvc.On("Start", mock.Anything, mock.Anything, "user-1", 3600).Return(&res.CreateSessionRes{
					ID:                  "s1",
					MaxDuration:         300,
					WebRTCConnectionRes: webrtcRes,
				}, (*appErrors.AppError)(nil))
			},
		},
		{
			name: "config service get error",
			setupMock: func(configSvc *svcMocks.GlobalConfigService, sessionSvc *svcMocks.SessionService, speechSvc *svcMocks.SpeechProxyService, startConnSvc *svcMocks.StartConnectionService) {
				configSvc.On("Get", mock.Anything).Return((*models.GlobalConfig)(nil), appErrors.Internal("config error"))
			},
			wantErr: true,
			errCode: http.StatusInternalServerError,
		},
		{
			name: "create session conflict error",
			setupMock: func(configSvc *svcMocks.GlobalConfigService, sessionSvc *svcMocks.SessionService, speechSvc *svcMocks.SpeechProxyService, startConnSvc *svcMocks.StartConnectionService) {
				configSvc.On("Get", mock.Anything).Return(defaultConfig, (*appErrors.AppError)(nil))
				sessionSvc.On("Create", mock.Anything, "user-1").Return((*models.Session)(nil), appErrors.Conflict("close current session before starting a new one"))
			},
			wantErr: true,
			errCode: http.StatusConflict,
		},
		{
			name: "start connection error",
			setupMock: func(configSvc *svcMocks.GlobalConfigService, sessionSvc *svcMocks.SessionService, speechSvc *svcMocks.SpeechProxyService, startConnSvc *svcMocks.StartConnectionService) {
				configSvc.On("Get", mock.Anything).Return(defaultConfig, (*appErrors.AppError)(nil))
				sessionSvc.On("Create", mock.Anything, "user-1").Return(&models.Session{BaseModel: models.BaseModel{ID: "s1"}, UserID: "user-1", Status: enums.SessionStatusPending}, (*appErrors.AppError)(nil))
				startConnSvc.On("Start", mock.Anything, mock.Anything, "user-1", 3600).Return((*res.CreateSessionRes)(nil), appErrors.Forbidden("quota exceeded"))
			},
			wantErr: true,
			errCode: http.StatusForbidden,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			configSvc := new(svcMocks.GlobalConfigService)
			sessionSvc := new(svcMocks.SessionService)
			speechSvc := new(svcMocks.SpeechProxyService)
			startConnSvc := new(svcMocks.StartConnectionService)

			tt.setupMock(configSvc, sessionSvc, speechSvc, startConnSvc)

			svc := services.NewModelGatewayService(configSvc, sessionSvc, speechSvc, startConnSvc, nil)
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

func TestModelGatewayService_StartConnection(t *testing.T) {
	defaultConfig := &models.GlobalConfig{
		Config: datatypes.JSON(`{"limits":{"session":{"max_session_lockTTL":60},"user":{"daily_voice_seconds":3600}}}`),
	}
	pendingSession := &models.Session{BaseModel: models.BaseModel{ID: "s1"}, UserID: "user-1", Status: enums.SessionStatusPending}
	webrtcRes := &res.WebRTCConnectionRes{SessionID: "speech-s1"}

	tests := []struct {
		name      string
		setupMock func(
			configSvc *svcMocks.GlobalConfigService,
			sessionSvc *svcMocks.SessionService,
			speechSvc *svcMocks.SpeechProxyService,
			startConnSvc *svcMocks.StartConnectionService,
		)
		wantErr bool
		errCode int
	}{
		{
			name: "success",
			setupMock: func(configSvc *svcMocks.GlobalConfigService, sessionSvc *svcMocks.SessionService, speechSvc *svcMocks.SpeechProxyService, startConnSvc *svcMocks.StartConnectionService) {
				sessionSvc.On("Get", mock.Anything, "s1", "user-1").Return(pendingSession, (*appErrors.AppError)(nil))
				configSvc.On("Get", mock.Anything).Return(defaultConfig, (*appErrors.AppError)(nil))
				startConnSvc.On("Start", mock.Anything, pendingSession, "user-1", 3600).Return(&res.CreateSessionRes{
					ID:                  "s1",
					MaxDuration:         300,
					WebRTCConnectionRes: webrtcRes,
				}, (*appErrors.AppError)(nil))
			},
		},
		{
			name: "session not found",
			setupMock: func(configSvc *svcMocks.GlobalConfigService, sessionSvc *svcMocks.SessionService, speechSvc *svcMocks.SpeechProxyService, startConnSvc *svcMocks.StartConnectionService) {
				sessionSvc.On("Get", mock.Anything, "s1", "user-1").Return((*models.Session)(nil), appErrors.Internal("not found"))
			},
			wantErr: true,
			errCode: http.StatusInternalServerError,
		},
		{
			name: "session not startable (active)",
			setupMock: func(configSvc *svcMocks.GlobalConfigService, sessionSvc *svcMocks.SessionService, speechSvc *svcMocks.SpeechProxyService, startConnSvc *svcMocks.StartConnectionService) {
				activeSession := &models.Session{BaseModel: models.BaseModel{ID: "s1"}, UserID: "user-1", Status: enums.SessionStatusActive}
				sessionSvc.On("Get", mock.Anything, "s1", "user-1").Return(activeSession, (*appErrors.AppError)(nil))
			},
			wantErr: true,
			errCode: http.StatusBadRequest,
		},
		{
			name: "session not startable (inactive)",
			setupMock: func(configSvc *svcMocks.GlobalConfigService, sessionSvc *svcMocks.SessionService, speechSvc *svcMocks.SpeechProxyService, startConnSvc *svcMocks.StartConnectionService) {
				inactiveSession := &models.Session{BaseModel: models.BaseModel{ID: "s1"}, UserID: "user-1", Status: enums.SessionStatusInactive}
				sessionSvc.On("Get", mock.Anything, "s1", "user-1").Return(inactiveSession, (*appErrors.AppError)(nil))
			},
			wantErr: true,
			errCode: http.StatusBadRequest,
		},
		{
			name: "config error",
			setupMock: func(configSvc *svcMocks.GlobalConfigService, sessionSvc *svcMocks.SessionService, speechSvc *svcMocks.SpeechProxyService, startConnSvc *svcMocks.StartConnectionService) {
				sessionSvc.On("Get", mock.Anything, "s1", "user-1").Return(pendingSession, (*appErrors.AppError)(nil))
				configSvc.On("Get", mock.Anything).Return((*models.GlobalConfig)(nil), appErrors.Internal("config error"))
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
			startConnSvc := new(svcMocks.StartConnectionService)

			tt.setupMock(configSvc, sessionSvc, speechSvc, startConnSvc)

			svc := services.NewModelGatewayService(configSvc, sessionSvc, speechSvc, startConnSvc, nil)
			ctx := setupSessionCtx("user-1")

			result, appErr := svc.StartConnection(ctx, "s1")

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
	pendingSession := &models.Session{BaseModel: models.BaseModel{ID: "s1"}, UserID: "user-1", Status: enums.SessionStatusPending, ReservedAmount: 300}
	activeSession := &models.Session{BaseModel: models.BaseModel{ID: "s1"}, UserID: "user-1", Status: enums.SessionStatusActive, ReservedAmount: 300}
	responseBody := []byte(`{"ok":true}`)

	tests := []struct {
		name       string
		sessionID  string
		method     string
		body       []byte
		setupMock func(
			sessionSvc *svcMocks.SessionService,
			speechSvc *svcMocks.SpeechProxyService,
		)
		wantErr    bool
		errCode    int
		wantStatus int
		wantBody   []byte
	}{
		{
			name:      "success with 200 marks pending session active",
			sessionID: "speech-s1",
			method:    http.MethodPost,
			body:      []byte(`{}`),
			setupMock: func(sessionSvc *svcMocks.SessionService, speechSvc *svcMocks.SpeechProxyService) {
				sessionSvc.On("GetBySpeechSessionID", mock.Anything, "speech-s1", "user-1").Return(pendingSession, (*appErrors.AppError)(nil))
				speechSvc.On("ProxyOffer", mock.Anything, "speech-s1", http.MethodPost, mock.Anything).Return(responseBody, http.StatusOK, (*appErrors.AppError)(nil))
				sessionSvc.On("MarkSessionActive", mock.Anything, "s1").Return((*appErrors.AppError)(nil))
			},
			wantStatus: http.StatusOK,
		},
		{
			name:      "success with 200 on already active session does not call MarkSessionActive",
			sessionID: "speech-s1",
			method:    http.MethodPost,
			body:      []byte(`{}`),
			setupMock: func(sessionSvc *svcMocks.SessionService, speechSvc *svcMocks.SpeechProxyService) {
				sessionSvc.On("GetBySpeechSessionID", mock.Anything, "speech-s1", "user-1").Return(activeSession, (*appErrors.AppError)(nil))
				speechSvc.On("ProxyOffer", mock.Anything, "speech-s1", http.MethodPost, mock.Anything).Return(responseBody, http.StatusOK, (*appErrors.AppError)(nil))
			},
			wantStatus: http.StatusOK,
		},
		{
			name:      "empty session id",
			sessionID: "",
			method:    http.MethodPost,
			body:      []byte(`{}`),
			setupMock: func(sessionSvc *svcMocks.SessionService, speechSvc *svcMocks.SpeechProxyService) {},
			wantErr:   true,
			errCode:   http.StatusBadRequest,
		},
		{
			name:      "get session by speech id error",
			sessionID: "speech-s1",
			method:    http.MethodPost,
			body:      []byte(`{}`),
			setupMock: func(sessionSvc *svcMocks.SessionService, speechSvc *svcMocks.SpeechProxyService) {
				sessionSvc.On("GetBySpeechSessionID", mock.Anything, "speech-s1", "user-1").Return((*models.Session)(nil), appErrors.Internal("not found"))
			},
			wantErr: true,
			errCode: http.StatusInternalServerError,
		},
		{
			name:      "speech proxy error",
			sessionID: "speech-s1",
			method:    http.MethodPost,
			body:      []byte(`{}`),
			setupMock: func(sessionSvc *svcMocks.SessionService, speechSvc *svcMocks.SpeechProxyService) {
				sessionSvc.On("GetBySpeechSessionID", mock.Anything, "speech-s1", "user-1").Return(pendingSession, (*appErrors.AppError)(nil))
				speechSvc.On("ProxyOffer", mock.Anything, "speech-s1", http.MethodPost, mock.Anything).Return([]byte{}, 0, appErrors.Internal("proxy error"))
			},
			wantErr: true,
			errCode: http.StatusInternalServerError,
		},
		{
			name:      "non-2xx status passes through without marking active",
			sessionID: "speech-s1",
			method:    http.MethodPost,
			body:      []byte(`{}`),
			setupMock: func(sessionSvc *svcMocks.SessionService, speechSvc *svcMocks.SpeechProxyService) {
				errorBody := []byte(`{"error":"gateway error"}`)
				sessionSvc.On("GetBySpeechSessionID", mock.Anything, "speech-s1", "user-1").Return(pendingSession, (*appErrors.AppError)(nil))
				speechSvc.On("ProxyOffer", mock.Anything, "speech-s1", http.MethodPost, mock.Anything).Return(errorBody, http.StatusBadGateway, (*appErrors.AppError)(nil))
			},
			wantStatus: http.StatusBadGateway,
			wantBody:   []byte(`{"error":"gateway error"}`),
		},
		{
			name:      "mark active error is logged but response still returned",
			sessionID: "speech-s1",
			method:    http.MethodPost,
			body:      []byte(`{}`),
			setupMock: func(sessionSvc *svcMocks.SessionService, speechSvc *svcMocks.SpeechProxyService) {
				sessionSvc.On("GetBySpeechSessionID", mock.Anything, "speech-s1", "user-1").Return(pendingSession, (*appErrors.AppError)(nil))
				speechSvc.On("ProxyOffer", mock.Anything, "speech-s1", http.MethodPost, mock.Anything).Return(responseBody, http.StatusOK, (*appErrors.AppError)(nil))
				sessionSvc.On("MarkSessionActive", mock.Anything, "s1").Return(appErrors.Internal("mark active error"))
			},
			wantStatus: http.StatusOK,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			configSvc := new(svcMocks.GlobalConfigService)
			sessionSvc := new(svcMocks.SessionService)
			speechSvc := new(svcMocks.SpeechProxyService)
			startConnSvc := new(svcMocks.StartConnectionService)

			tt.setupMock(sessionSvc, speechSvc)

			svc := services.NewModelGatewayService(configSvc, sessionSvc, speechSvc, startConnSvc, nil)
			ctx := setupSessionCtx("user-1")

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
	tests := []struct {
		name   string
		req    *req.CloseSessionReq
		setupMock func(
			sessionRepo *repoMocks.SessionRepository,
			quotaRepo *repoMocks.UserQuotaRepository,
			provider *svcMocks.Provider,
			uow *svcMocks.UnitOfWork,
		)
		wantErr bool
		errCode int
	}{
		{
			name: "success with quota release",
			req:  &req.CloseSessionReq{SessionID: "s1", ActualUsage: 60},
			setupMock: func(sessionRepo *repoMocks.SessionRepository, quotaRepo *repoMocks.UserQuotaRepository, provider *svcMocks.Provider, uow *svcMocks.UnitOfWork) {
				quotaDate := time.Now().Truncate(24 * time.Hour)
				session := &models.Session{BaseModel: models.BaseModel{ID: "s1"}, UserID: "user-1", Status: "active", ReservedAmount: 300, QuotaDate: &quotaDate}
				sessionRepo.On("GetForUpdate", mock.Anything, "s1").Return(session, nil)
				quotaRepo.On("Release", mock.Anything, "user-1", "voice", quotaDate, int64(240)).Return(nil)
				sessionRepo.On("UpdateStatus", mock.Anything, "s1", enums.SessionStatusInactive).Return(nil)
				provider.On("Session").Return(sessionRepo)
				provider.On("UserQuota").Return(quotaRepo)
				uow.SetProvider(provider)
				uow.On("Do", mock.Anything, mock.Anything).Return(nil)
			},
		},
		{
			name: "nil req body",
			req:  nil,
			setupMock: func(sessionRepo *repoMocks.SessionRepository, quotaRepo *repoMocks.UserQuotaRepository, provider *svcMocks.Provider, uow *svcMocks.UnitOfWork) {},
			wantErr: true,
			errCode: http.StatusBadRequest,
		},
		{
			name: "empty session id",
			req:  &req.CloseSessionReq{SessionID: "", ActualUsage: 60},
			setupMock: func(sessionRepo *repoMocks.SessionRepository, quotaRepo *repoMocks.UserQuotaRepository, provider *svcMocks.Provider, uow *svcMocks.UnitOfWork) {},
			wantErr: true,
			errCode: http.StatusBadRequest,
		},
		{
			name: "negative actualUsage",
			req:  &req.CloseSessionReq{SessionID: "s1", ActualUsage: -1},
			setupMock: func(sessionRepo *repoMocks.SessionRepository, quotaRepo *repoMocks.UserQuotaRepository, provider *svcMocks.Provider, uow *svcMocks.UnitOfWork) {},
			wantErr: true,
			errCode: http.StatusBadRequest,
		},
		{
			name: "session already inactive",
			req:  &req.CloseSessionReq{SessionID: "s1", ActualUsage: 60},
			setupMock: func(sessionRepo *repoMocks.SessionRepository, quotaRepo *repoMocks.UserQuotaRepository, provider *svcMocks.Provider, uow *svcMocks.UnitOfWork) {
				inactiveSession := &models.Session{BaseModel: models.BaseModel{ID: "s1"}, UserID: "user-1", Status: "inactive"}
				sessionRepo.On("GetForUpdate", mock.Anything, "s1").Return(inactiveSession, nil)
				provider.On("Session").Return(sessionRepo)
				provider.On("UserQuota").Return(quotaRepo)
				uow.SetProvider(provider)
				uow.On("Do", mock.Anything, mock.Anything).Return(nil)
			},
			wantErr: true,
			errCode: http.StatusBadRequest,
		},
		{
			name: "no quota release when quota_date is nil",
			req:  &req.CloseSessionReq{SessionID: "s1", ActualUsage: 60},
			setupMock: func(sessionRepo *repoMocks.SessionRepository, quotaRepo *repoMocks.UserQuotaRepository, provider *svcMocks.Provider, uow *svcMocks.UnitOfWork) {
				session := &models.Session{BaseModel: models.BaseModel{ID: "s1"}, UserID: "user-1", Status: "active", ReservedAmount: 300, QuotaDate: nil}
				sessionRepo.On("GetForUpdate", mock.Anything, "s1").Return(session, nil)
				sessionRepo.On("UpdateStatus", mock.Anything, "s1", enums.SessionStatusInactive).Return(nil)
				provider.On("Session").Return(sessionRepo)
				provider.On("UserQuota").Return(quotaRepo)
				uow.SetProvider(provider)
				uow.On("Do", mock.Anything, mock.Anything).Return(nil)
			},
		},
		{
			name: "actualUsage exceeds reservedAmount gets clamped",
			req:  &req.CloseSessionReq{SessionID: "s1", ActualUsage: 500},
			setupMock: func(sessionRepo *repoMocks.SessionRepository, quotaRepo *repoMocks.UserQuotaRepository, provider *svcMocks.Provider, uow *svcMocks.UnitOfWork) {
				quotaDate := time.Now().Truncate(24 * time.Hour)
				session := &models.Session{BaseModel: models.BaseModel{ID: "s1"}, UserID: "user-1", Status: "active", ReservedAmount: 300, QuotaDate: &quotaDate}
				sessionRepo.On("GetForUpdate", mock.Anything, "s1").Return(session, nil)
				quotaRepo.On("Release", mock.Anything, "user-1", "voice", quotaDate, int64(0)).Return(nil)
				sessionRepo.On("UpdateStatus", mock.Anything, "s1", enums.SessionStatusInactive).Return(nil)
				provider.On("Session").Return(sessionRepo)
				provider.On("UserQuota").Return(quotaRepo)
				uow.SetProvider(provider)
				uow.On("Do", mock.Anything, mock.Anything).Return(nil)
			},
		},
		{
			name: "quota release error",
			req:  &req.CloseSessionReq{SessionID: "s1", ActualUsage: 60},
			setupMock: func(sessionRepo *repoMocks.SessionRepository, quotaRepo *repoMocks.UserQuotaRepository, provider *svcMocks.Provider, uow *svcMocks.UnitOfWork) {
				quotaDate := time.Now().Truncate(24 * time.Hour)
				session := &models.Session{BaseModel: models.BaseModel{ID: "s1"}, UserID: "user-1", Status: "active", ReservedAmount: 300, QuotaDate: &quotaDate}
				sessionRepo.On("GetForUpdate", mock.Anything, "s1").Return(session, nil)
				quotaRepo.On("Release", mock.Anything, "user-1", "voice", quotaDate, int64(240)).Return(assert.AnError)
				provider.On("Session").Return(sessionRepo)
				provider.On("UserQuota").Return(quotaRepo)
				uow.SetProvider(provider)
				uow.On("Do", mock.Anything, mock.Anything).Return(nil)
			},
			wantErr: true,
			errCode: http.StatusInternalServerError,
		},
		{
			name: "update status error",
			req:  &req.CloseSessionReq{SessionID: "s1", ActualUsage: 60},
			setupMock: func(sessionRepo *repoMocks.SessionRepository, quotaRepo *repoMocks.UserQuotaRepository, provider *svcMocks.Provider, uow *svcMocks.UnitOfWork) {
				quotaDate := time.Now().Truncate(24 * time.Hour)
				session := &models.Session{BaseModel: models.BaseModel{ID: "s1"}, UserID: "user-1", Status: "active", ReservedAmount: 300, QuotaDate: &quotaDate}
				sessionRepo.On("GetForUpdate", mock.Anything, "s1").Return(session, nil)
				quotaRepo.On("Release", mock.Anything, "user-1", "voice", quotaDate, int64(240)).Return(nil)
				sessionRepo.On("UpdateStatus", mock.Anything, "s1", enums.SessionStatusInactive).Return(assert.AnError)
				provider.On("Session").Return(sessionRepo)
				provider.On("UserQuota").Return(quotaRepo)
				uow.SetProvider(provider)
				uow.On("Do", mock.Anything, mock.Anything).Return(nil)
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
			startConnSvc := new(svcMocks.StartConnectionService)
			sessionRepo := new(repoMocks.SessionRepository)
			quotaRepo := new(repoMocks.UserQuotaRepository)
			provider := new(svcMocks.Provider)
			uow := new(svcMocks.UnitOfWork)

			tt.setupMock(sessionRepo, quotaRepo, provider, uow)

			svc := services.NewModelGatewayService(configSvc, sessionSvc, speechSvc, startConnSvc, uow)
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

func TestModelGatewayService_CreateSession_JSONSerializable(t *testing.T) {
	defaultConfig := &models.GlobalConfig{
		Config: datatypes.JSON(`{"limits":{"session":{"max_session_lockTTL":60},"user":{"daily_voice_seconds":3600}}}`),
	}
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
	startConnSvc := new(svcMocks.StartConnectionService)

	configSvc.On("Get", mock.Anything).Return(defaultConfig, (*appErrors.AppError)(nil))
	sessionSvc.On("Create", mock.Anything, "user-1").Return(&models.Session{BaseModel: models.BaseModel{ID: "s1"}, UserID: "user-1", Status: enums.SessionStatusPending}, (*appErrors.AppError)(nil))
	startConnSvc.On("Start", mock.Anything, mock.Anything, "user-1", 3600).Return(&res.CreateSessionRes{
		ID:                  "s1",
		MaxDuration:         100,
		WebRTCConnectionRes: webrtcRes,
	}, (*appErrors.AppError)(nil))

	svc := services.NewModelGatewayService(configSvc, sessionSvc, speechSvc, startConnSvc, nil)
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
	startConnSvc := new(svcMocks.StartConnectionService)

	configSvc.On("Get", mock.Anything).Return(&models.GlobalConfig{
		Config: datatypes.JSON(`{"limits":{"session":{"max_session_lockTTL":60},"user":{"daily_voice_seconds":3600}}}`),
	}, (*appErrors.AppError)(nil))
	sessionSvc.On("Create", mock.Anything, "user-1").Return((*models.Session)(nil), appErrors.Conflict("close current session before starting a new one"))

	svc := services.NewModelGatewayService(configSvc, sessionSvc, speechSvc, startConnSvc, nil)
	ctx := setupSessionCtx("user-1")

	_, appErr := svc.CreateSession(ctx)
	require.NotNil(t, appErr)
	assert.Equal(t, http.StatusConflict, appErr.Code)
}
