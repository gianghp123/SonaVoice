package tests

import (
	"context"
	"errors"
	"net/http"
	"testing"
	"time"

	"github.com/gianghp123/SonaVoice/api/internal/core/enums"
	"github.com/gianghp123/SonaVoice/api/internal/database/models"
	repoMocks "github.com/gianghp123/SonaVoice/api/internal/database/repository-interfaces/mocks"
	"github.com/gianghp123/SonaVoice/api/internal/modules/model-gateway/services"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
)

func setupSessionCtx(userID string) context.Context {
	ctx := context.Background()
	ctx = context.WithValue(ctx, enums.ContextKeyUserID, userID)
	return ctx
}

func TestSessionService_CreateSession(t *testing.T) {
	tests := []struct {
		name      string
		userID    string
		setupMock func(mockRepo *repoMocks.SessionRepository)
		wantErr   bool
		errCode   int
	}{
		{
			name:   "success",
			userID: "user-1",
			setupMock: func(mockRepo *repoMocks.SessionRepository) {
				mockRepo.On("Create", mock.Anything, mock.MatchedBy(func(m *models.Session) bool {
					m.ID = "generated-id"
					return m.UserID == "user-1" && m.Status == enums.SessionStatusPending
				})).Return(nil)
			},
		},
		{
			name:   "repo create error",
			userID: "user-1",
			setupMock: func(mockRepo *repoMocks.SessionRepository) {
				mockRepo.On("Create", mock.Anything, mock.Anything).Return(errors.New("db error"))
			},
			wantErr: true,
			errCode: http.StatusInternalServerError,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockRepo := new(repoMocks.SessionRepository)
			tt.setupMock(mockRepo)

			svc := services.NewSessionService(mockRepo)
			ctx := setupSessionCtx(tt.userID)

			result, appErr := svc.CreateSession(ctx)

			if tt.wantErr {
				require.NotNil(t, appErr)
				assert.Equal(t, tt.errCode, appErr.Code)
				return
			}
			require.Nil(t, appErr)
			assert.NotNil(t, result)
			assert.Equal(t, tt.userID, result.UserID)
			assert.NotEmpty(t, result.ID)
			mockRepo.AssertExpectations(t)
		})
	}
}

func TestSessionService_GetSession(t *testing.T) {
	session := &models.Session{BaseModel: models.BaseModel{ID: "s1", CreatedAt: time.Now()}, UserID: "user-1", Status: enums.SessionStatusPending}
	tests := []struct {
		name      string
		userID    string
		sessionID string
		setupMock func(mockRepo *repoMocks.SessionRepository)
		wantErr   bool
		errCode   int
	}{
		{
			name:      "success",
			userID:    "user-1",
			sessionID: "s1",
			setupMock: func(mockRepo *repoMocks.SessionRepository) {
				mockRepo.On("Get", mock.Anything, "s1").Return(session, nil)
			},
		},
		{
			name:      "repo get error",
			userID:    "user-1",
			sessionID: "s1",
			setupMock: func(mockRepo *repoMocks.SessionRepository) {
				mockRepo.On("Get", mock.Anything, "s1").Return(nil, errors.New("not found"))
			},
			wantErr: true,
			errCode: http.StatusInternalServerError,
		},
		{
			name:      "ownership check fails",
			userID:    "user-2",
			sessionID: "s1",
			setupMock: func(mockRepo *repoMocks.SessionRepository) {
				mockRepo.On("Get", mock.Anything, "s1").Return(session, nil)
			},
			wantErr: true,
			errCode: http.StatusForbidden,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockRepo := new(repoMocks.SessionRepository)
			tt.setupMock(mockRepo)

			svc := services.NewSessionService(mockRepo)
			ctx := setupSessionCtx(tt.userID)

			result, appErr := svc.GetSession(ctx, tt.sessionID)

			if tt.wantErr {
				require.NotNil(t, appErr)
				assert.Equal(t, tt.errCode, appErr.Code)
				return
			}
			require.Nil(t, appErr)
			assert.Equal(t, "s1", result.ID)
			mockRepo.AssertExpectations(t)
		})
	}
}

