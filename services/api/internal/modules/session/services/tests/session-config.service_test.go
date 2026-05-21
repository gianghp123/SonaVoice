package tests

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"testing"

	"github.com/gianghp123/SonaVoice/api/internal/database/models"
	repoMocks "github.com/gianghp123/SonaVoice/api/internal/database/repository-interfaces/mocks"
	"github.com/gianghp123/SonaVoice/api/internal/modules/session/dtos"
	"github.com/gianghp123/SonaVoice/api/internal/modules/session/services"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
	"gorm.io/datatypes"
)

func makeConfigJSON(t *testing.T, cfg dtos.SessionConfig) datatypes.JSON {
	t.Helper()
	b, err := json.Marshal(cfg.Config)
	require.NoError(t, err)
	return datatypes.JSON(b)
}

func TestSessionConfigService_Get(t *testing.T) {
	cfg := dtos.SessionConfig{
		Config: dtos.ConfigPayload{
			Enabled: true,
			Limits: dtos.LimitsConfig{
				User: dtos.UserLimitConfig{DailyVoiceSeconds: 300},
			},
		},
	}
	configModel := &models.SessionConfig{Config: makeConfigJSON(t, cfg)}

	tests := []struct {
		name      string
		setupMock func(mockRepo *repoMocks.SessionConfigRepository)
		wantErr   bool
		errCode   int
		wantCfg   *models.SessionConfig
	}{
		{
			name: "success",
			setupMock: func(mockRepo *repoMocks.SessionConfigRepository) {
				mockRepo.On("Get", mock.Anything).Return(configModel, nil)
			},
			wantCfg: configModel,
		},
		{
			name: "repo get error",
			setupMock: func(mockRepo *repoMocks.SessionConfigRepository) {
				mockRepo.On("Get", mock.Anything).Return(nil, errors.New("db error"))
			},
			wantErr: true,
			errCode: http.StatusInternalServerError,
		},
		{
			name: "nil model returns nil",
			setupMock: func(mockRepo *repoMocks.SessionConfigRepository) {
				mockRepo.On("Get", mock.Anything).Return((*models.SessionConfig)(nil), nil)
			},
			wantCfg: (*models.SessionConfig)(nil),
		},
		{
			name: "empty config bytes",
			setupMock: func(mockRepo *repoMocks.SessionConfigRepository) {
				mockRepo.On("Get", mock.Anything).Return(&models.SessionConfig{Config: datatypes.JSON{}}, nil)
			},
			wantCfg: &models.SessionConfig{Config: datatypes.JSON{}},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockRepo := new(repoMocks.SessionConfigRepository)
			tt.setupMock(mockRepo)

			svc := services.NewSessionConfigService(mockRepo)
			ctx := context.Background()

			result, appErr := svc.Get(ctx)

			if tt.wantErr {
				require.NotNil(t, appErr)
				assert.Equal(t, tt.errCode, appErr.Code)
				return
			}
			require.Nil(t, appErr)
			assert.Equal(t, tt.wantCfg, result)
			mockRepo.AssertExpectations(t)
		})
	}
}

func TestSessionConfigService_Update(t *testing.T) {
	existingCfg := dtos.SessionConfig{
		Config: dtos.ConfigPayload{
			Enabled: false,
			Limits: dtos.LimitsConfig{
				User: dtos.UserLimitConfig{DailyVoiceSeconds: 100},
			},
		},
	}
	existingModel := &models.SessionConfig{Config: makeConfigJSON(t, existingCfg)}

	newCfg := dtos.SessionConfig{
		Config: dtos.ConfigPayload{
			Enabled: true,
			Limits: dtos.LimitsConfig{
				User: dtos.UserLimitConfig{DailyVoiceSeconds: 500},
			},
		},
	}

	tests := []struct {
		name      string
		newCfg    *dtos.SessionConfig
		setupMock func(mockRepo *repoMocks.SessionConfigRepository)
		wantErr   bool
		errCode   int
	}{
		{
			name:   "success",
			newCfg: &newCfg,
			setupMock: func(mockRepo *repoMocks.SessionConfigRepository) {
				mockRepo.On("Get", mock.Anything).Return(existingModel, nil)
				mockRepo.On("Save", mock.Anything, mock.Anything).Return(nil)
			},
		},
		{
			name:   "repo get error",
			newCfg: &newCfg,
			setupMock: func(mockRepo *repoMocks.SessionConfigRepository) {
				mockRepo.On("Get", mock.Anything).Return(nil, errors.New("db error"))
			},
			wantErr: true,
			errCode: http.StatusInternalServerError,
		},
		{
			name:   "repo save error",
			newCfg: &newCfg,
			setupMock: func(mockRepo *repoMocks.SessionConfigRepository) {
				mockRepo.On("Get", mock.Anything).Return(existingModel, nil)
				mockRepo.On("Save", mock.Anything, mock.Anything).Return(errors.New("save failed"))
			},
			wantErr: true,
			errCode: http.StatusInternalServerError,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockRepo := new(repoMocks.SessionConfigRepository)
			tt.setupMock(mockRepo)

			svc := services.NewSessionConfigService(mockRepo)
			ctx := context.Background()

			result, appErr := svc.Update(ctx, tt.newCfg)

			if tt.wantErr {
				require.NotNil(t, appErr)
				assert.Equal(t, tt.errCode, appErr.Code)
				return
			}
			require.Nil(t, appErr)
			require.NotNil(t, result)
			mockRepo.AssertExpectations(t)
		})
	}
}
