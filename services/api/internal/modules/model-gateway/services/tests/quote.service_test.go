package tests

import (
	"context"
	"errors"
	"net/http"
	"testing"
	"time"

	appErrors "github.com/gianghp123/SonaVoice/api/internal/core/errors"
	redisMocks "github.com/gianghp123/SonaVoice/api/internal/redis-client/mocks"
	"github.com/gianghp123/SonaVoice/api/internal/modules/model-gateway/services"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
)

func newQuoteService(t *testing.T, mockRedis *redisMocks.RedisClient) services.IQuoteService {
	t.Helper()
	mockRedis.On("LoadScript", mock.Anything, mock.Anything).Return(nil)
	return services.NewQuoteService(mockRedis)
}

func TestQuoteService_ReserveAllRemaining(t *testing.T) {
	tests := []struct {
		name        string
		userID      string
		dailyQuota  int64
		setupMock   func(mockRedis *redisMocks.RedisClient)
		wantErr     bool
		errContains string
		wantReserve int64
	}{
		{
			name:       "success",
			userID:     "user-1",
			dailyQuota: 3600,
			setupMock: func(mockRedis *redisMocks.RedisClient) {
				mockRedis.On("LoadScript", mock.Anything, mock.Anything).Return(nil)
				mockRedis.On("RunScript", mock.Anything, mock.Anything, mock.Anything, mock.Anything).
					Return(int64(300), nil)
			},
			wantReserve: 300,
		},
		{
			name:        "empty userID",
			userID:      "",
			dailyQuota:  3600,
			setupMock: func(mockRedis *redisMocks.RedisClient) {
				mockRedis.On("LoadScript", mock.Anything, mock.Anything).Return(nil)
			},
			wantErr:     true,
			errContains: "missing user id",
		},
		{
			name:       "non-positive dailyQuota returns 0 without error",
			userID:     "user-1",
			dailyQuota: 0,
			setupMock: func(mockRedis *redisMocks.RedisClient) {
				mockRedis.On("LoadScript", mock.Anything, mock.Anything).Return(nil)
			},
			wantReserve: 0,
		},
		{
			name:       "negative dailyQuota returns 0 without error",
			userID:     "user-1",
			dailyQuota: -1,
			setupMock: func(mockRedis *redisMocks.RedisClient) {
				mockRedis.On("LoadScript", mock.Anything, mock.Anything).Return(nil)
			},
			wantReserve: 0,
		},
		{
			name:       "redis RunScript error",
			userID:     "user-1",
			dailyQuota: 3600,
			setupMock: func(mockRedis *redisMocks.RedisClient) {
				mockRedis.On("LoadScript", mock.Anything, mock.Anything).Return(nil)
				mockRedis.On("RunScript", mock.Anything, mock.Anything, mock.Anything, mock.Anything).
					Return(nil, errors.New("redis error"))
			},
			wantErr: true,
		},
		{
			name:       "wrong result type",
			userID:     "user-1",
			dailyQuota: 3600,
			setupMock: func(mockRedis *redisMocks.RedisClient) {
				mockRedis.On("LoadScript", mock.Anything, mock.Anything).Return(nil)
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

			svc := newQuoteService(t, mockRedis)
			ctx := context.Background()

			result, err := svc.ReserveAllRemaining(ctx, tt.userID, tt.dailyQuota)

			if tt.wantErr {
				require.Error(t, err)
				if tt.errContains != "" {
					assert.Contains(t, err.Error(), tt.errContains)
				}
				return
			}
			require.NoError(t, err)
			assert.Equal(t, tt.wantReserve, result)
		})
	}
}