func TestSessionService_GetSessionBySpeechSessionID(t *testing.T) {
	session := &models.Session{BaseModel: models.BaseModel{ID: "s1", CreatedAt: time.Now()}, UserID: "user-1", Status: enums.SessionStatusActive}
	tests := []struct {
		name            string
		userID          string
		speechSessionID string
		setupMock       func(mockRepo *repoMocks.SessionRepository)
		wantErr         bool
		errCode         int
	}{
		{
			name:            "success",
			userID:          "user-1",
			speechSessionID: "speech-1",
			setupMock: func(mockRepo *repoMocks.SessionRepository) {
				mockRepo.On("GetBySpeechSessionID", mock.Anything, "speech-1").Return(session, nil)
			},
		},
		{
			name:            "repo error",
			userID:          "user-1",
			speechSessionID: "speech-1",
			setupMock: func(mockRepo *repoMocks.SessionRepository) {
				mockRepo.On("GetBySpeechSessionID", mock.Anything, "speech-1").Return(nil, errors.New("db error"))
			},
			wantErr: true,
			errCode: http.StatusInternalServerError,
		},
		{
			name:            "ownership check fails",
			userID:          "user-2",
			speechSessionID: "speech-1",
			setupMock: func(mockRepo *repoMocks.SessionRepository) {
				mockRepo.On("GetBySpeechSessionID", mock.Anything, "speech-1").Return(session, nil)
			},
			wantErr: true,
			errCode: http.StatusForbidden,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockRepo := new(repoMocks.SessionRepository)
			tt.setupMock(mockRepo)

			svc := services.NewSessionService(mockRepo)
			ctx := setupSessionCtx(tt.userID)

			result, appErr := svc.GetSessionBySpeechSessionID(ctx, tt.speechSessionID)

			if tt.wantErr {
				require.NotNil(t, appErr)
				assert.Equal(t, tt.errCode, appErr.Code)
				return
			}
			require.Nil(t, appErr)
			assert.Equal(t, "s1", result.ID)
			mockRepo.AssertExpectations(t)
		})
	}
}

func TestSessionService_SetSpeechSessionID(t *testing.T) {
	session := &models.Session{BaseModel: models.BaseModel{ID: "s1"}, UserID: "user-1"}
	tests := []struct {
		name            string
		sessionID       string
		speechSessionID string
		setupMock       func(mockRepo *repoMocks.SessionRepository)
		wantErr         bool
		errCode         int
	}{
		{
			name:            "success",
			sessionID:       "s1",
			speechSessionID: "speech-new",
			setupMock: func(mockRepo *repoMocks.SessionRepository) {
				mockRepo.On("Get", mock.Anything, "s1").Return(session, nil)
				mockRepo.On("Update", mock.Anything, mock.MatchedBy(func(m *models.Session) bool {
					return m.SpeechSessionID == "speech-new"
				})).Return(nil)
			},
		},
		{
			name:            "repo get error",
			sessionID:       "s1",
			speechSessionID: "speech-new",
			setupMock: func(mockRepo *repoMocks.SessionRepository) {
				mockRepo.On("Get", mock.Anything, "s1").Return(nil, errors.New("not found"))
			},
			wantErr: true,
			errCode: http.StatusInternalServerError,
		},
		{
			name:            "repo update error",
			sessionID:       "s1",
			speechSessionID: "speech-new",
			setupMock: func(mockRepo *repoMocks.SessionRepository) {
				mockRepo.On("Get", mock.Anything, "s1").Return(session, nil)
				mockRepo.On("Update", mock.Anything, mock.Anything).Return(errors.New("update failed"))
			},
			wantErr: true,
			errCode: http.StatusInternalServerError,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockRepo := new(repoMocks.SessionRepository)
			tt.setupMock(mockRepo)

			svc := services.NewSessionService(mockRepo)
			ctx := context.Background()

			appErr := svc.SetSpeechSessionID(ctx, tt.sessionID, tt.speechSessionID)

			if tt.wantErr {
				require.NotNil(t, appErr)
				assert.Equal(t, tt.errCode, appErr.Code)
				return
			}
			require.Nil(t, appErr)
			mockRepo.AssertExpectations(t)
		})
	}
}

func TestSessionService_MarkSessionFailed(t *testing.T) {
	session := &models.Session{BaseModel: models.BaseModel{ID: "s1"}, Status: enums.SessionStatusPending}
	tests := []struct {
		name      string
		sessionID string
		setupMock func(mockRepo *repoMocks.SessionRepository)
		wantErr   bool
		errCode   int
	}{
		{
			name:      "success",
			sessionID: "s1",
			setupMock: func(mockRepo *repoMocks.SessionRepository) {
				mockRepo.On("Get", mock.Anything, "s1").Return(session, nil)
				mockRepo.On("Update", mock.Anything, mock.MatchedBy(func(m *models.Session) bool {
					return m.Status == enums.SessionStatusFailed
				})).Return(nil)
			},
		},
		{
			name:      "repo get error",
			sessionID: "s1",
			setupMock: func(mockRepo *repoMocks.SessionRepository) {
				mockRepo.On("Get", mock.Anything, "s1").Return(nil, errors.New("not found"))
			},
			wantErr: true,
			errCode: http.StatusInternalServerError,
		},
		{
			name:      "repo update error",
			sessionID: "s1",
			setupMock: func(mockRepo *repoMocks.SessionRepository) {
				mockRepo.On("Get", mock.Anything, "s1").Return(session, nil)
				mockRepo.On("Update", mock.Anything, mock.Anything).Return(errors.New("update failed"))
			},
			wantErr: true,
			errCode: http.StatusInternalServerError,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockRepo := new(repoMocks.SessionRepository)
			tt.setupMock(mockRepo)

			svc := services.NewSessionService(mockRepo)
			ctx := context.Background()

			appErr := svc.MarkSessionFailed(ctx, tt.sessionID)

			if tt.wantErr {
				require.NotNil(t, appErr)
				assert.Equal(t, tt.errCode, appErr.Code)
				return
			}
			require.Nil(t, appErr)
			mockRepo.AssertExpectations(t)
		})
	}
}

