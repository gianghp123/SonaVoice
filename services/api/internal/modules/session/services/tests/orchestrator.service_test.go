package tests

import (
	"context"
	"net/http"
	"testing"
	"time"

	appErrors "github.com/gianghp123/SonaVoice/api/internal/core/errors"
	"github.com/gianghp123/SonaVoice/api/internal/core/enums"
	"github.com/gianghp123/SonaVoice/api/internal/core/response"
	"github.com/gianghp123/SonaVoice/api/internal/database/models"
	repoMocks "github.com/gianghp123/SonaVoice/api/internal/database/repository-interfaces/mocks"
	"github.com/gianghp123/SonaVoice/api/internal/modules/session/dtos/req"
	"github.com/gianghp123/SonaVoice/api/internal/modules/session/dtos/res"
	"github.com/gianghp123/SonaVoice/api/internal/modules/session/services"
	svcMocks "github.com/gianghp123/SonaVoice/api/internal/modules/session/services/mocks"
	"github.com/jackc/pgx/v5/pgconn"
	"gorm.io/datatypes"
	"gorm.io/gorm"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
)

func setupSessionCtx(userID string) context.Context {
	ctx := context.Background()
	ctx = context.WithValue(ctx, enums.ContextKeyUserID, userID)
	return ctx
}

func TestOrchestratorService_CreateSession(t *testing.T) {
	configPayload := `{"enabled":true,"limits":{"user":{"daily_voice_seconds":300,"daily_request_count":100}}}`

	tests := []struct {
		name      string
		setupMock func(
			configSvc *svcMocks.SessionConfigService,
			sessionSvc *svcMocks.SessionService,
			speechSvc *svcMocks.SpeechProxyService,
			startConnSvc *svcMocks.StartConnectionService,
			quotaSvc *svcMocks.QuotaService,
			uow *svcMocks.UnitOfWork,
			provider *svcMocks.Provider,
			sessionRepo *repoMocks.SessionRepository,
		)
		wantErr bool
		errCode int
	}{
		{
			name: "success",
			setupMock: func(configSvc *svcMocks.SessionConfigService, sessionSvc *svcMocks.SessionService, speechSvc *svcMocks.SpeechProxyService, startConnSvc *svcMocks.StartConnectionService, quotaSvc *svcMocks.QuotaService, uow *svcMocks.UnitOfWork, provider *svcMocks.Provider, sessionRepo *repoMocks.SessionRepository) {
				configSvc.On("Get", mock.Anything).Return(&models.SessionConfig{Config: datatypes.JSON(configPayload)}, (*appErrors.AppError)(nil))
				quotaSvc.On("CheckRemaining", mock.Anything, "user-1", int64(300)).Return(int64(300), nil)
				provider.On("Session").Return(sessionRepo)
				sessionRepo.On("GetPendingByUserIDForUpdate", mock.Anything, "user-1").Return(nil, gorm.ErrRecordNotFound)
				sessionRepo.On("Create", mock.Anything, mock.Anything).Return(nil)
				uow.On("Do", mock.Anything, mock.Anything).Return(nil)
			},
		},
		{
			name: "quota exceeded",
			setupMock: func(configSvc *svcMocks.SessionConfigService, sessionSvc *svcMocks.SessionService, speechSvc *svcMocks.SpeechProxyService, startConnSvc *svcMocks.StartConnectionService, quotaSvc *svcMocks.QuotaService, uow *svcMocks.UnitOfWork, provider *svcMocks.Provider, sessionRepo *repoMocks.SessionRepository) {
				configSvc.On("Get", mock.Anything).Return(&models.SessionConfig{Config: datatypes.JSON(configPayload)}, (*appErrors.AppError)(nil))
				quotaSvc.On("CheckRemaining", mock.Anything, "user-1", int64(300)).Return(int64(0), nil)
			},
			wantErr: true,
			errCode: http.StatusForbidden,
		},
		{
			name: "create session conflict error",
			setupMock: func(configSvc *svcMocks.SessionConfigService, sessionSvc *svcMocks.SessionService, speechSvc *svcMocks.SpeechProxyService, startConnSvc *svcMocks.StartConnectionService, quotaSvc *svcMocks.QuotaService, uow *svcMocks.UnitOfWork, provider *svcMocks.Provider, sessionRepo *repoMocks.SessionRepository) {
				configSvc.On("Get", mock.Anything).Return(&models.SessionConfig{Config: datatypes.JSON(configPayload)}, (*appErrors.AppError)(nil))
				quotaSvc.On("CheckRemaining", mock.Anything, "user-1", int64(300)).Return(int64(300), nil)
				provider.On("Session").Return(sessionRepo)
				sessionRepo.On("GetPendingByUserIDForUpdate", mock.Anything, "user-1").Return(nil, gorm.ErrRecordNotFound)
				sessionRepo.On("Create", mock.Anything, mock.Anything).Return(&pgconn.PgError{Code: "23505"})
				uow.On("Do", mock.Anything, mock.Anything).Return(nil)
			},
			wantErr: true,
			errCode: http.StatusConflict,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			configSvc := new(svcMocks.SessionConfigService)
			sessionSvc := new(svcMocks.SessionService)
			speechSvc := new(svcMocks.SpeechProxyService)
			startConnSvc := new(svcMocks.StartConnectionService)
			quotaSvc := new(svcMocks.QuotaService)
			uow := new(svcMocks.UnitOfWork)
			provider := new(svcMocks.Provider)
			sessionRepo := new(repoMocks.SessionRepository)

			tt.setupMock(configSvc, sessionSvc, speechSvc, startConnSvc, quotaSvc, uow, provider, sessionRepo)

			uow.SetProvider(provider)

			svc := services.NewOrchestratorService(configSvc, sessionSvc, speechSvc, startConnSvc, quotaSvc, uow)
			ctx := setupSessionCtx("user-1")
			result, appErr := svc.CreateSession(ctx)

			if tt.wantErr {
				require.NotNil(t, appErr)
				assert.Equal(t, tt.errCode, appErr.Code)
			} else {
				require.Nil(t, appErr)
				require.NotNil(t, result)
			}
		})
	}
}

