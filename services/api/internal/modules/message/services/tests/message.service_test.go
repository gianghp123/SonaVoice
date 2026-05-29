package tests

import (
	"context"
	"net/http"
	"testing"
	"time"

	"github.com/gianghp123/SonaVoice/api/internal/core/enums"
	"github.com/gianghp123/SonaVoice/api/internal/core/response"
	"github.com/gianghp123/SonaVoice/api/internal/database/models"
	repoMocks "github.com/gianghp123/SonaVoice/api/internal/database/repository-interfaces/mocks"
	"github.com/gianghp123/SonaVoice/api/internal/modules/message/dtos/req"
	"github.com/gianghp123/SonaVoice/api/internal/modules/message/services"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
	"gorm.io/gorm"
)

func setupMessageCtx(userID string) context.Context {
	ctx := context.Background()
	return context.WithValue(ctx, enums.ContextKeyUserID, userID)
}

func TestMessageService_List(t *testing.T) {
	now := time.Now().UTC()
	session := &models.Session{
		BaseModel: models.BaseModel{ID: "session-1", CreatedAt: now, UpdatedAt: now},
		UserID:    "user-1",
		Status:    enums.SessionStatusActive,
	}

	messages := []*models.Message{
		{
			BaseModel: models.BaseModel{ID: "msg-1", CreatedAt: now, UpdatedAt: now},
			SessionID: "session-1",
			Role:      enums.MessageRoleUser,
			Transcript: "hello",
		},
		{
			BaseModel: models.BaseModel{ID: "msg-2", CreatedAt: now, UpdatedAt: now},
			SessionID: "session-1",
			Role:      enums.MessageRoleAssistant,
			Transcript: "hi there",
		},
	}

	paginated := &response.PaginatedResult[*models.Message]{
		Data: messages,
		Meta: response.NewMeta(1, 10, 2),
	}

	tests := []struct {
		name      string
		sessionID string
		query     req.MessageListQuery
		userID    string
		setupMock func(msgRepo *repoMocks.MessageRepository, sessionRepo *repoMocks.SessionRepository)
		wantErr   bool
		errCode   int
		wantCount int
	}{
		{
			name:      "success",
			sessionID: "session-1",
			query:     req.MessageListQuery{Page: 1, Limit: 10},
			userID:    "user-1",
			setupMock: func(msgRepo *repoMocks.MessageRepository, sessionRepo *repoMocks.SessionRepository) {
				sessionRepo.On("Get", mock.Anything, "session-1").Return(session, nil)
				msgRepo.On("ListBySessionID", mock.Anything, "session-1", mock.Anything).Return(paginated, nil)
			},
			wantCount: 2,
		},
		{
			name:      "session not found",
			sessionID: "session-unknown",
			query:     req.MessageListQuery{Page: 1, Limit: 10},
			userID:    "user-1",
			setupMock: func(msgRepo *repoMocks.MessageRepository, sessionRepo *repoMocks.SessionRepository) {
				sessionRepo.On("Get", mock.Anything, "session-unknown").Return(nil, gorm.ErrRecordNotFound)
			},
			wantErr: true,
			errCode: http.StatusNotFound,
		},
		{
			name:      "forbidden - not session owner",
			sessionID: "session-1",
			query:     req.MessageListQuery{Page: 1, Limit: 10},
			userID:    "user-2",
			setupMock: func(msgRepo *repoMocks.MessageRepository, sessionRepo *repoMocks.SessionRepository) {
				sessionRepo.On("Get", mock.Anything, "session-1").Return(session, nil)
			},
			wantErr: true,
			errCode: http.StatusForbidden,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			msgRepo := new(repoMocks.MessageRepository)
			sessionRepo := new(repoMocks.SessionRepository)

			tt.setupMock(msgRepo, sessionRepo)

			svc := services.NewMessageService(msgRepo, sessionRepo)
			ctx := setupMessageCtx(tt.userID)
			result, appErr := svc.List(ctx, tt.sessionID, tt.query)

			if tt.wantErr {
				require.NotNil(t, appErr)
				assert.Equal(t, tt.errCode, appErr.Code)
				return
			}

			require.Nil(t, appErr)
			require.NotNil(t, result)
			assert.Equal(t, tt.wantCount, len(result.Data))
			msgRepo.AssertExpectations(t)
			sessionRepo.AssertExpectations(t)
		})
	}
}