func TestSessionService_MarkSessionActive(t *testing.T) {
	session := &models.Session{BaseModel: models.BaseModel{ID: "s1"}, Status: enums.SessionStatusPending}
	tests := []struct {
		name      string
		sessionID string
		setupMock func(mockRepo *repoMocks.SessionRepository)
		wantErr   bool
		errCode   int
	}{
		{
			name:      "success",
			sessionID: "s1",
			setupMock: func(mockRepo *repoMocks.SessionRepository) {
				mockRepo.On("Get", mock.Anything, "s1").Return(session, nil)
				mockRepo.On("Update", mock.Anything, mock.MatchedBy(func(m *models.Session) bool {
					return m.Status == enums.SessionStatusActive
				})).Return(nil)
			},
		},
		{
			name:      "repo get error",
			sessionID: "s1",
			setupMock: func(mockRepo *repoMocks.SessionRepository) {
				mockRepo.On("Get", mock.Anything, "s1").Return(nil, errors.New("not found"))
			},
			wantErr: true,
			errCode: http.StatusInternalServerError,
		},
		{
			name:      "repo update error",
			sessionID: "s1",
			setupMock: func(mockRepo *repoMocks.SessionRepository) {
				mockRepo.On("Get", mock.Anything, "s1").Return(session, nil)
				mockRepo.On("Update", mock.Anything, mock.Anything).Return(errors.New("update failed"))
			},
			wantErr: true,
			errCode: http.StatusInternalServerError,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockRepo := new(repoMocks.SessionRepository)
			tt.setupMock(mockRepo)

			svc := services.NewSessionService(mockRepo)
			ctx := context.Background()

			appErr := svc.MarkSessionActive(ctx, tt.sessionID)

			if tt.wantErr {
				require.NotNil(t, appErr)
				assert.Equal(t, tt.errCode, appErr.Code)
				return
			}
			require.Nil(t, appErr)
			mockRepo.AssertExpectations(t)
		})
	}
}

func TestSessionService_MarkSessionInactive(t *testing.T) {
	session := &models.Session{BaseModel: models.BaseModel{ID: "s1"}, Status: enums.SessionStatusActive}
	tests := []struct {
		name      string
		sessionID string
		setupMock func(mockRepo *repoMocks.SessionRepository)
		wantErr   bool
		errCode   int
	}{
		{
			name:      "success",
			sessionID: "s1",
			setupMock: func(mockRepo *repoMocks.SessionRepository) {
				mockRepo.On("Get", mock.Anything, "s1").Return(session, nil)
				mockRepo.On("Update", mock.Anything, mock.MatchedBy(func(m *models.Session) bool {
					return m.Status == enums.SessionStatusInactive
				})).Return(nil)
			},
		},
		{
			name:      "repo get error",
			sessionID: "s1",
			setupMock: func(mockRepo *repoMocks.SessionRepository) {
				mockRepo.On("Get", mock.Anything, "s1").Return(nil, errors.New("not found"))
			},
			wantErr: true,
			errCode: http.StatusInternalServerError,
		},
		{
			name:      "repo update error",
			sessionID: "s1",
			setupMock: func(mockRepo *repoMocks.SessionRepository) {
				mockRepo.On("Get", mock.Anything, "s1").Return(session, nil)
				mockRepo.On("Update", mock.Anything, mock.Anything).Return(errors.New("update failed"))
			},
			wantErr: true,
			errCode: http.StatusInternalServerError,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockRepo := new(repoMocks.SessionRepository)
			tt.setupMock(mockRepo)

			svc := services.NewSessionService(mockRepo)
			ctx := context.Background()

			appErr := svc.MarkSessionInactive(ctx, tt.sessionID)

			if tt.wantErr {
				require.NotNil(t, appErr)
				assert.Equal(t, tt.errCode, appErr.Code)
				return
			}
			require.Nil(t, appErr)
			mockRepo.AssertExpectations(t)
		})
	}
}