func TestOrchestratorService_StartConnection(t *testing.T) {
	tests := []struct {
		name      string
		setupMock func(
			configSvc *svcMocks.SessionConfigService,
			sessionSvc *svcMocks.SessionService,
			speechSvc *svcMocks.SpeechProxyService,
			startConnSvc *svcMocks.StartConnectionService,
			quotaSvc *svcMocks.QuotaService,
			uow *svcMocks.UnitOfWork,
		)
		wantErr bool
		errCode int
	}{
		{
			name: "success",
			setupMock: func(configSvc *svcMocks.SessionConfigService, sessionSvc *svcMocks.SessionService, speechSvc *svcMocks.SpeechProxyService, startConnSvc *svcMocks.StartConnectionService, quotaSvc *svcMocks.QuotaService, uow *svcMocks.UnitOfWork) {
				sessionSvc.On("Get", mock.Anything, "s1").Return(&models.Session{BaseModel: models.BaseModel{ID: "s1"}, UserID: "user-1", Status: enums.SessionStatusPending, MaxDuration: 300}, (*appErrors.AppError)(nil))
				startConnSvc.On("Start", mock.Anything, mock.Anything, "user-1").Return(&res.WebRTCConnectionRes{SessionID: "speech-s1"}, (*appErrors.AppError)(nil))
			},
		},
		{
			name: "session not startable",
			setupMock: func(configSvc *svcMocks.SessionConfigService, sessionSvc *svcMocks.SessionService, speechSvc *svcMocks.SpeechProxyService, startConnSvc *svcMocks.StartConnectionService, quotaSvc *svcMocks.QuotaService, uow *svcMocks.UnitOfWork) {
				sessionSvc.On("Get", mock.Anything, "s1").Return(&models.Session{BaseModel: models.BaseModel{ID: "s1"}, UserID: "user-1", Status: enums.SessionStatusInactive}, (*appErrors.AppError)(nil))
			},
			wantErr: true,
			errCode: http.StatusBadRequest,
		},
		{
			name: "start connection service error",
			setupMock: func(configSvc *svcMocks.SessionConfigService, sessionSvc *svcMocks.SessionService, speechSvc *svcMocks.SpeechProxyService, startConnSvc *svcMocks.StartConnectionService, quotaSvc *svcMocks.QuotaService, uow *svcMocks.UnitOfWork) {
				sessionSvc.On("Get", mock.Anything, "s1").Return(&models.Session{BaseModel: models.BaseModel{ID: "s1"}, UserID: "user-1", Status: enums.SessionStatusPending, MaxDuration: 300}, (*appErrors.AppError)(nil))
				startConnSvc.On("Start", mock.Anything, mock.Anything, "user-1").Return((*res.WebRTCConnectionRes)(nil), appErrors.Internal("speech engine error"))
			},
			wantErr: true,
			errCode: http.StatusInternalServerError,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			configSvc := new(svcMocks.SessionConfigService)
			sessionSvc := new(svcMocks.SessionService)
			speechSvc := new(svcMocks.SpeechProxyService)
			startConnSvc := new(svcMocks.StartConnectionService)
			quotaSvc := new(svcMocks.QuotaService)
			uow := new(svcMocks.UnitOfWork)

			tt.setupMock(configSvc, sessionSvc, speechSvc, startConnSvc, quotaSvc, uow)

			svc := services.NewOrchestratorService(configSvc, sessionSvc, speechSvc, startConnSvc, quotaSvc, uow)
			ctx := setupSessionCtx("user-1")
			result, appErr := svc.StartConnection(ctx, "s1")

			if tt.wantErr {
				require.NotNil(t, appErr)
				assert.Equal(t, tt.errCode, appErr.Code)
			} else {
				require.Nil(t, appErr)
				require.NotNil(t, result)
				assert.Equal(t, "speech-s1", result.SessionID)
			}
		})
	}
}