func TestMessageService_Create(t *testing.T) {
	now := time.Now().UTC()
	session := &models.Session{
		BaseModel: models.BaseModel{ID: "session-1", CreatedAt: now, UpdatedAt: now},
		UserID:    "user-1",
		Status:    enums.SessionStatusActive,
	}

	tests := []struct {
		name      string
		sessionID string
		body      *req.CreateMessagesReq
		userID    string
		setupMock func(msgRepo *repoMocks.MessageRepository, sessionRepo *repoMocks.SessionRepository)
		wantErr   bool
		errCode   int
		wantCount int
	}{
		{
			name:      "success",
			sessionID: "session-1",
			body: &req.CreateMessagesReq{
				Messages: []req.MessageItem{
					{Role: enums.MessageRoleUser, Transcript: "hello"},
					{Role: enums.MessageRoleAssistant, Transcript: "hi"},
				},
			},
			userID: "user-1",
			setupMock: func(msgRepo *repoMocks.MessageRepository, sessionRepo *repoMocks.SessionRepository) {
				sessionRepo.On("Get", mock.Anything, "session-1").Return(session, nil)
				msgRepo.On("CreateBatch", mock.Anything, mock.Anything).Return(nil)
			},
			wantCount: 2,
		},
		{
			name:      "session not found",
			sessionID: "session-unknown",
			body: &req.CreateMessagesReq{
				Messages: []req.MessageItem{
					{Role: enums.MessageRoleUser, Transcript: "hello"},
				},
			},
			userID: "user-1",
			setupMock: func(msgRepo *repoMocks.MessageRepository, sessionRepo *repoMocks.SessionRepository) {
				sessionRepo.On("Get", mock.Anything, "session-unknown").Return(nil, gorm.ErrRecordNotFound)
			},
			wantErr: true,
			errCode: http.StatusNotFound,
		},
		{
			name:      "forbidden - not session owner",
			sessionID: "session-1",
			body: &req.CreateMessagesReq{
				Messages: []req.MessageItem{
					{Role: enums.MessageRoleUser, Transcript: "hello"},
				},
			},
			userID: "user-2",
			setupMock: func(msgRepo *repoMocks.MessageRepository, sessionRepo *repoMocks.SessionRepository) {
				sessionRepo.On("Get", mock.Anything, "session-1").Return(session, nil)
			},
			wantErr: true,
			errCode: http.StatusForbidden,
		},
		{
			name:      "empty messages - bad request",
			sessionID: "session-1",
			body: &req.CreateMessagesReq{
				Messages: []req.MessageItem{},
			},
			userID: "user-1",
			setupMock: func(msgRepo *repoMocks.MessageRepository, sessionRepo *repoMocks.SessionRepository) {
			},
			wantErr: true,
			errCode: http.StatusBadRequest,
		},
		{
			name:      "db error",
			sessionID: "session-1",
			body: &req.CreateMessagesReq{
				Messages: []req.MessageItem{
					{Role: enums.MessageRoleUser, Transcript: "hello"},
				},
			},
			userID: "user-1",
			setupMock: func(msgRepo *repoMocks.MessageRepository, sessionRepo *repoMocks.SessionRepository) {
				sessionRepo.On("Get", mock.Anything, "session-1").Return(session, nil)
				msgRepo.On("CreateBatch", mock.Anything, mock.Anything).Return(gorm.ErrInvalidData)
			},
			wantErr: true,
			errCode: http.StatusInternalServerError,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			msgRepo := new(repoMocks.MessageRepository)
			sessionRepo := new(repoMocks.SessionRepository)

			tt.setupMock(msgRepo, sessionRepo)

			svc := services.NewMessageService(msgRepo, sessionRepo)
			ctx := setupMessageCtx(tt.userID)
			result, appErr := svc.Create(ctx, tt.sessionID, tt.body)

			if tt.wantErr {
				require.NotNil(t, appErr)
				assert.Equal(t, tt.errCode, appErr.Code)
				return
			}

			require.Nil(t, appErr)
			require.NotNil(t, result)
			assert.Equal(t, tt.wantCount, len(result))
			msgRepo.AssertExpectations(t)
			sessionRepo.AssertExpectations(t)
		})
	}
}
