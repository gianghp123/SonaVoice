package quota

import (
	"context"
	"errors"
	"testing"

	redisMocks "github.com/gianghp123/SonaVoice/api/internal/redis-client/mocks"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
)

func newQuotaService(t *testing.T, mockRedis *redisMocks.RedisClient) IQuotaService {
	t.Helper()
	mockRedis.On("LoadScript", mock.Anything, mock.Anything).Return(nil)
	return NewQuotaService(mockRedis)
}

func TestQuotaService_ReserveAll(t *testing.T) {
	cfg := QuotaConfig{Key: "voice", DailyLimit: 3600}
	tests := []struct {
		name       string
		userID     string
		cfg        QuotaConfig
		setupMock  func(mockRedis *redisMocks.RedisClient)
		wantErr    bool
		wantReserve int64
	}{
		{
			name:   "success",
			userID: "user-1",
			cfg:    cfg,
			setupMock: func(mockRedis *redisMocks.RedisClient) {
				mockRedis.On("RunScript", mock.Anything, mock.Anything, mock.Anything, mock.Anything).
					Return(int64(300), nil)
			},
			wantReserve: 300,
		},
		{
			name:   "empty userID",
			userID: "",
			cfg:    cfg,
			setupMock: func(mockRedis *redisMocks.RedisClient) {
			},
			wantErr: true,
		},
		{
			name:   "non-positive dailyQuota returns 0 without error",
			userID: "user-1",
			cfg:    QuotaConfig{Key: "voice", DailyLimit: 0},
			setupMock: func(mockRedis *redisMocks.RedisClient) {
			},
			wantReserve: 0,
		},
		{
			name:   "negative dailyQuota returns 0 without error",
			userID: "user-1",
			cfg:    QuotaConfig{Key: "voice", DailyLimit: -1},
			setupMock: func(mockRedis *redisMocks.RedisClient) {
			},
			wantReserve: 0,
		},
		{
			name:   "redis RunScript error",
			userID: "user-1",
			cfg:    cfg,
			setupMock: func(mockRedis *redisMocks.RedisClient) {
				mockRedis.On("RunScript", mock.Anything, mock.Anything, mock.Anything, mock.Anything).
					Return(nil, errors.New("redis error"))
			},
			wantErr: true,
		},
		{
			name:   "wrong result type",
			userID: "user-1",
			cfg:    cfg,
			setupMock: func(mockRedis *redisMocks.RedisClient) {
				mockRedis.On("RunScript", mock.Anything, mock.Anything, mock.Anything, mock.Anything).
					Return("not an int64", nil)
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockRedis := new(redisMocks.RedisClient)
			tt.setupMock(mockRedis)
			svc := newQuotaService(t, mockRedis)
			ctx := context.Background()
			result, err := svc.ReserveAll(ctx, tt.userID, tt.cfg)
			if tt.wantErr {
				require.Error(t, err)
				return
			}
			require.NoError(t, err)
			assert.Equal(t, tt.wantReserve, result)
		})
	}
}

func TestQuotaService_Release(t *testing.T) {
	cfg := QuotaConfig{Key: "voice", DailyLimit: 3600}
	tests := []struct {
		name           string
		userID         string
		cfg            QuotaConfig
		reservedAmount int64
		actualUsage    int64
		setupMock      func(mockRedis *redisMocks.RedisClient)
		wantErr        bool
	}{
		{
			name:           "success",
			userID:         "user-1",
			cfg:            cfg,
			reservedAmount: 300,
			actualUsage:    120,
			setupMock: func(mockRedis *redisMocks.RedisClient) {
				mockRedis.On("RunScript", mock.Anything, mock.Anything, mock.Anything, mock.Anything).
					Return(nil, nil)
			},
		},
		{
			name:           "empty userID",
			userID:         "",
			cfg:            cfg,
			reservedAmount: 300,
			actualUsage:    120,
			setupMock: func(mockRedis *redisMocks.RedisClient) {
			},
			wantErr: true,
		},
		{
			name:           "negative reservedAmount clamped to 0",
			userID:         "user-1",
			cfg:            cfg,
			reservedAmount: -1,
			actualUsage:    0,
			setupMock: func(mockRedis *redisMocks.RedisClient) {
				mockRedis.On("RunScript", mock.Anything, mock.Anything, mock.Anything, mock.MatchedBy(func(args []interface{}) bool {
					return len(args) >= 4 && args[0].(int64) == 0
				})).Return(nil, nil)
			},
		},
		{
			name:           "actualUsage exceeds reservedAmount gets clamped",
			userID:         "user-1",
			cfg:            cfg,
			reservedAmount: 100,
			actualUsage:    500,
			setupMock: func(mockRedis *redisMocks.RedisClient) {
				mockRedis.On("RunScript", mock.Anything, mock.Anything, mock.Anything, mock.MatchedBy(func(args []interface{}) bool {
					return len(args) >= 4 && args[0].(int64) == 100 && args[1].(int64) == 100
				})).Return(nil, nil)
			},
		},
		{
			name:           "negative dailyQuota clamped to 0",
			userID:         "user-1",
			cfg:            QuotaConfig{Key: "voice", DailyLimit: -5},
			reservedAmount: 300,
			actualUsage:    120,
			setupMock: func(mockRedis *redisMocks.RedisClient) {
				mockRedis.On("RunScript", mock.Anything, mock.Anything, mock.Anything, mock.MatchedBy(func(args []interface{}) bool {
					return len(args) >= 3 && args[2].(int64) == 0
				})).Return(nil, nil)
			},
		},
		{
			name:           "redis RunScript error",
			userID:         "user-1",
			cfg:            cfg,
			reservedAmount: 300,
			actualUsage:    120,
			setupMock: func(mockRedis *redisMocks.RedisClient) {
				mockRedis.On("RunScript", mock.Anything, mock.Anything, mock.Anything, mock.Anything).
					Return(nil, errors.New("redis error"))
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockRedis := new(redisMocks.RedisClient)
			tt.setupMock(mockRedis)
			svc := newQuotaService(t, mockRedis)
			ctx := context.Background()
			err := svc.Release(ctx, tt.userID, tt.cfg, tt.reservedAmount, tt.actualUsage)
			if tt.wantErr {
				require.Error(t, err)
				return
			}
			require.NoError(t, err)
		})
	}
}