func TestQuoteService_Release(t *testing.T) {
	tests := []struct {
		name           string
		userID         string
		reservedAmount int64
		actualUsage    int64
		dailyQuota     int64
		setupMock      func(mockRedis *redisMocks.RedisClient)
		wantErr        bool
		errContains    string
	}{
		{
			name:           "success",
			userID:         "user-1",
			reservedAmount: 300,
			actualUsage:    120,
			dailyQuota:     3600,
			setupMock: func(mockRedis *redisMocks.RedisClient) {
				mockRedis.On("LoadScript", mock.Anything, mock.Anything).Return(nil)
				mockRedis.On("RunScript", mock.Anything, mock.Anything, mock.Anything, mock.Anything).
					Return(nil, nil)
			},
		},
		{
			name:           "empty userID",
			userID:         "",
			reservedAmount: 300,
			actualUsage:    120,
			dailyQuota:     3600,
			setupMock: func(mockRedis *redisMocks.RedisClient) {
				mockRedis.On("LoadScript", mock.Anything, mock.Anything).Return(nil)
			},
			wantErr:     true,
			errContains: "missing user id",
		},
		{
			name:           "negative reservedAmount clamped to 0",
			userID:         "user-1",
			reservedAmount: -1,
			actualUsage:    0,
			dailyQuota:     3600,
			setupMock: func(mockRedis *redisMocks.RedisClient) {
				mockRedis.On("LoadScript", mock.Anything, mock.Anything).Return(nil)
				mockRedis.On("RunScript", mock.Anything, mock.Anything, mock.Anything, mock.MatchedBy(func(args []interface{}) bool {
					return len(args) >= 4 && args[0].(int64) == 0
				})).Return(nil, nil)
			},
		},
		{
			name:           "actualUsage exceeds reservedAmount gets clamped",
			userID:         "user-1",
			reservedAmount: 100,
			actualUsage:    500,
			dailyQuota:     3600,
			setupMock: func(mockRedis *redisMocks.RedisClient) {
				mockRedis.On("LoadScript", mock.Anything, mock.Anything).Return(nil)
				mockRedis.On("RunScript", mock.Anything, mock.Anything, mock.Anything, mock.MatchedBy(func(args []interface{}) bool {
					return len(args) >= 4 && args[0].(int64) == 100 && args[1].(int64) == 100
				})).Return(nil, nil)
			},
		},
		{
			name:           "negative dailyQuota clamped to 0",
			userID:         "user-1",
			reservedAmount: 300,
			actualUsage:    120,
			dailyQuota:     -5,
			setupMock: func(mockRedis *redisMocks.RedisClient) {
				mockRedis.On("LoadScript", mock.Anything, mock.Anything).Return(nil)
				mockRedis.On("RunScript", mock.Anything, mock.Anything, mock.Anything, mock.MatchedBy(func(args []interface{}) bool {
					return len(args) >= 3 && args[2].(int64) == 0
				})).Return(nil, nil)
			},
		},
		{
			name:           "redis RunScript error",
			userID:         "user-1",
			reservedAmount: 300,
			actualUsage:    120,
			dailyQuota:     3600,
			setupMock: func(mockRedis *redisMocks.RedisClient) {
				mockRedis.On("LoadScript", mock.Anything, mock.Anything).Return(nil)
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

			svc := newQuoteService(t, mockRedis)
			ctx := context.Background()

			err := svc.Release(ctx, tt.userID, tt.reservedAmount, tt.actualUsage, tt.dailyQuota)

			if tt.wantErr {
				require.Error(t, err)
				if tt.errContains != "" {
					assert.Contains(t, err.Error(), tt.errContains)
				}
				return
			}
			require.NoError(t, err)
		})
	}
}