func TestOrchestratorService_ProxyOffer(t *testing.T) {
	responseBody := []byte(`{"sdp":"test-sdp"}`)
	errorBody := []byte(`{"error":"bad"}`)

	tests := []struct {
		name          string
		sessionID     string
		method        string
		body          []byte
		session       *models.Session
		setupProviderMock func(provider *svcMocks.Provider, sessionRepo *repoMocks.SessionRepository, speechSvc *svcMocks.SpeechProxyService)
		setupMocks func(
			sessionSvc *svcMocks.SessionService,
			speechSvc *svcMocks.SpeechProxyService,
		)
		wantErr       bool
		wantStatusCode int
	}{
		{
			name:      "successful proxy offer",
			sessionID: "speech-s1",
			method:    http.MethodPost,
			body:      []byte(`{"offer":"test"}`),
			session:   &models.Session{BaseModel: models.BaseModel{ID: "s1"}, UserID: "user-1", Status: enums.SessionStatusPending},
			setupProviderMock: func(provider *svcMocks.Provider, sessionRepo *repoMocks.SessionRepository, speechSvc *svcMocks.SpeechProxyService) {
				provider.On("Session").Return(sessionRepo)
				sessionRepo.On("GetForUpdate", mock.Anything, "s1").Return(&models.Session{BaseModel: models.BaseModel{ID: "s1"}, UserID: "user-1", Status: enums.SessionStatusPending}, nil)
				sessionRepo.On("SetSessionActive", mock.Anything, "s1", mock.Anything).Return(nil)
				speechSvc.On("ProxyOffer", mock.Anything, "speech-s1", http.MethodPost, mock.Anything).Return(responseBody, http.StatusOK, (*appErrors.AppError)(nil))
			},
			setupMocks: func(sessionSvc *svcMocks.SessionService, speechSvc *svcMocks.SpeechProxyService) {
				sessionSvc.On("GetBySpeechSessionID", mock.Anything, "speech-s1").Return(&models.Session{BaseModel: models.BaseModel{ID: "s1"}, UserID: "user-1", Status: enums.SessionStatusPending}, (*appErrors.AppError)(nil))
			},
			wantStatusCode: http.StatusOK,
		},
		{
			name:      "speech engine error",
			sessionID: "speech-s1",
			method:    http.MethodPost,
			body:      []byte(`{"offer":"test"}`),
			session:   &models.Session{BaseModel: models.BaseModel{ID: "s1"}, UserID: "user-1", Status: enums.SessionStatusPending},
			setupProviderMock: func(provider *svcMocks.Provider, sessionRepo *repoMocks.SessionRepository, speechSvc *svcMocks.SpeechProxyService) {
				provider.On("Session").Return(sessionRepo)
				sessionRepo.On("GetForUpdate", mock.Anything, "s1").Return(&models.Session{BaseModel: models.BaseModel{ID: "s1"}, UserID: "user-1", Status: enums.SessionStatusPending}, nil)
				sessionRepo.On("SetSessionFailed", mock.Anything, "s1").Return(nil)
				speechSvc.On("ProxyOffer", mock.Anything, "speech-s1", http.MethodPost, mock.Anything).Return([]byte{}, 0, appErrors.Internal("proxy error"))
			},
			setupMocks: func(sessionSvc *svcMocks.SessionService, speechSvc *svcMocks.SpeechProxyService) {
				sessionSvc.On("GetBySpeechSessionID", mock.Anything, "speech-s1").Return(&models.Session{BaseModel: models.BaseModel{ID: "s1"}, UserID: "user-1", Status: enums.SessionStatusPending}, (*appErrors.AppError)(nil))
			},
			wantErr: true,
		},
		{
			name:      "non-2xx from speech engine",
			sessionID: "speech-s1",
			method:    http.MethodPost,
			body:      []byte(`{"offer":"test"}`),
			session:   &models.Session{BaseModel: models.BaseModel{ID: "s1"}, UserID: "user-1", Status: enums.SessionStatusPending},
			setupProviderMock: func(provider *svcMocks.Provider, sessionRepo *repoMocks.SessionRepository, speechSvc *svcMocks.SpeechProxyService) {
				provider.On("Session").Return(sessionRepo)
				sessionRepo.On("GetForUpdate", mock.Anything, "s1").Return(&models.Session{BaseModel: models.BaseModel{ID: "s1"}, UserID: "user-1", Status: enums.SessionStatusPending}, nil)
				sessionRepo.On("SetSessionFailed", mock.Anything, "s1").Return(nil)
				speechSvc.On("ProxyOffer", mock.Anything, "speech-s1", http.MethodPost, mock.Anything).Return(errorBody, http.StatusBadGateway, appErrors.Internal("proxy error"))
			},
			setupMocks: func(sessionSvc *svcMocks.SessionService, speechSvc *svcMocks.SpeechProxyService) {
				sessionSvc.On("GetBySpeechSessionID", mock.Anything, "speech-s1").Return(&models.Session{BaseModel: models.BaseModel{ID: "s1"}, UserID: "user-1", Status: enums.SessionStatusPending}, (*appErrors.AppError)(nil))
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			configSvc := new(svcMocks.SessionConfigService)
			sessionSvc := new(svcMocks.SessionService)
			speechSvc := new(svcMocks.SpeechProxyService)
			startConnSvc := new(svcMocks.StartConnectionService)
			quotaSvc := new(svcMocks.QuotaService)

			sessionRepo := new(repoMocks.SessionRepository)
			provider := new(svcMocks.Provider)
			tt.setupProviderMock(provider, sessionRepo, speechSvc)
			tt.setupMocks(sessionSvc, speechSvc)

		uow := new(svcMocks.UnitOfWork)
		uow.SetProvider(provider)
		uow.On("Do", mock.Anything, mock.Anything).Return(nil)

		svc := services.NewOrchestratorService(configSvc, sessionSvc, speechSvc, startConnSvc, quotaSvc, uow)
		ctx := setupSessionCtx("user-1")
		respBody, statusCode, appErr := svc.ProxyOffer(ctx, tt.sessionID, tt.method, tt.body)

			if tt.wantErr {
				require.NotNil(t, appErr)
			} else {
				assert.Equal(t, tt.wantStatusCode, statusCode)
				if tt.wantStatusCode == http.StatusOK {
					require.Nil(t, appErr)
					require.NotNil(t, respBody)
				}
			}
		})
	}
}

