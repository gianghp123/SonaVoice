package tests

import (
	"context"
	"encoding/json"
	"net/http"
	"testing"

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
			quotaSvc *svcMocks.SessionQuotaService,
			janitorSvc *svcMocks.SessionJanitorService,
			starterSvc *svcMocks.SessionStarterService,
			quotaRepo *repoMocks.UserQuotaRepository,
			sessionRepo *repoMocks.SessionRepository,
			provider *svcMocks.Provider,
			uow *svcMocks.UnitOfWork,
		)
		wantErr bool
		errCode int
	}{
		{
			name: "success",
			setupMock: func(configSvc *svcMocks.GlobalConfigService, sessionSvc *svcMocks.SessionService, speechSvc *svcMocks.SpeechProxyService, quotaSvc *svcMocks.SessionQuotaService, janitorSvc *svcMocks.SessionJanitorService, starterSvc *svcMocks.SessionStarterService, quotaRepo *repoMocks.UserQuotaRepository, sessionRepo *repoMocks.SessionRepository, provider *svcMocks.Provider, uow *svcMocks.UnitOfWork) {
				configSvc.On("Get", mock.Anything).Return(defaultConfig, (*appErrors.AppError)(nil))
				sessionRepo.On("AcquireLock", mock.Anything, "user-1").Return(nil)
				janitorSvc.On("CleanupStaleSessions", mock.Anything, mock.Anything, "user-1", int64(60), int64(3600)).Return(nil)
				sessionRepo.On("FindActiveByUserID", mock.Anything, "user-1").Return(nil, nil)
				sessionRepo.On("Create", mock.Anything, mock.Anything).Return(nil).Run(func(args mock.Arguments) {
					s := args.Get(1).(*models.Session)
					s.ID = "s1"
				})
				provider.On("Session").Return(sessionRepo)
				uow.SetProvider(provider)
				uow.On("Do", mock.Anything, mock.Anything).Return(nil)
				starterSvc.On("StartOrResume", mock.Anything, mock.Anything, "user-1", 3600).Return(&res.CreateSessionRes{
					ID:                  "s1",
					MaxDuration:         300,
					WebRTCConnectionRes: webrtcRes,
				}, (*appErrors.AppError)(nil))
			},
		},
		{
			name: "config service get error",
			setupMock: func(configSvc *svcMocks.GlobalConfigService, sessionSvc *svcMocks.SessionService, speechSvc *svcMocks.SpeechProxyService, quotaSvc *svcMocks.SessionQuotaService, janitorSvc *svcMocks.SessionJanitorService, starterSvc *svcMocks.SessionStarterService, quotaRepo *repoMocks.UserQuotaRepository, sessionRepo *repoMocks.SessionRepository, provider *svcMocks.Provider, uow *svcMocks.UnitOfWork) {
				configSvc.On("Get", mock.Anything).Return((*models.GlobalConfig)(nil), appErrors.Internal("config error"))
			},
			wantErr: true,
			errCode: http.StatusInternalServerError,
		},
		{
			name: "acquire lock error",
			setupMock: func(configSvc *svcMocks.GlobalConfigService, sessionSvc *svcMocks.SessionService, speechSvc *svcMocks.SpeechProxyService, quotaSvc *svcMocks.SessionQuotaService, janitorSvc *svcMocks.SessionJanitorService, starterSvc *svcMocks.SessionStarterService, quotaRepo *repoMocks.UserQuotaRepository, sessionRepo *repoMocks.SessionRepository, provider *svcMocks.Provider, uow *svcMocks.UnitOfWork) {
				configSvc.On("Get", mock.Anything).Return(defaultConfig, (*appErrors.AppError)(nil))
				sessionRepo.On("AcquireLock", mock.Anything, "user-1").Return(assert.AnError)
				provider.On("Session").Return(sessionRepo)
				uow.SetProvider(provider)
				uow.On("Do", mock.Anything, mock.Anything).Return(nil)
			},
			wantErr: true,
			errCode: http.StatusInternalServerError,
		},
		{
			name: "reserve quota error",
			setupMock: func(configSvc *svcMocks.GlobalConfigService, sessionSvc *svcMocks.SessionService, speechSvc *svcMocks.SpeechProxyService, quotaSvc *svcMocks.SessionQuotaService, janitorSvc *svcMocks.SessionJanitorService, starterSvc *svcMocks.SessionStarterService, quotaRepo *repoMocks.UserQuotaRepository, sessionRepo *repoMocks.SessionRepository, provider *svcMocks.Provider, uow *svcMocks.UnitOfWork) {
				configSvc.On("Get", mock.Anything).Return(defaultConfig, (*appErrors.AppError)(nil))
				sessionRepo.On("AcquireLock", mock.Anything, "user-1").Return(nil)
				janitorSvc.On("CleanupStaleSessions", mock.Anything, mock.Anything, "user-1", int64(60), int64(3600)).Return(nil)
				sessionRepo.On("FindActiveByUserID", mock.Anything, "user-1").Return(nil, nil)
				sessionRepo.On("Create", mock.Anything, mock.Anything).Return(nil).Run(func(args mock.Arguments) {
					s := args.Get(1).(*models.Session)
					s.ID = "s1"
				})
				provider.On("Session").Return(sessionRepo)
				uow.SetProvider(provider)
				uow.On("Do", mock.Anything, mock.Anything).Return(nil)
				starterSvc.On("StartOrResume", mock.Anything, mock.Anything, "user-1", 3600).Return((*res.CreateSessionRes)(nil), appErrors.Forbidden("quota exceeded"))
			},
			wantErr: true,
			errCode: http.StatusForbidden,
		},
		{
			name: "session create error",
			setupMock: func(configSvc *svcMocks.GlobalConfigService, sessionSvc *svcMocks.SessionService, speechSvc *svcMocks.SpeechProxyService, quotaSvc *svcMocks.SessionQuotaService, janitorSvc *svcMocks.SessionJanitorService, starterSvc *svcMocks.SessionStarterService, quotaRepo *repoMocks.UserQuotaRepository, sessionRepo *repoMocks.SessionRepository, provider *svcMocks.Provider, uow *svcMocks.UnitOfWork) {
				configSvc.On("Get", mock.Anything).Return(defaultConfig, (*appErrors.AppError)(nil))
				sessionRepo.On("AcquireLock", mock.Anything, "user-1").Return(nil)
				janitorSvc.On("CleanupStaleSessions", mock.Anything, mock.Anything, "user-1", int64(60), int64(3600)).Return(nil)
				sessionRepo.On("FindActiveByUserID", mock.Anything, "user-1").Return(nil, nil)
				sessionRepo.On("Create", mock.Anything, mock.Anything).Return(assert.AnError)
				provider.On("Session").Return(sessionRepo)
				uow.SetProvider(provider)
				uow.On("Do", mock.Anything, mock.Anything).Return(nil)
			},
			wantErr: true,
			errCode: http.StatusInternalServerError,
		},
		{
			name: "speech proxy start error marks session failed and releases quota",
			setupMock: func(configSvc *svcMocks.GlobalConfigService, sessionSvc *svcMocks.SessionService, speechSvc *svcMocks.SpeechProxyService, quotaSvc *svcMocks.SessionQuotaService, janitorSvc *svcMocks.SessionJanitorService, starterSvc *svcMocks.SessionStarterService, quotaRepo *repoMocks.UserQuotaRepository, sessionRepo *repoMocks.SessionRepository, provider *svcMocks.Provider, uow *svcMocks.UnitOfWork) {
				configSvc.On("Get", mock.Anything).Return(defaultConfig, (*appErrors.AppError)(nil))
				sessionRepo.On("AcquireLock", mock.Anything, "user-1").Return(nil)
				janitorSvc.On("CleanupStaleSessions", mock.Anything, mock.Anything, "user-1", int64(60), int64(3600)).Return(nil)
				sessionRepo.On("FindActiveByUserID", mock.Anything, "user-1").Return(nil, nil)
				sessionRepo.On("Create", mock.Anything, mock.Anything).Return(nil).Run(func(args mock.Arguments) {
					s := args.Get(1).(*models.Session)
					s.ID = "s1"
				})
				provider.On("Session").Return(sessionRepo)
				uow.SetProvider(provider)
				uow.On("Do", mock.Anything, mock.Anything).Return(nil)
				starterSvc.On("StartOrResume", mock.Anything, mock.Anything, "user-1", 3600).Return((*res.CreateSessionRes)(nil), appErrors.Internal("speech error"))
			},
			wantErr: true,
			errCode: http.StatusInternalServerError,
		},
		{
			name: "set speech session id error releases quota",
			setupMock: func(configSvc *svcMocks.GlobalConfigService, sessionSvc *svcMocks.SessionService, speechSvc *svcMocks.SpeechProxyService, quotaSvc *svcMocks.SessionQuotaService, janitorSvc *svcMocks.SessionJanitorService, starterSvc *svcMocks.SessionStarterService, quotaRepo *repoMocks.UserQuotaRepository, sessionRepo *repoMocks.SessionRepository, provider *svcMocks.Provider, uow *svcMocks.UnitOfWork) {
				configSvc.On("Get", mock.Anything).Return(defaultConfig, (*appErrors.AppError)(nil))
				sessionRepo.On("AcquireLock", mock.Anything, "user-1").Return(nil)
				janitorSvc.On("CleanupStaleSessions", mock.Anything, mock.Anything, "user-1", int64(60), int64(3600)).Return(nil)
				sessionRepo.On("FindActiveByUserID", mock.Anything, "user-1").Return(nil, nil)
				sessionRepo.On("Create", mock.Anything, mock.Anything).Return(nil).Run(func(args mock.Arguments) {
					s := args.Get(1).(*models.Session)
					s.ID = "s1"
				})
				provider.On("Session").Return(sessionRepo)
				uow.SetProvider(provider)
				uow.On("Do", mock.Anything, mock.Anything).Return(nil)
				starterSvc.On("StartOrResume", mock.Anything, mock.Anything, "user-1", 3600).Return((*res.CreateSessionRes)(nil), appErrors.Internal("update error"))
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
			quotaSvc := new(svcMocks.SessionQuotaService)
			janitorSvc := new(svcMocks.SessionJanitorService)
			starterSvc := new(svcMocks.SessionStarterService)
			quotaRepo := new(repoMocks.UserQuotaRepository)
			sessionRepo := new(repoMocks.SessionRepository)
			provider := new(svcMocks.Provider)
			uow := new(svcMocks.UnitOfWork)

			tt.setupMock(configSvc, sessionSvc, speechSvc, quotaSvc, janitorSvc, starterSvc, quotaRepo, sessionRepo, provider, uow)

			svc := services.NewModelGatewayService(configSvc, sessionSvc, speechSvc, quotaSvc, janitorSvc, starterSvc, uow)
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
	sessionRes := &models.Session{BaseModel: models.BaseModel{ID: "s1"}, UserID: "user-1", Status: "active", ReservedAmount: 300, DailyQuota: 3600, QuotaReleased: false}
	responseBody := []byte(`{"ok":true}`)
	nonSuccessBody := []byte(`{"error":"gateway error"}`)

	tests := []struct {
		name       string
		sessionID  string
		method     string
		body       []byte
		setupMock func(
			sessionSvc *svcMocks.SessionService,
			speechSvc *svcMocks.SpeechProxyService,
			quotaSvc *svcMocks.SessionQuotaService,
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
			setupMock: func(sessionSvc *svcMocks.SessionService, speechSvc *svcMocks.SpeechProxyService, quotaSvc *svcMocks.SessionQuotaService) {
				sessionSvc.On("GetBySpeechSessionID", mock.Anything, "speech-s1", "user-1").Return(sessionRes, (*appErrors.AppError)(nil))
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
			setupMock: func(sessionSvc *svcMocks.SessionService, speechSvc *svcMocks.SpeechProxyService, quotaSvc *svcMocks.SessionQuotaService) {
			},
			wantErr: true,
			errCode: http.StatusBadRequest,
		},
		{
			name:      "get session by speech id error",
			sessionID: "speech-s1",
			method:    http.MethodPost,
			body:      []byte(`{}`),
			setupMock: func(sessionSvc *svcMocks.SessionService, speechSvc *svcMocks.SpeechProxyService, quotaSvc *svcMocks.SessionQuotaService) {
				sessionSvc.On("GetBySpeechSessionID", mock.Anything, "speech-s1", "user-1").Return((*models.Session)(nil), appErrors.Internal("not found"))
			},
			wantErr: true,
			errCode: http.StatusInternalServerError,
		},
		{
			name:      "speech proxy error releases quota",
			sessionID: "speech-s1",
			method:    http.MethodPost,
			body:      []byte(`{}`),
			setupMock: func(sessionSvc *svcMocks.SessionService, speechSvc *svcMocks.SpeechProxyService, quotaSvc *svcMocks.SessionQuotaService) {
				sessionSvc.On("GetBySpeechSessionID", mock.Anything, "speech-s1", "user-1").Return(sessionRes, (*appErrors.AppError)(nil))
				speechSvc.On("ProxyOffer", mock.Anything, "speech-s1", http.MethodPost, mock.Anything).Return([]byte{}, 0, appErrors.Internal("proxy error"))
				sessionSvc.On("MarkSessionFailed", mock.Anything, "s1").Return((*appErrors.AppError)(nil))
				quotaSvc.On("ReleaseAll", mock.Anything, "user-1", int64(300)).Return(nil)
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
			setupMock: func(sessionSvc *svcMocks.SessionService, speechSvc *svcMocks.SpeechProxyService, quotaSvc *svcMocks.SessionQuotaService) {
				sessionSvc.On("GetBySpeechSessionID", mock.Anything, "speech-s1", "user-1").Return(sessionRes, (*appErrors.AppError)(nil))
				speechSvc.On("ProxyOffer", mock.Anything, "speech-s1", http.MethodPost, mock.Anything).Return(nonSuccessBody, http.StatusBadGateway, (*appErrors.AppError)(nil))
				sessionSvc.On("MarkSessionFailed", mock.Anything, "s1").Return((*appErrors.AppError)(nil))
				quotaSvc.On("ReleaseAll", mock.Anything, "user-1", int64(300)).Return(nil)
				sessionSvc.On("MarkQuotaReleased", mock.Anything, "s1").Return((*appErrors.AppError)(nil))
			},
			wantStatus: http.StatusBadGateway,
			wantBody:   nonSuccessBody,
		},
		{
			name:      "mark active error after success",
			sessionID: "speech-s1",
			method:    http.MethodPost,
			body:      []byte(`{}`),
			setupMock: func(sessionSvc *svcMocks.SessionService, speechSvc *svcMocks.SpeechProxyService, quotaSvc *svcMocks.SessionQuotaService) {
				sessionSvc.On("GetBySpeechSessionID", mock.Anything, "speech-s1", "user-1").Return(sessionRes, (*appErrors.AppError)(nil))
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
			quotaSvc := new(svcMocks.SessionQuotaService)
			janitorSvc := new(svcMocks.SessionJanitorService)
			starterSvc := new(svcMocks.SessionStarterService)
			uow := new(svcMocks.UnitOfWork)

			tt.setupMock(sessionSvc, speechSvc, quotaSvc)

			svc := services.NewModelGatewayService(configSvc, sessionSvc, speechSvc, quotaSvc, janitorSvc, starterSvc, uow)
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
	activeSession := &models.Session{BaseModel: models.BaseModel{ID: "s1"}, UserID: "user-1", Status: "active", ReservedAmount: 300, DailyQuota: 3600, QuotaReleased: false}

	tests := []struct {
		name   string
		req    *req.CloseSessionReq
		setupMock func(
			sessionSvc *svcMocks.SessionService,
			sessionRepo *repoMocks.SessionRepository,
			quotaRepo *repoMocks.UserQuotaRepository,
			provider *svcMocks.Provider,
			uow *svcMocks.UnitOfWork,
		)
		wantErr bool
		errCode int
	}{
		{
			name: "success",
			req:  &req.CloseSessionReq{SessionID: "s1", ActualUsage: 60},
			setupMock: func(sessionSvc *svcMocks.SessionService, sessionRepo *repoMocks.SessionRepository, quotaRepo *repoMocks.UserQuotaRepository, provider *svcMocks.Provider, uow *svcMocks.UnitOfWork) {
				sessionRepo.On("Get", mock.Anything, "s1").Return(activeSession, nil)
				quotaRepo.On("Release", mock.Anything, "user-1", "voice", mock.Anything, int64(240)).Return(nil)
				sessionRepo.On("UpdateQuotaReleased", mock.Anything, "s1").Return(nil)
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
			setupMock: func(sessionSvc *svcMocks.SessionService, sessionRepo *repoMocks.SessionRepository, quotaRepo *repoMocks.UserQuotaRepository, provider *svcMocks.Provider, uow *svcMocks.UnitOfWork) {
			},
			wantErr: true,
			errCode: http.StatusBadRequest,
		},
		{
			name: "empty session id",
			req:  &req.CloseSessionReq{SessionID: "", ActualUsage: 60},
			setupMock: func(sessionSvc *svcMocks.SessionService, sessionRepo *repoMocks.SessionRepository, quotaRepo *repoMocks.UserQuotaRepository, provider *svcMocks.Provider, uow *svcMocks.UnitOfWork) {
			},
			wantErr: true,
			errCode: http.StatusBadRequest,
		},
		{
			name: "negative actualUsage",
			req:  &req.CloseSessionReq{SessionID: "s1", ActualUsage: -1},
			setupMock: func(sessionSvc *svcMocks.SessionService, sessionRepo *repoMocks.SessionRepository, quotaRepo *repoMocks.UserQuotaRepository, provider *svcMocks.Provider, uow *svcMocks.UnitOfWork) {
			},
			wantErr: true,
			errCode: http.StatusBadRequest,
		},
		{
			name: "session get error",
			req:  &req.CloseSessionReq{SessionID: "s1", ActualUsage: 60},
			setupMock: func(sessionSvc *svcMocks.SessionService, sessionRepo *repoMocks.SessionRepository, quotaRepo *repoMocks.UserQuotaRepository, provider *svcMocks.Provider, uow *svcMocks.UnitOfWork) {
				sessionRepo.On("Get", mock.Anything, "s1").Return((*models.Session)(nil), assert.AnError)
				provider.On("Session").Return(sessionRepo)
				provider.On("UserQuota").Return(quotaRepo)
				uow.SetProvider(provider)
				uow.On("Do", mock.Anything, mock.Anything).Return(nil)
			},
			wantErr: true,
			errCode: http.StatusInternalServerError,
		},
		{
			name: "session already inactive",
			req:  &req.CloseSessionReq{SessionID: "s1", ActualUsage: 60},
			setupMock: func(sessionSvc *svcMocks.SessionService, sessionRepo *repoMocks.SessionRepository, quotaRepo *repoMocks.UserQuotaRepository, provider *svcMocks.Provider, uow *svcMocks.UnitOfWork) {
				inactiveSession := &models.Session{BaseModel: models.BaseModel{ID: "s1"}, UserID: "user-1", Status: "inactive", ReservedAmount: 300, DailyQuota: 3600, QuotaReleased: true}
				sessionRepo.On("Get", mock.Anything, "s1").Return(inactiveSession, nil)
				provider.On("Session").Return(sessionRepo)
				provider.On("UserQuota").Return(quotaRepo)
				uow.SetProvider(provider)
				uow.On("Do", mock.Anything, mock.Anything).Return(nil)
			},
			wantErr: true,
			errCode: http.StatusBadRequest,
		},
		{
			name: "quota already released skips release",
			req:  &req.CloseSessionReq{SessionID: "s1", ActualUsage: 60},
			setupMock: func(sessionSvc *svcMocks.SessionService, sessionRepo *repoMocks.SessionRepository, quotaRepo *repoMocks.UserQuotaRepository, provider *svcMocks.Provider, uow *svcMocks.UnitOfWork) {
				releasedSession := &models.Session{BaseModel: models.BaseModel{ID: "s1"}, UserID: "user-1", Status: "active", ReservedAmount: 300, DailyQuota: 3600, QuotaReleased: true}
				sessionRepo.On("Get", mock.Anything, "s1").Return(releasedSession, nil)
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
			setupMock: func(sessionSvc *svcMocks.SessionService, sessionRepo *repoMocks.SessionRepository, quotaRepo *repoMocks.UserQuotaRepository, provider *svcMocks.Provider, uow *svcMocks.UnitOfWork) {
				sessionRepo.On("Get", mock.Anything, "s1").Return(activeSession, nil)
				quotaRepo.On("Release", mock.Anything, "user-1", "voice", mock.Anything, int64(0)).Return(nil)
				sessionRepo.On("UpdateQuotaReleased", mock.Anything, "s1").Return(nil)
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
			setupMock: func(sessionSvc *svcMocks.SessionService, sessionRepo *repoMocks.SessionRepository, quotaRepo *repoMocks.UserQuotaRepository, provider *svcMocks.Provider, uow *svcMocks.UnitOfWork) {
				sessionRepo.On("Get", mock.Anything, "s1").Return(activeSession, nil)
				quotaRepo.On("Release", mock.Anything, "user-1", "voice", mock.Anything, int64(240)).Return(assert.AnError)
				provider.On("Session").Return(sessionRepo)
				provider.On("UserQuota").Return(quotaRepo)
				uow.SetProvider(provider)
				uow.On("Do", mock.Anything, mock.Anything).Return(nil)
			},
			wantErr: true,
			errCode: http.StatusInternalServerError,
		},
		{
			name: "mark session inactive error",
			req:  &req.CloseSessionReq{SessionID: "s1", ActualUsage: 60},
			setupMock: func(sessionSvc *svcMocks.SessionService, sessionRepo *repoMocks.SessionRepository, quotaRepo *repoMocks.UserQuotaRepository, provider *svcMocks.Provider, uow *svcMocks.UnitOfWork) {
				sessionRepo.On("Get", mock.Anything, "s1").Return(activeSession, nil)
				quotaRepo.On("Release", mock.Anything, "user-1", "voice", mock.Anything, int64(240)).Return(nil)
				sessionRepo.On("UpdateQuotaReleased", mock.Anything, "s1").Return(nil)
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
			quotaSvc := new(svcMocks.SessionQuotaService)
			janitorSvc := new(svcMocks.SessionJanitorService)
			starterSvc := new(svcMocks.SessionStarterService)
			sessionRepo := new(repoMocks.SessionRepository)
			quotaRepo := new(repoMocks.UserQuotaRepository)
			provider := new(svcMocks.Provider)
			uow := new(svcMocks.UnitOfWork)

			tt.setupMock(sessionSvc, sessionRepo, quotaRepo, provider, uow)

			svc := services.NewModelGatewayService(configSvc, sessionSvc, speechSvc, quotaSvc, janitorSvc, starterSvc, uow)
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
	activeSession := &models.Session{BaseModel: models.BaseModel{ID: "s1"}, UserID: "user-1", Status: "active", ReservedAmount: 200, DailyQuota: 3600, QuotaReleased: false}

	sessionRepo := new(repoMocks.SessionRepository)
	quotaRepo := new(repoMocks.UserQuotaRepository)
	configSvc := new(svcMocks.GlobalConfigService)
	sessionSvc := new(svcMocks.SessionService)
	speechSvc := new(svcMocks.SpeechProxyService)
	quotaSvc := new(svcMocks.SessionQuotaService)
	janitorSvc := new(svcMocks.SessionJanitorService)
	starterSvc := new(svcMocks.SessionStarterService)
	provider := new(svcMocks.Provider)
	uow := new(svcMocks.UnitOfWork)

	sessionRepo.On("Get", mock.Anything, "s1").Return(activeSession, nil)
	quotaRepo.On("Release", mock.Anything, "user-1", "voice", mock.Anything, int64(0)).Return(nil)
	sessionRepo.On("UpdateQuotaReleased", mock.Anything, "s1").Return(nil)
	sessionRepo.On("UpdateStatus", mock.Anything, "s1", enums.SessionStatusInactive).Return(nil)
	provider.On("Session").Return(sessionRepo)
	provider.On("UserQuota").Return(quotaRepo)
	uow.SetProvider(provider)
	uow.On("Do", mock.Anything, mock.Anything).Return(nil)

	svc := services.NewModelGatewayService(configSvc, sessionSvc, speechSvc, quotaSvc, janitorSvc, starterSvc, uow)
	ctx := context.Background()

	appErr := svc.CloseSession(ctx, &req.CloseSessionReq{SessionID: "s1", ActualUsage: 500})
	require.Nil(t, appErr)
}

func TestModelGatewayService_CreateSession_PanicOnCleanup(t *testing.T) {
	defaultConfig := &models.GlobalConfig{
		Config: datatypes.JSON(`{"limits":{"session":{"max_session_lockTTL":60},"user":{"daily_voice_seconds":3600}}}`),
	}

	configSvc := new(svcMocks.GlobalConfigService)
	sessionSvc := new(svcMocks.SessionService)
	speechSvc := new(svcMocks.SpeechProxyService)
	quotaSvc := new(svcMocks.SessionQuotaService)
	janitorSvc := new(svcMocks.SessionJanitorService)
	starterSvc := new(svcMocks.SessionStarterService)
	sessionRepo := new(repoMocks.SessionRepository)
	provider := new(svcMocks.Provider)
	uow := new(svcMocks.UnitOfWork)

	configSvc.On("Get", mock.Anything).Return(defaultConfig, (*appErrors.AppError)(nil))
	sessionRepo.On("AcquireLock", mock.Anything, "user-1").Return(nil)
	janitorSvc.On("CleanupStaleSessions", mock.Anything, mock.Anything, "user-1", int64(60), int64(3600)).Return(nil)
	sessionRepo.On("FindActiveByUserID", mock.Anything, "user-1").Return(nil, nil)
	sessionRepo.On("Create", mock.Anything, mock.Anything).Return(nil).Run(func(args mock.Arguments) {
		s := args.Get(1).(*models.Session)
		s.ID = "s1"
	})
	provider.On("Session").Return(sessionRepo)
	uow.SetProvider(provider)
	uow.On("Do", mock.Anything, mock.Anything).Return(nil)
	starterSvc.On("StartOrResume", mock.Anything, mock.Anything, "user-1", 3600).Return((*res.CreateSessionRes)(nil), appErrors.Internal("speech error"))

	svc := services.NewModelGatewayService(configSvc, sessionSvc, speechSvc, quotaSvc, janitorSvc, starterSvc, uow)
	ctx := setupSessionCtx("user-1")

	result, appErr := svc.CreateSession(ctx)
	require.NotNil(t, appErr)
	require.Nil(t, result)
}

func TestModelGatewayService_CreateSession_CleanupOnSetReservationError(t *testing.T) {
	defaultConfig := &models.GlobalConfig{
		Config: datatypes.JSON(`{"limits":{"session":{"max_session_lockTTL":60},"user":{"daily_voice_seconds":3600}}}`),
	}

	configSvc := new(svcMocks.GlobalConfigService)
	sessionSvc := new(svcMocks.SessionService)
	speechSvc := new(svcMocks.SpeechProxyService)
	quotaSvc := new(svcMocks.SessionQuotaService)
	janitorSvc := new(svcMocks.SessionJanitorService)
	starterSvc := new(svcMocks.SessionStarterService)
	sessionRepo := new(repoMocks.SessionRepository)
	provider := new(svcMocks.Provider)
	uow := new(svcMocks.UnitOfWork)

	configSvc.On("Get", mock.Anything).Return(defaultConfig, (*appErrors.AppError)(nil))
	sessionRepo.On("AcquireLock", mock.Anything, "user-1").Return(nil)
	janitorSvc.On("CleanupStaleSessions", mock.Anything, mock.Anything, "user-1", int64(60), int64(3600)).Return(nil)
	sessionRepo.On("FindActiveByUserID", mock.Anything, "user-1").Return(nil, nil)
	sessionRepo.On("Create", mock.Anything, mock.Anything).Return(nil).Run(func(args mock.Arguments) {
		s := args.Get(1).(*models.Session)
		s.ID = "s1"
	})
	provider.On("Session").Return(sessionRepo)
	uow.SetProvider(provider)
	uow.On("Do", mock.Anything, mock.Anything).Return(nil)
	starterSvc.On("StartOrResume", mock.Anything, mock.Anything, "user-1", 3600).Return((*res.CreateSessionRes)(nil), appErrors.Internal("set reservation error"))

	svc := services.NewModelGatewayService(configSvc, sessionSvc, speechSvc, quotaSvc, janitorSvc, starterSvc, uow)
	ctx := setupSessionCtx("user-1")

	result, appErr := svc.CreateSession(ctx)
	require.NotNil(t, appErr)
	require.Nil(t, result)
}

func TestModelGatewayService_ProxyOffer_NonJSONBody(t *testing.T) {
	sessionRes := &models.Session{BaseModel: models.BaseModel{ID: "s1"}, UserID: "user-1", Status: "active", ReservedAmount: 300, DailyQuota: 3600, QuotaReleased: false}
	rawBody := []byte(`raw body`)

	sessionSvc := new(svcMocks.SessionService)
	speechSvc := new(svcMocks.SpeechProxyService)
	configSvc := new(svcMocks.GlobalConfigService)
	quotaSvc := new(svcMocks.SessionQuotaService)
	janitorSvc := new(svcMocks.SessionJanitorService)
	starterSvc := new(svcMocks.SessionStarterService)
	uow := new(svcMocks.UnitOfWork)

	sessionSvc.On("GetBySpeechSessionID", mock.Anything, "speech-s1", "user-1").Return(sessionRes, (*appErrors.AppError)(nil))
	speechSvc.On("ProxyOffer", mock.Anything, "speech-s1", http.MethodPost, rawBody).Return([]byte("resp"), http.StatusOK, (*appErrors.AppError)(nil))
	sessionSvc.On("MarkSessionActive", mock.Anything, "s1").Return((*appErrors.AppError)(nil))

	svc := services.NewModelGatewayService(configSvc, sessionSvc, speechSvc, quotaSvc, janitorSvc, starterSvc, uow)
	ctx := setupSessionCtx("user-1")

	respBody, statusCode, appErr := svc.ProxyOffer(ctx, "speech-s1", http.MethodPost, rawBody)
	require.Nil(t, appErr)
	assert.Equal(t, http.StatusOK, statusCode)
	assert.Equal(t, []byte("resp"), respBody)
}

func TestModelGatewayService_ProxyOffer_NonSuccessWithFailedMark(t *testing.T) {
	sessionRes := &models.Session{BaseModel: models.BaseModel{ID: "s1"}, UserID: "user-1", Status: "active", ReservedAmount: 300, DailyQuota: 3600, QuotaReleased: false}
	sessionSvc := new(svcMocks.SessionService)
	speechSvc := new(svcMocks.SpeechProxyService)
	configSvc := new(svcMocks.GlobalConfigService)
	quotaSvc := new(svcMocks.SessionQuotaService)
	janitorSvc := new(svcMocks.SessionJanitorService)
	starterSvc := new(svcMocks.SessionStarterService)
	uow := new(svcMocks.UnitOfWork)

	sessionSvc.On("GetBySpeechSessionID", mock.Anything, "speech-s1", "user-1").Return(sessionRes, (*appErrors.AppError)(nil))
	speechSvc.On("ProxyOffer", mock.Anything, "speech-s1", http.MethodPost, mock.Anything).Return([]byte("err"), http.StatusBadRequest, (*appErrors.AppError)(nil))
	sessionSvc.On("MarkSessionFailed", mock.Anything, "s1").Return((*appErrors.AppError)(nil))
	quotaSvc.On("ReleaseAll", mock.Anything, "user-1", int64(300)).Return(nil)
	sessionSvc.On("MarkQuotaReleased", mock.Anything, "s1").Return((*appErrors.AppError)(nil))

	svc := services.NewModelGatewayService(configSvc, sessionSvc, speechSvc, quotaSvc, janitorSvc, starterSvc, uow)
	ctx := setupSessionCtx("user-1")

	respBody, statusCode, appErr := svc.ProxyOffer(ctx, "speech-s1", http.MethodPost, []byte(`{}`))
	require.Nil(t, appErr)
	assert.Equal(t, http.StatusBadRequest, statusCode)
	assert.Equal(t, []byte("err"), respBody)
}

func TestModelGatewayService_ProxyOffer_2xxStatusCodes(t *testing.T) {
	sessionRes := &models.Session{BaseModel: models.BaseModel{ID: "s1"}, UserID: "user-1", Status: "active", ReservedAmount: 300, DailyQuota: 3600, QuotaReleased: false}

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
			quotaSvc := new(svcMocks.SessionQuotaService)
			janitorSvc := new(svcMocks.SessionJanitorService)
			starterSvc := new(svcMocks.SessionStarterService)
			uow := new(svcMocks.UnitOfWork)

			sessionSvc.On("GetBySpeechSessionID", mock.Anything, "speech-s1", "user-1").Return(sessionRes, (*appErrors.AppError)(nil))
			speechSvc.On("ProxyOffer", mock.Anything, "speech-s1", http.MethodPost, mock.Anything).Return([]byte("body"), tt.statusCode, (*appErrors.AppError)(nil))
			sessionSvc.On("MarkSessionActive", mock.Anything, "s1").Return((*appErrors.AppError)(nil))

			svc := services.NewModelGatewayService(configSvc, sessionSvc, speechSvc, quotaSvc, janitorSvc, starterSvc, uow)
			ctx := setupSessionCtx("user-1")

			respBody, statusCode, appErr := svc.ProxyOffer(ctx, "speech-s1", http.MethodPost, []byte(`{}`))
			require.Nil(t, appErr)
			assert.Equal(t, tt.statusCode, statusCode)
			assert.Equal(t, []byte("body"), respBody)
		})
	}
}

func TestModelGatewayService_ProxyOffer_Non200StatusCodes(t *testing.T) {
	sessionRes := &models.Session{BaseModel: models.BaseModel{ID: "s1"}, UserID: "user-1", Status: "active", ReservedAmount: 300, DailyQuota: 3600, QuotaReleased: false}

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
			quotaSvc := new(svcMocks.SessionQuotaService)
			janitorSvc := new(svcMocks.SessionJanitorService)
			starterSvc := new(svcMocks.SessionStarterService)
			uow := new(svcMocks.UnitOfWork)

			sessionSvc.On("GetBySpeechSessionID", mock.Anything, "speech-s1", "user-1").Return(sessionRes, (*appErrors.AppError)(nil))
			speechSvc.On("ProxyOffer", mock.Anything, "speech-s1", http.MethodPost, mock.Anything).Return([]byte("body"), tt.statusCode, (*appErrors.AppError)(nil))
			sessionSvc.On("MarkSessionFailed", mock.Anything, "s1").Return((*appErrors.AppError)(nil))
			quotaSvc.On("ReleaseAll", mock.Anything, "user-1", int64(300)).Return(nil)
			sessionSvc.On("MarkQuotaReleased", mock.Anything, "s1").Return((*appErrors.AppError)(nil))

			svc := services.NewModelGatewayService(configSvc, sessionSvc, speechSvc, quotaSvc, janitorSvc, starterSvc, uow)
			ctx := setupSessionCtx("user-1")

			respBody, statusCode, appErr := svc.ProxyOffer(ctx, "speech-s1", http.MethodPost, []byte(`{}`))
			require.Nil(t, appErr)
			assert.Equal(t, tt.statusCode, statusCode)
			assert.Equal(t, []byte("body"), respBody)
		})
	}
}

func TestModelGatewayService_CloseSession_PendingSessionSucceeds(t *testing.T) {
	pendingSession := &models.Session{BaseModel: models.BaseModel{ID: "s1"}, UserID: "user-1", Status: "pending", ReservedAmount: 300, DailyQuota: 3600, QuotaReleased: false}

	sessionRepo := new(repoMocks.SessionRepository)
	quotaRepo := new(repoMocks.UserQuotaRepository)
	configSvc := new(svcMocks.GlobalConfigService)
	sessionSvc := new(svcMocks.SessionService)
	speechSvc := new(svcMocks.SpeechProxyService)
	quotaSvc := new(svcMocks.SessionQuotaService)
	janitorSvc := new(svcMocks.SessionJanitorService)
	starterSvc := new(svcMocks.SessionStarterService)
	provider := new(svcMocks.Provider)
	uow := new(svcMocks.UnitOfWork)

	sessionRepo.On("Get", mock.Anything, "s1").Return(pendingSession, nil)
	quotaRepo.On("Release", mock.Anything, "user-1", "voice", mock.Anything, int64(300)).Return(nil)
	sessionRepo.On("UpdateQuotaReleased", mock.Anything, "s1").Return(nil)
	sessionRepo.On("UpdateStatus", mock.Anything, "s1", enums.SessionStatusInactive).Return(nil)
	provider.On("Session").Return(sessionRepo)
	provider.On("UserQuota").Return(quotaRepo)
	uow.SetProvider(provider)
	uow.On("Do", mock.Anything, mock.Anything).Return(nil)

	svc := services.NewModelGatewayService(configSvc, sessionSvc, speechSvc, quotaSvc, janitorSvc, starterSvc, uow)
	ctx := context.Background()

	appErr := svc.CloseSession(ctx, &req.CloseSessionReq{SessionID: "s1", ActualUsage: 0})
	require.Nil(t, appErr)
}

func TestModelGatewayService_CloseSession_AlreadyReleased(t *testing.T) {
	releasedSession := &models.Session{BaseModel: models.BaseModel{ID: "s1"}, UserID: "user-1", Status: "active", ReservedAmount: 300, DailyQuota: 3600, QuotaReleased: true}

	sessionRepo := new(repoMocks.SessionRepository)
	quotaRepo := new(repoMocks.UserQuotaRepository)
	configSvc := new(svcMocks.GlobalConfigService)
	sessionSvc := new(svcMocks.SessionService)
	speechSvc := new(svcMocks.SpeechProxyService)
	quotaSvc := new(svcMocks.SessionQuotaService)
	janitorSvc := new(svcMocks.SessionJanitorService)
	starterSvc := new(svcMocks.SessionStarterService)
	provider := new(svcMocks.Provider)
	uow := new(svcMocks.UnitOfWork)

	sessionRepo.On("Get", mock.Anything, "s1").Return(releasedSession, nil)
	sessionRepo.On("UpdateStatus", mock.Anything, "s1", enums.SessionStatusInactive).Return(nil)
	provider.On("Session").Return(sessionRepo)
	provider.On("UserQuota").Return(quotaRepo)
	uow.SetProvider(provider)
	uow.On("Do", mock.Anything, mock.Anything).Return(nil)

	svc := services.NewModelGatewayService(configSvc, sessionSvc, speechSvc, quotaSvc, janitorSvc, starterSvc, uow)
	ctx := context.Background()

	appErr := svc.CloseSession(ctx, &req.CloseSessionReq{SessionID: "s1", ActualUsage: 60})
	require.Nil(t, appErr)
}

func TestModelGatewayService_ProxyOffer_ContextPropagation(t *testing.T) {
	sessionRes := &models.Session{BaseModel: models.BaseModel{ID: "s1"}, UserID: "user-1", Status: "active", ReservedAmount: 300, DailyQuota: 3600, QuotaReleased: false}
	type ctxKey string
	testKey := ctxKey("test-key")
	ctx := context.WithValue(setupSessionCtx("user-1"), testKey, "test-value")

	sessionSvc := new(svcMocks.SessionService)
	speechSvc := new(svcMocks.SpeechProxyService)
	configSvc := new(svcMocks.GlobalConfigService)
	quotaSvc := new(svcMocks.SessionQuotaService)
	janitorSvc := new(svcMocks.SessionJanitorService)
	starterSvc := new(svcMocks.SessionStarterService)
	uow := new(svcMocks.UnitOfWork)

	sessionSvc.On("GetBySpeechSessionID", mock.Anything, "speech-s1", "user-1").Return(sessionRes, (*appErrors.AppError)(nil))
	speechSvc.On("ProxyOffer", mock.Anything, "speech-s1", http.MethodPost, mock.Anything).Return([]byte("ok"), http.StatusOK, (*appErrors.AppError)(nil))
	sessionSvc.On("MarkSessionActive", mock.Anything, "s1").Return((*appErrors.AppError)(nil))

	svc := services.NewModelGatewayService(configSvc, sessionSvc, speechSvc, quotaSvc, janitorSvc, starterSvc, uow)

	respBody, statusCode, appErr := svc.ProxyOffer(ctx, "speech-s1", http.MethodPost, []byte(`{}`))
	require.Nil(t, appErr)
	assert.Equal(t, http.StatusOK, statusCode)
	assert.Equal(t, []byte("ok"), respBody)
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
	quotaSvc := new(svcMocks.SessionQuotaService)
	janitorSvc := new(svcMocks.SessionJanitorService)
	starterSvc := new(svcMocks.SessionStarterService)
	sessionRepo := new(repoMocks.SessionRepository)
	provider := new(svcMocks.Provider)
	uow := new(svcMocks.UnitOfWork)

	configSvc.On("Get", mock.Anything).Return(defaultConfig, (*appErrors.AppError)(nil))
	sessionRepo.On("AcquireLock", mock.Anything, "user-1").Return(nil)
	janitorSvc.On("CleanupStaleSessions", mock.Anything, mock.Anything, "user-1", int64(60), int64(3600)).Return(nil)
	sessionRepo.On("FindActiveByUserID", mock.Anything, "user-1").Return(nil, nil)
	sessionRepo.On("Create", mock.Anything, mock.Anything).Return(nil).Run(func(args mock.Arguments) {
		s := args.Get(1).(*models.Session)
		s.ID = "s1"
	})
	provider.On("Session").Return(sessionRepo)
	uow.SetProvider(provider)
	uow.On("Do", mock.Anything, mock.Anything).Return(nil)
	starterSvc.On("StartOrResume", mock.Anything, mock.Anything, "user-1", 3600).Return(&res.CreateSessionRes{
		ID:                  "s1",
		MaxDuration:         100,
		WebRTCConnectionRes: webrtcRes,
	}, (*appErrors.AppError)(nil))

	svc := services.NewModelGatewayService(configSvc, sessionSvc, speechSvc, quotaSvc, janitorSvc, starterSvc, uow)
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
	quotaSvc := new(svcMocks.SessionQuotaService)
	janitorSvc := new(svcMocks.SessionJanitorService)
	starterSvc := new(svcMocks.SessionStarterService)
	sessionRepo := new(repoMocks.SessionRepository)
	provider := new(svcMocks.Provider)
	uow := new(svcMocks.UnitOfWork)

	configWithLockTTL := &models.GlobalConfig{
		Config: datatypes.JSON(`{"limits":{"session":{"max_session_lockTTL":60}}}`),
	}
	configSvc.On("Get", mock.Anything).Return(configWithLockTTL, (*appErrors.AppError)(nil))
	sessionRepo.On("AcquireLock", mock.Anything, "user-1").Return(nil)
	janitorSvc.On("CleanupStaleSessions", mock.Anything, mock.Anything, "user-1", int64(60), int64(0)).Return(nil)
	sessionRepo.On("FindActiveByUserID", mock.Anything, "user-1").Return(&models.Session{BaseModel: models.BaseModel{ID: "existing-active"}, UserID: "user-1", Status: "active"}, nil)
	provider.On("Session").Return(sessionRepo)
	uow.SetProvider(provider)
	uow.On("Do", mock.Anything, mock.Anything).Return(nil)

	svc := services.NewModelGatewayService(configSvc, sessionSvc, speechSvc, quotaSvc, janitorSvc, starterSvc, uow)
	ctx := setupSessionCtx("user-1")

	_, appErr := svc.CreateSession(ctx)
	require.NotNil(t, appErr)
	assert.Equal(t, http.StatusConflict, appErr.Code)
}

func TestModelGatewayService_ResumeSession_NotOwner(t *testing.T) {
	configSvc := new(svcMocks.GlobalConfigService)
	sessionSvc := new(svcMocks.SessionService)
	speechSvc := new(svcMocks.SpeechProxyService)
	quotaSvc := new(svcMocks.SessionQuotaService)
	janitorSvc := new(svcMocks.SessionJanitorService)
	starterSvc := new(svcMocks.SessionStarterService)
	uow := new(svcMocks.UnitOfWork)

	sessionSvc.On("GetInternal", mock.Anything, "session-1").Return(&models.Session{BaseModel: models.BaseModel{ID: "session-1"}, UserID: "user-2", Status: "inactive"}, (*appErrors.AppError)(nil))

	svc := services.NewModelGatewayService(configSvc, sessionSvc, speechSvc, quotaSvc, janitorSvc, starterSvc, uow)
	ctx := setupSessionCtx("user-1")

	_, appErr := svc.ResumeSession(ctx, "session-1")
	require.NotNil(t, appErr)
	assert.Equal(t, http.StatusForbidden, appErr.Code)
}

func TestModelGatewayService_ResumeSession_NotResumable(t *testing.T) {
	configSvc := new(svcMocks.GlobalConfigService)
	sessionSvc := new(svcMocks.SessionService)
	speechSvc := new(svcMocks.SpeechProxyService)
	quotaSvc := new(svcMocks.SessionQuotaService)
	janitorSvc := new(svcMocks.SessionJanitorService)
	starterSvc := new(svcMocks.SessionStarterService)
	uow := new(svcMocks.UnitOfWork)

	sessionSvc.On("GetInternal", mock.Anything, "session-1").Return(&models.Session{BaseModel: models.BaseModel{ID: "session-1"}, UserID: "user-1", Status: "active"}, (*appErrors.AppError)(nil))

	svc := services.NewModelGatewayService(configSvc, sessionSvc, speechSvc, quotaSvc, janitorSvc, starterSvc, uow)
	ctx := setupSessionCtx("user-1")

	_, appErr := svc.ResumeSession(ctx, "session-1")
	require.NotNil(t, appErr)
	assert.Equal(t, http.StatusBadRequest, appErr.Code)
}

func TestModelGatewayService_ResumeSession_ConflictWhenActiveSession(t *testing.T) {
	configSvc := new(svcMocks.GlobalConfigService)
	sessionSvc := new(svcMocks.SessionService)
	speechSvc := new(svcMocks.SpeechProxyService)
	quotaSvc := new(svcMocks.SessionQuotaService)
	janitorSvc := new(svcMocks.SessionJanitorService)
	starterSvc := new(svcMocks.SessionStarterService)
	quotaRepo := new(repoMocks.UserQuotaRepository)
	sessionRepo := new(repoMocks.SessionRepository)
	provider := new(svcMocks.Provider)
	uow := new(svcMocks.UnitOfWork)

	configWithLockTTL := &models.GlobalConfig{
		Config: datatypes.JSON(`{"limits":{"session":{"max_session_lockTTL":60}}}`),
	}
	sessionSvc.On("GetInternal", mock.Anything, "session-1").Return(&models.Session{BaseModel: models.BaseModel{ID: "session-1"}, UserID: "user-1", Status: "inactive"}, (*appErrors.AppError)(nil))
	configSvc.On("Get", mock.Anything).Return(configWithLockTTL, (*appErrors.AppError)(nil))
	sessionRepo.On("AcquireLock", mock.Anything, "user-1").Return(nil)
	janitorSvc.On("CleanupStaleSessions", mock.Anything, mock.Anything, "user-1", int64(60), int64(0)).Return(nil)
	sessionRepo.On("FindActiveByUserID", mock.Anything, "user-1").Return(&models.Session{BaseModel: models.BaseModel{ID: "session-2"}, UserID: "user-1", Status: "active"}, nil)
	provider.On("Session").Return(sessionRepo)
	provider.On("UserQuota").Return(quotaRepo)
	uow.SetProvider(provider)
	uow.On("Do", mock.Anything, mock.Anything).Return(nil)

	svc := services.NewModelGatewayService(configSvc, sessionSvc, speechSvc, quotaSvc, janitorSvc, starterSvc, uow)
	ctx := setupSessionCtx("user-1")

	_, appErr := svc.ResumeSession(ctx, "session-1")
	require.NotNil(t, appErr)
	assert.Equal(t, http.StatusConflict, appErr.Code)
}