func TestQuoteService_AcquireSessionLock(t *testing.T) {
	tests := []struct {
		name        string
		userID      string
		ttl         time.Duration
		setupMock   func(mockRedis *redisMocks.RedisClient)
		wantErr     bool
		errContains string
		errCode     int
	}{
		{
			name:   "success",
			userID: "user-1",
			ttl:    30 * time.Second,
			setupMock: func(mockRedis *redisMocks.RedisClient) {
				mockRedis.On("LoadScript", mock.Anything, mock.Anything).Return(nil)
				mockRedis.On("SetNX", mock.Anything, mock.Anything, mock.Anything, mock.Anything).
					Return(true, nil)
			},
		},
		{
			name:        "empty userID",
			userID:      "",
			ttl:         30 * time.Second,
			setupMock: func(mockRedis *redisMocks.RedisClient) {
				mockRedis.On("LoadScript", mock.Anything, mock.Anything).Return(nil)
			},
			wantErr:     true,
			errContains: "missing user id",
		},
		{
			name:   "non-positive ttl",
			userID: "user-1",
			ttl:    0,
			setupMock: func(mockRedis *redisMocks.RedisClient) {
				mockRedis.On("LoadScript", mock.Anything, mock.Anything).Return(nil)
			},
			wantErr:     true,
			errContains: "invalid lock ttl",
		},
		{
			name:   "negative ttl",
			userID: "user-1",
			ttl:    -1 * time.Second,
			setupMock: func(mockRedis *redisMocks.RedisClient) {
				mockRedis.On("LoadScript", mock.Anything, mock.Anything).Return(nil)
			},
			wantErr:     true,
			errContains: "invalid lock ttl",
		},
		{
			name:   "redis SetNX error",
			userID: "user-1",
			ttl:    30 * time.Second,
			setupMock: func(mockRedis *redisMocks.RedisClient) {
				mockRedis.On("LoadScript", mock.Anything, mock.Anything).Return(nil)
				mockRedis.On("SetNX", mock.Anything, mock.Anything, mock.Anything, mock.Anything).
					Return(false, errors.New("redis error"))
			},
			wantErr: true,
		},
		{
			name:   "lock already exists",
			userID: "user-1",
			ttl:    30 * time.Second,
			setupMock: func(mockRedis *redisMocks.RedisClient) {
				mockRedis.On("LoadScript", mock.Anything, mock.Anything).Return(nil)
				mockRedis.On("SetNX", mock.Anything, mock.Anything, mock.Anything, mock.Anything).
					Return(false, nil)
			},
			wantErr: true,
			errCode: http.StatusConflict,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockRedis := new(redisMocks.RedisClient)
			tt.setupMock(mockRedis)

			svc := newQuoteService(t, mockRedis)
			ctx := context.Background()

			result, err := svc.AcquireSessionLock(ctx, tt.userID, tt.ttl)

			if tt.wantErr {
				require.Error(t, err)
				if tt.errContains != "" {
					assert.Contains(t, err.Error(), tt.errContains)
				}
				if tt.errCode != 0 {
					appErr, ok := err.(*appErrors.AppError)
					if ok {
						assert.Equal(t, tt.errCode, appErr.Code)
					}
				}
				return
			}
			require.NoError(t, err)
			assert.NotEmpty(t, result)
		})
	}
}

func TestQuoteService_ReleaseSessionLock(t *testing.T) {
	tests := []struct {
		name        string
		userID      string
		lockValue   string
		setupMock   func(mockRedis *redisMocks.RedisClient)
		wantErr     bool
		errContains string
	}{
		{
			name:      "success",
			userID:    "user-1",
			lockValue: "lock:user-1:123",
			setupMock: func(mockRedis *redisMocks.RedisClient) {
				mockRedis.On("LoadScript", mock.Anything, mock.Anything).Return(nil)
				mockRedis.On("RunScript", mock.Anything, mock.Anything, mock.Anything, mock.Anything).
					Return(nil, nil)
			},
		},
		{
			name:        "empty userID",
			userID:      "",
			lockValue:   "lock:val",
			setupMock: func(mockRedis *redisMocks.RedisClient) {
				mockRedis.On("LoadScript", mock.Anything, mock.Anything).Return(nil)
			},
			wantErr:     true,
			errContains: "missing user id",
		},
		{
			name:        "empty lockValue",
			userID:      "user-1",
			lockValue:   "",
			setupMock: func(mockRedis *redisMocks.RedisClient) {
				mockRedis.On("LoadScript", mock.Anything, mock.Anything).Return(nil)
			},
			wantErr:     true,
			errContains: "missing lock value",
		},
		{
			name:      "redis RunScript error",
			userID:    "user-1",
			lockValue: "lock:val",
			setupMock: func(mockRedis *redisMocks.RedisClient) {
				mockRedis.On("LoadScript", mock.Anything, mock.Anything).Return(nil)
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

			svc := newQuoteService(t, mockRedis)
			ctx := context.Background()

			err := svc.ReleaseSessionLock(ctx, tt.userID, tt.lockValue)

			if tt.wantErr {
				require.Error(t, err)
				if tt.errContains != "" {
					assert.Contains(t, err.Error(), tt.errContains)
				}
				return
			}
			require.NoError(t, err)
		})
	}
}

