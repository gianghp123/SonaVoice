package tests

import (
	"context"
	"testing"

	"github.com/gianghp123/SonaVoice/api/internal/core/enums"
	"github.com/gianghp123/SonaVoice/api/internal/database/models"
	repoMocks "github.com/gianghp123/SonaVoice/api/internal/database/repository-interfaces/mocks"
	svcMocks "github.com/gianghp123/SonaVoice/api/internal/modules/model-gateway/services/mocks"
	services "github.com/gianghp123/SonaVoice/api/internal/modules/model-gateway/services"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
)

func TestSessionJanitorService_CleanupStaleSessions(t *testing.T) {
	tests := []struct {
		name              string
		staleSessions     []*models.Session
		dailyVoiceSeconds int64
		setupMock         func(sessionRepo *repoMocks.SessionRepository, quotaSvc *svcMocks.SessionQuotaService, provider *svcMocks.Provider)
		wantErr           bool
	}{
		{
			name:              "no stale sessions",
			dailyVoiceSeconds: 3600,
			setupMock: func(sessionRepo *repoMocks.SessionRepository, quotaSvc *svcMocks.SessionQuotaService, provider *svcMocks.Provider) {
				sessionRepo.On("FindStaleByUserID", mock.Anything, "user-1", int64(60)).Return([]*models.Session{}, nil)
			},
		},
		{
			name:              "stale session with quota to release",
			dailyVoiceSeconds: 3600,
			setupMock: func(sessionRepo *repoMocks.SessionRepository, quotaSvc *svcMocks.SessionQuotaService, provider *svcMocks.Provider) {
				sessionRepo.On("FindStaleByUserID", mock.Anything, "user-1", int64(60)).Return([]*models.Session{
					{BaseModel: models.BaseModel{ID: "stale-1"}, UserID: "user-1", ReservedAmount: 300, DailyQuota: 3600, QuotaReleased: false, Status: enums.SessionStatusPending},
				}, nil)
				quotaSvc.On("ReleaseAll", mock.Anything, "user-1", int64(300)).Return(nil)
				sessionRepo.On("MarkStaleInactive", mock.Anything, []string{"stale-1"}).Return(nil)
			},
		},
		{
			name:              "quotas already released - no quota release call",
			dailyVoiceSeconds: 3600,
			setupMock: func(sessionRepo *repoMocks.SessionRepository, quotaSvc *svcMocks.SessionQuotaService, provider *svcMocks.Provider) {
				sessionRepo.On("FindStaleByUserID", mock.Anything, "user-1", int64(60)).Return([]*models.Session{
					{BaseModel: models.BaseModel{ID: "stale-2"}, UserID: "user-1", ReservedAmount: 300, DailyQuota: 3600, QuotaReleased: true, Status: enums.SessionStatusActive},
				}, nil)
				sessionRepo.On("MarkStaleInactive", mock.Anything, []string{"stale-2"}).Return(nil)
			},
		},
		{
			name:              "multiple stale sessions - sums quota",
			dailyVoiceSeconds: 3600,
			setupMock: func(sessionRepo *repoMocks.SessionRepository, quotaSvc *svcMocks.SessionQuotaService, provider *svcMocks.Provider) {
				sessionRepo.On("FindStaleByUserID", mock.Anything, "user-1", int64(60)).Return([]*models.Session{
					{BaseModel: models.BaseModel{ID: "s-1"}, UserID: "user-1", ReservedAmount: 200, DailyQuota: 3600, QuotaReleased: false, Status: enums.SessionStatusPending},
					{BaseModel: models.BaseModel{ID: "s-2"}, UserID: "user-1", ReservedAmount: 100, DailyQuota: 3600, QuotaReleased: false, Status: enums.SessionStatusPending},
					{BaseModel: models.BaseModel{ID: "s-3"}, UserID: "user-1", ReservedAmount: 50, DailyQuota: 3600, QuotaReleased: true, Status: enums.SessionStatusActive},
				}, nil)
				quotaSvc.On("ReleaseAll", mock.Anything, "user-1", int64(300)).Return(nil)
				sessionRepo.On("MarkStaleInactive", mock.Anything, []string{"s-1", "s-2", "s-3"}).Return(nil)
			},
		},
		{
			name:              "falls back to dailyVoiceSeconds when reserved/daily quota are zero",
			dailyVoiceSeconds: 3600,
			setupMock: func(sessionRepo *repoMocks.SessionRepository, quotaSvc *svcMocks.SessionQuotaService, provider *svcMocks.Provider) {
				sessionRepo.On("FindStaleByUserID", mock.Anything, "user-1", int64(60)).Return([]*models.Session{
					{BaseModel: models.BaseModel{ID: "s-1"}, UserID: "user-1", ReservedAmount: 0, DailyQuota: 0, QuotaReleased: false, Status: enums.SessionStatusPending},
				}, nil)
				quotaSvc.On("ReleaseAll", mock.Anything, "user-1", int64(3600)).Return(nil)
				sessionRepo.On("MarkStaleInactive", mock.Anything, []string{"s-1"}).Return(nil)
			},
		},
		{
			name:              "find stale sessions error",
			dailyVoiceSeconds: 3600,
			setupMock: func(sessionRepo *repoMocks.SessionRepository, quotaSvc *svcMocks.SessionQuotaService, provider *svcMocks.Provider) {
				sessionRepo.On("FindStaleByUserID", mock.Anything, "user-1", int64(60)).Return(nil, assert.AnError)
			},
			wantErr: true,
		},
		{
			name:              "mark stale inactive error",
			dailyVoiceSeconds: 3600,
			setupMock: func(sessionRepo *repoMocks.SessionRepository, quotaSvc *svcMocks.SessionQuotaService, provider *svcMocks.Provider) {
				sessionRepo.On("FindStaleByUserID", mock.Anything, "user-1", int64(60)).Return([]*models.Session{
					{BaseModel: models.BaseModel{ID: "s-1"}, UserID: "user-1", ReservedAmount: 300, DailyQuota: 3600, QuotaReleased: true, Status: enums.SessionStatusPending},
				}, nil)
				sessionRepo.On("MarkStaleInactive", mock.Anything, []string{"s-1"}).Return(assert.AnError)
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			quotaSvc := new(svcMocks.SessionQuotaService)
			sessionRepo := new(repoMocks.SessionRepository)
			provider := new(svcMocks.Provider)
			provider.On("Session").Return(sessionRepo)

			tt.setupMock(sessionRepo, quotaSvc, provider)

			svc := services.NewSessionJanitorService(quotaSvc)
			err := svc.CleanupStaleSessions(context.Background(), provider, "user-1", 60, tt.dailyVoiceSeconds)

			if tt.wantErr {
				require.Error(t, err)
				return
			}
			require.NoError(t, err)
		})
	}
}