func TestOrchestratorService_CloseSession(t *testing.T) {
	quotaDate := time.Date(2026, 5, 19, 0, 0, 0, 0, time.UTC)

	tests := []struct {
		name            string
		session         *models.Session
		actualUsage     int
		setupProviderMock func(provider *svcMocks.Provider, sessionRepo *repoMocks.SessionRepository, quotaRepo *repoMocks.UserQuotaRepository)
		wantErr         bool
		errCode         int
	}{
		{
			name:        "close active session with unused quota",
			session:     &models.Session{BaseModel: models.BaseModel{ID: "s1"}, UserID: "user-1", Status: enums.SessionStatusActive, MaxDuration: 300, QuotaDate: &quotaDate},
			actualUsage: 60,
			setupProviderMock: func(provider *svcMocks.Provider, sessionRepo *repoMocks.SessionRepository, quotaRepo *repoMocks.UserQuotaRepository) {
				provider.On("Session").Return(sessionRepo)
				provider.On("UserQuota").Return(quotaRepo)
				sessionRepo.On("GetForUpdate", mock.Anything, "s1").Return(&models.Session{BaseModel: models.BaseModel{ID: "s1"}, UserID: "user-1", Status: enums.SessionStatusActive, MaxDuration: 300, QuotaDate: &quotaDate}, nil)
				quotaRepo.On("Deduct", mock.Anything, "user-1", "voice", quotaDate, int64(60)).Return(nil)
				sessionRepo.On("SetActualUsage", mock.Anything, "s1", int64(60)).Return(nil)
				sessionRepo.On("SetSessionInactive", mock.Anything, "s1", mock.Anything).Return(nil)
			},
		},
		{
			name:        "actualUsage clamped to MaxDuration",
			session:     &models.Session{BaseModel: models.BaseModel{ID: "s1"}, UserID: "user-1", Status: enums.SessionStatusActive, MaxDuration: 300, QuotaDate: &quotaDate},
			actualUsage: 500,
			setupProviderMock: func(provider *svcMocks.Provider, sessionRepo *repoMocks.SessionRepository, quotaRepo *repoMocks.UserQuotaRepository) {
				provider.On("Session").Return(sessionRepo)
				provider.On("UserQuota").Return(quotaRepo)
				sessionRepo.On("GetForUpdate", mock.Anything, "s1").Return(&models.Session{BaseModel: models.BaseModel{ID: "s1"}, UserID: "user-1", Status: enums.SessionStatusActive, MaxDuration: 300, QuotaDate: &quotaDate}, nil)
				quotaRepo.On("Deduct", mock.Anything, "user-1", "voice", quotaDate, int64(300)).Return(nil)
				sessionRepo.On("SetActualUsage", mock.Anything, "s1", int64(300)).Return(nil)
				sessionRepo.On("SetSessionInactive", mock.Anything, "s1", mock.Anything).Return(nil)
			},
		},
		{
			name:        "deduct fails",
			session:     &models.Session{BaseModel: models.BaseModel{ID: "s1"}, UserID: "user-1", Status: enums.SessionStatusActive, MaxDuration: 300, QuotaDate: &quotaDate},
			actualUsage: 60,
			setupProviderMock: func(provider *svcMocks.Provider, sessionRepo *repoMocks.SessionRepository, quotaRepo *repoMocks.UserQuotaRepository) {
				provider.On("Session").Return(sessionRepo)
				provider.On("UserQuota").Return(quotaRepo)
				sessionRepo.On("GetForUpdate", mock.Anything, "s1").Return(&models.Session{BaseModel: models.BaseModel{ID: "s1"}, UserID: "user-1", Status: enums.SessionStatusActive, MaxDuration: 300, QuotaDate: &quotaDate}, nil)
				quotaRepo.On("Deduct", mock.Anything, "user-1", "voice", quotaDate, int64(60)).Return(assert.AnError)
			},
			wantErr: true,
			errCode: http.StatusInternalServerError,
		},
		{
			name:        "session already closed",
			session:     &models.Session{BaseModel: models.BaseModel{ID: "s1"}, UserID: "user-1", Status: enums.SessionStatusInactive},
			actualUsage: 60,
			setupProviderMock: func(provider *svcMocks.Provider, sessionRepo *repoMocks.SessionRepository, quotaRepo *repoMocks.UserQuotaRepository) {
				provider.On("Session").Return(sessionRepo)
				sessionRepo.On("GetForUpdate", mock.Anything, "s1").Return(&models.Session{BaseModel: models.BaseModel{ID: "s1"}, UserID: "user-1", Status: enums.SessionStatusInactive}, nil)
			},
			wantErr: true,
			errCode: http.StatusBadRequest,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			sessionRepo := new(repoMocks.SessionRepository)
			quotaRepo := new(repoMocks.UserQuotaRepository)
			provider := new(svcMocks.Provider)
			tt.setupProviderMock(provider, sessionRepo, quotaRepo)

			configSvc := new(svcMocks.SessionConfigService)
			sessionSvc := new(svcMocks.SessionService)
			speechSvc := new(svcMocks.SpeechProxyService)
			startConnSvc := new(svcMocks.StartConnectionService)
			quotaSvc := new(svcMocks.QuotaService)

		uow := new(svcMocks.UnitOfWork)
		uow.SetProvider(provider)
		uow.On("Do", mock.Anything, mock.Anything).Return(nil)

		svc := services.NewOrchestratorService(configSvc, sessionSvc, speechSvc, startConnSvc, quotaSvc, uow)
		ctx := setupSessionCtx("user-1")

		reqBody := &req.CloseSessionReq{
				SessionID:   "s1",
				ActualUsage: tt.actualUsage,
			}
			appErr := svc.CloseSession(ctx, reqBody)

			if tt.wantErr {
				require.NotNil(t, appErr)
				assert.Equal(t, tt.errCode, appErr.Code)
			} else {
				require.Nil(t, appErr)
			}
		})
	}
}