func TestQuoteService_ReserveQuota(t *testing.T) {
	tests := []struct {
		name       string
		userID     string
		dailyQuota int64
		setupMock  func(mockRedis *redisMocks.RedisClient)
		wantErr    bool
		errCode    int
	}{
		{
			name:       "success",
			userID:     "user-1",
			dailyQuota: 3600,
			setupMock: func(mockRedis *redisMocks.RedisClient) {
				mockRedis.On("LoadScript", mock.Anything, mock.Anything).Return(nil)
				mockRedis.On("RunScript", mock.Anything, mock.Anything, mock.Anything, mock.Anything).
					Return(int64(300), nil)
			},
		},
		{
			name:       "quota exceeded",
			userID:     "user-1",
			dailyQuota: 3600,
			setupMock: func(mockRedis *redisMocks.RedisClient) {
				mockRedis.On("LoadScript", mock.Anything, mock.Anything).Return(nil)
				mockRedis.On("RunScript", mock.Anything, mock.Anything, mock.Anything, mock.Anything).
					Return(int64(0), nil)
			},
			wantErr: true,
			errCode: http.StatusForbidden,
		},
		{
			name:       "reserve fails",
			userID:     "user-1",
			dailyQuota: 3600,
			setupMock: func(mockRedis *redisMocks.RedisClient) {
				mockRedis.On("LoadScript", mock.Anything, mock.Anything).Return(nil)
				mockRedis.On("RunScript", mock.Anything, mock.Anything, mock.Anything, mock.Anything).
					Return(nil, errors.New("redis error"))
			},
			wantErr: true,
			errCode: http.StatusInternalServerError,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockRedis := new(redisMocks.RedisClient)
			tt.setupMock(mockRedis)

			svc := newQuoteService(t, mockRedis)
			ctx := context.Background()

			reserved, cleanup, appErr := svc.ReserveQuota(ctx, tt.userID, tt.dailyQuota)

			if tt.wantErr {
				require.NotNil(t, appErr)
				assert.Equal(t, tt.errCode, appErr.Code)
				return
			}
			require.Nil(t, appErr)
			assert.Equal(t, int64(300), reserved)
			assert.NotNil(t, cleanup)

			cleanup(false)
		})
	}
}

func TestQuoteService_ReserveQuota_CleanupNotDoubleRelease(t *testing.T) {
	mockRedis := new(redisMocks.RedisClient)
	mockRedis.On("LoadScript", mock.Anything, mock.Anything).Return(nil)
	mockRedis.On("RunScript", mock.Anything, mock.Anything, mock.Anything, mock.Anything).
		Return(int64(300), nil).Once()

	svc := newQuoteService(t, mockRedis)
	ctx := context.Background()

	reserved, cleanup, appErr := svc.ReserveQuota(ctx, "user-1", 3600)
	require.Nil(t, appErr)
	assert.Equal(t, int64(300), reserved)

	cleanup(true)

	cleanup(false)
	mockRedis.AssertNumberOfCalls(t, "RunScript", 1)
}
