package tests

import (
	"context"
	"testing"

	"github.com/gianghp123/SonaVoice/api/internal/database/models"
	repoMocks "github.com/gianghp123/SonaVoice/api/internal/database/repository-interfaces/mocks"
	services "github.com/gianghp123/SonaVoice/api/internal/modules/model-gateway/services"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
)

func TestSessionQuotaService_Reserve(t *testing.T) {
	tests := []struct {
		name        string
		dailyQuota  int
		reservedAmt int64
		reserveErr  error
		wantErr     bool
		errCode     int
		wantAmount  int64
	}{
		{
			name: "success",
			dailyQuota: 3600, reservedAmt: 300, reserveErr: nil,
			wantErr: false, wantAmount: 300,
		},
		{
			name: "internal error",
			dailyQuota: 3600, reservedAmt: 0, reserveErr: assert.AnError,
			wantErr: true, errCode: 500,
		},
		{
			name: "quota exceeded",
			dailyQuota: 3600, reservedAmt: 0, reserveErr: nil,
			wantErr: true, errCode: 403,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			quotaRepo := new(repoMocks.UserQuotaRepository)
			quotaRepo.On("ReserveAll", mock.Anything, "user-1", "voice", mock.Anything, int64(tt.dailyQuota)).Return(tt.reservedAmt, tt.reserveErr)

			svc := services.NewSessionQuotaService(quotaRepo)
			amount, appErr := svc.Reserve(context.Background(), "user-1", tt.dailyQuota)

			if tt.wantErr {
				require.NotNil(t, appErr)
				assert.Equal(t, tt.errCode, appErr.Code)
				return
			}
			require.Nil(t, appErr)
			assert.Equal(t, tt.wantAmount, amount)
		})
	}
}

func TestSessionQuotaService_ReleaseAll(t *testing.T) {
	tests := []struct {
		name        string
		amount      int64
		releaseErr  error
		wantErr     bool
	}{
		{name: "success", amount: 300, releaseErr: nil, wantErr: false},
		{name: "skips when amount is zero", amount: 0, releaseErr: nil, wantErr: false},
		{name: "skips when amount is negative", amount: -1, releaseErr: nil, wantErr: false},
		{name: "release error", amount: 300, releaseErr: assert.AnError, wantErr: true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			quotaRepo := new(repoMocks.UserQuotaRepository)
			if tt.amount > 0 {
				quotaRepo.On("Release", mock.Anything, "user-1", "voice", mock.Anything, tt.amount).Return(tt.releaseErr)
			}

			svc := services.NewSessionQuotaService(quotaRepo)
			appErr := svc.ReleaseAll(context.Background(), "user-1", tt.amount)

			if tt.wantErr {
				require.NotNil(t, appErr)
				return
			}
			require.Nil(t, appErr)
		})
	}
}

func TestSessionQuotaService_ReleaseWithActualUsage(t *testing.T) {
	tests := []struct {
		name        string
		session     *models.Session
		actualUsage int64
		setupMock   func(quotaRepo *repoMocks.UserQuotaRepository)
		wantErr     bool
	}{
		{
			name: "releases unused = reserved - actual",
			session: &models.Session{
				BaseModel: models.BaseModel{ID: "s1"}, UserID: "user-1",
				ReservedAmount: 300, DailyQuota: 3600, QuotaReleased: false,
			},
			actualUsage: 60,
			setupMock: func(qr *repoMocks.UserQuotaRepository) {
				qr.On("Release", mock.Anything, "user-1", "voice", mock.Anything, int64(240)).Return(nil)
			},
		},
		{
			name: "clamps actual exceeding reserved",
			session: &models.Session{
				BaseModel: models.BaseModel{ID: "s1"}, UserID: "user-1",
				ReservedAmount: 200, DailyQuota: 3600, QuotaReleased: false,
			},
			actualUsage: 500,
			setupMock: func(qr *repoMocks.UserQuotaRepository) {
				qr.On("Release", mock.Anything, "user-1", "voice", mock.Anything, int64(0)).Return(nil)
			},
		},
		{
			name: "skips when quota already released",
			session: &models.Session{
				BaseModel: models.BaseModel{ID: "s1"}, UserID: "user-1",
				ReservedAmount: 300, DailyQuota: 3600, QuotaReleased: true,
			},
			actualUsage: 60,
			setupMock: func(qr *repoMocks.UserQuotaRepository) {},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			quotaRepo := new(repoMocks.UserQuotaRepository)
			tt.setupMock(quotaRepo)

			svc := services.NewSessionQuotaService(quotaRepo)
			appErr := svc.ReleaseWithActualUsage(context.Background(), tt.session, tt.actualUsage)

			if tt.wantErr {
				require.NotNil(t, appErr)
				return
			}
			require.Nil(t, appErr)
		})
	}
}