func TestOrchestratorService_CancelSession(t *testing.T) {
	tests := []struct {
		name            string
		sessionID       string
		session         *models.Session
		setupProviderMock func(provider *svcMocks.Provider, sessionRepo *repoMocks.SessionRepository)
		wantErr         bool
		errCode         int
	}{
		{
			name:      "cancel pending session",
			sessionID: "s1",
			session:   &models.Session{BaseModel: models.BaseModel{ID: "s1"}, UserID: "user-1", Status: enums.SessionStatusPending},
			setupProviderMock: func(provider *svcMocks.Provider, sessionRepo *repoMocks.SessionRepository) {
				provider.On("Session").Return(sessionRepo)
				sessionRepo.On("GetForUpdate", mock.Anything, "s1").Return(&models.Session{BaseModel: models.BaseModel{ID: "s1"}, UserID: "user-1", Status: enums.SessionStatusPending}, nil)
				sessionRepo.On("SetSessionInactive", mock.Anything, "s1", mock.Anything).Return(nil)
			},
		},
		{
			name:      "cancel active session",
			sessionID: "s1",
			session:   &models.Session{BaseModel: models.BaseModel{ID: "s1"}, UserID: "user-1", Status: enums.SessionStatusActive},
			setupProviderMock: func(provider *svcMocks.Provider, sessionRepo *repoMocks.SessionRepository) {
				provider.On("Session").Return(sessionRepo)
				sessionRepo.On("GetForUpdate", mock.Anything, "s1").Return(&models.Session{BaseModel: models.BaseModel{ID: "s1"}, UserID: "user-1", Status: enums.SessionStatusActive}, nil)
				sessionRepo.On("SetSessionInactive", mock.Anything, "s1", mock.Anything).Return(nil)
			},
		},
		{
			name:      "session already closed",
			sessionID: "s1",
			session:   &models.Session{BaseModel: models.BaseModel{ID: "s1"}, UserID: "user-1", Status: enums.SessionStatusInactive},
			setupProviderMock: func(provider *svcMocks.Provider, sessionRepo *repoMocks.SessionRepository) {
				provider.On("Session").Return(sessionRepo)
				sessionRepo.On("GetForUpdate", mock.Anything, "s1").Return(&models.Session{BaseModel: models.BaseModel{ID: "s1"}, UserID: "user-1", Status: enums.SessionStatusInactive}, nil)
			},
			wantErr: true,
			errCode: http.StatusBadRequest,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			sessionRepo := new(repoMocks.SessionRepository)
			provider := new(svcMocks.Provider)
			tt.setupProviderMock(provider, sessionRepo)

			configSvc := new(svcMocks.SessionConfigService)
			sessionSvc := new(svcMocks.SessionService)
			speechSvc := new(svcMocks.SpeechProxyService)
			startConnSvc := new(svcMocks.StartConnectionService)
			quotaSvc := new(svcMocks.QuotaService)

		uow := new(svcMocks.UnitOfWork)
		uow.SetProvider(provider)
		uow.On("Do", mock.Anything, mock.Anything).Return(nil)

		svc := services.NewOrchestratorService(configSvc, sessionSvc, speechSvc, startConnSvc, quotaSvc, uow)
		ctx := setupSessionCtx("user-1")
		appErr := svc.CancelSession(ctx, tt.sessionID)

			if tt.wantErr {
				require.NotNil(t, appErr)
				assert.Equal(t, tt.errCode, appErr.Code)
			} else {
				require.Nil(t, appErr)
			}
		})
	}
}

