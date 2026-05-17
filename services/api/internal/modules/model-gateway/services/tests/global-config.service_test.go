package tests

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"testing"

	"github.com/gianghp123/SonaVoice/api/internal/database/models"
	repoMocks "github.com/gianghp123/SonaVoice/api/internal/database/repository-interfaces/mocks"
	"github.com/gianghp123/SonaVoice/api/internal/modules/model-gateway/dtos"
	"github.com/gianghp123/SonaVoice/api/internal/modules/model-gateway/services"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
	"gorm.io/datatypes"
)

func makeConfigJSON(t *testing.T, cfg dtos.GlobalConfig) datatypes.JSON {
	t.Helper()
	b, err := json.Marshal(cfg.Config)
	require.NoError(t, err)
	return datatypes.JSON(b)
}

func TestGlobalConfigService_Get(t *testing.T) {
	cfg := dtos.GlobalConfig{
		Config: dtos.ConfigPayload{
			Enabled: true,
			Limits: dtos.LimitsConfig{
				User: dtos.UserLimitConfig{DailyVoiceSeconds: 300},
			},
		},
	}
	configModel := &models.GlobalConfig{Config: makeConfigJSON(t, cfg)}

	tests := []struct {
		name      string
		setupMock func(mockRepo *repoMocks.GlobalConfigRepository)
		wantErr   bool
		errCode   int
		wantCfg   *models.GlobalConfig
	}{
		{
			name: "success",
			setupMock: func(mockRepo *repoMocks.GlobalConfigRepository) {
				mockRepo.On("Get", mock.Anything).Return(configModel, nil)
			},
			wantCfg: configModel,
		},
		{
			name: "repo get error",
			setupMock: func(mockRepo *repoMocks.GlobalConfigRepository) {
				mockRepo.On("Get", mock.Anything).Return(nil, errors.New("db error"))
			},
			wantErr: true,
			errCode: http.StatusInternalServerError,
		},
		{
			name: "nil model returns nil",
			setupMock: func(mockRepo *repoMocks.GlobalConfigRepository) {
				mockRepo.On("Get", mock.Anything).Return((*models.GlobalConfig)(nil), nil)
			},
			wantCfg: (*models.GlobalConfig)(nil),
		},
		{
			name: "empty config bytes",
			setupMock: func(mockRepo *repoMocks.GlobalConfigRepository) {
				mockRepo.On("Get", mock.Anything).Return(&models.GlobalConfig{Config: datatypes.JSON{}}, nil)
			},
			wantCfg: &models.GlobalConfig{Config: datatypes.JSON{}},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockRepo := new(repoMocks.GlobalConfigRepository)
			tt.setupMock(mockRepo)

			svc := services.NewGlobalConfigService(mockRepo)
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

func TestGlobalConfigService_Update(t *testing.T) {
	existingCfg := dtos.GlobalConfig{
		Config: dtos.ConfigPayload{
			Enabled: false,
			Limits: dtos.LimitsConfig{
				User: dtos.UserLimitConfig{DailyVoiceSeconds: 100},
			},
		},
	}
	existingModel := &models.GlobalConfig{Config: makeConfigJSON(t, existingCfg)}

	newCfg := dtos.GlobalConfig{
		Config: dtos.ConfigPayload{
			Enabled: true,
			Limits: dtos.LimitsConfig{
				User: dtos.UserLimitConfig{DailyVoiceSeconds: 500},
			},
		},
	}

	tests := []struct {
		name      string
		newCfg    *dtos.GlobalConfig
		setupMock func(mockRepo *repoMocks.GlobalConfigRepository)
		wantErr   bool
		errCode   int
	}{
		{
			name:   "success",
			newCfg: &newCfg,
			setupMock: func(mockRepo *repoMocks.GlobalConfigRepository) {
				mockRepo.On("Get", mock.Anything).Return(existingModel, nil)
				mockRepo.On("Save", mock.Anything, mock.Anything).Return(nil)
			},
		},
		{
			name:   "repo get error",
			newCfg: &newCfg,
			setupMock: func(mockRepo *repoMocks.GlobalConfigRepository) {
				mockRepo.On("Get", mock.Anything).Return(nil, errors.New("db error"))
			},
			wantErr: true,
			errCode: http.StatusInternalServerError,
		},
		{
			name:   "repo save error",
			newCfg: &newCfg,
			setupMock: func(mockRepo *repoMocks.GlobalConfigRepository) {
				mockRepo.On("Get", mock.Anything).Return(existingModel, nil)
				mockRepo.On("Save", mock.Anything, mock.Anything).Return(errors.New("save failed"))
			},
			wantErr: true,
			errCode: http.StatusInternalServerError,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockRepo := new(repoMocks.GlobalConfigRepository)
			tt.setupMock(mockRepo)

			svc := services.NewGlobalConfigService(mockRepo)
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