func TestOrchestratorService_ListSessions(t *testing.T) {
	tests := []struct {
		name      string
		query     req.SessionListQuery
		setupMock func(sessionSvc *svcMocks.SessionService)
		wantErr   bool
		wantCount int
	}{
		{
			name:  "success",
			query: req.SessionListQuery{Page: 1, Limit: 10},
			setupMock: func(sessionSvc *svcMocks.SessionService) {
				sessionSvc.On("List", mock.Anything, req.SessionListQuery{
					UserID: strPtr("user-1"),
					Page:   1,
					Limit:  10,
				}).Return(&response.PaginatedResult[*models.Session]{
					Data: []*models.Session{
						{BaseModel: models.BaseModel{ID: "s1"}, Status: enums.SessionStatusPending},
						{BaseModel: models.BaseModel{ID: "s2"}, Status: enums.SessionStatusActive},
					},
					Meta: &response.Meta{Page: 1, Limit: 10, Total: 2, TotalPages: 1},
				}, (*appErrors.AppError)(nil))
			},
			wantCount: 2,
		},
		{
			name:  "empty list",
			query: req.SessionListQuery{Page: 1, Limit: 10},
			setupMock: func(sessionSvc *svcMocks.SessionService) {
				sessionSvc.On("List", mock.Anything, req.SessionListQuery{
					UserID: strPtr("user-1"),
					Page:   1,
					Limit:  10,
				}).Return(&response.PaginatedResult[*models.Session]{
					Data: []*models.Session{},
					Meta: &response.Meta{Page: 1, Limit: 10, Total: 0, TotalPages: 0},
				}, (*appErrors.AppError)(nil))
			},
			wantCount: 0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			configSvc := new(svcMocks.SessionConfigService)
			sessionSvc := new(svcMocks.SessionService)
			speechSvc := new(svcMocks.SpeechProxyService)
			startConnSvc := new(svcMocks.StartConnectionService)
			quotaSvc := new(svcMocks.QuotaService)
			uow := new(svcMocks.UnitOfWork)

			tt.setupMock(sessionSvc)

			svc := services.NewOrchestratorService(configSvc, sessionSvc, speechSvc, startConnSvc, quotaSvc, uow)
			ctx := setupSessionCtx("user-1")
			result, appErr := svc.ListSessions(ctx, tt.query)

			if tt.wantErr {
				require.NotNil(t, appErr)
			} else {
				require.Nil(t, appErr)
				require.NotNil(t, result)
				assert.Equal(t, tt.wantCount, len(result.Data))
				assert.NotNil(t, result.Meta)
			}
		})
	}
}

func strPtr(s string) *string { return &s }
