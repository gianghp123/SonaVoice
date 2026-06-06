package tests

import (
	"context"
	"net/http"
	"testing"

	"github.com/gianghp123/SonaVoice/api/internal/core/enums"
	"github.com/gianghp123/SonaVoice/api/internal/database/models"
	repoMocks "github.com/gianghp123/SonaVoice/api/internal/database/repository-interfaces/mocks"
	"github.com/gianghp123/SonaVoice/api/internal/modules/user_profile/dtos/req"
	"github.com/gianghp123/SonaVoice/api/internal/modules/user_profile/services"
	svcMocks "github.com/gianghp123/SonaVoice/api/internal/modules/user_profile/services/mocks"
	clerkMocks "github.com/gianghp123/SonaVoice/api/internal/clerk-client/mocks"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
	"gorm.io/datatypes"
	"gorm.io/gorm"
)

func setupCtx(userID string) context.Context {
	ctx := context.Background()
	ctx = context.WithValue(ctx, enums.ContextKeyUserID, userID)
	return ctx
}

func TestUserProfileService_GetByUserID(t *testing.T) {
	tests := []struct {
		name      string
		setupMock func(profileRepo *repoMocks.UserProfileRepository)
		wantErr   bool
		errCode   int
	}{
		{
			name: "success",
			setupMock: func(profileRepo *repoMocks.UserProfileRepository) {
				profileRepo.On("GetByUserID", mock.Anything, "user-1").Return(&models.UserProfile{
					UserID:       "user-1",
					DisplayName:  "John",
					EnglishLevel: "intermediate",
				}, nil)
			},
		},
		{
			name: "repo error",
			setupMock: func(profileRepo *repoMocks.UserProfileRepository) {
				profileRepo.On("GetByUserID", mock.Anything, "user-1").Return(nil, gorm.ErrRecordNotFound)
			},
			wantErr: true,
			errCode: http.StatusNotFound,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			profileRepo := new(repoMocks.UserProfileRepository)
			tt.setupMock(profileRepo)

			uow := new(svcMocks.UnitOfWork)
			clerkClient := new(clerkMocks.ClerkClient)
			svc := services.NewUserProfileService(profileRepo, uow, clerkClient)
			ctx := setupCtx("user-1")
			result, appErr := svc.GetByUserID(ctx, "user-1")

			if tt.wantErr {
				require.NotNil(t, appErr)
				assert.Equal(t, tt.errCode, appErr.Code)
			} else {
				require.Nil(t, appErr)
				require.NotNil(t, result)
				assert.Equal(t, "user-1", result.UserID)
			}
		})
	}
}

func TestUserProfileService_Update(t *testing.T) {
	existingPrefs := datatypes.JSON([]byte(`{"native_language":"vi","improvement_goals":["fluency"],"topics":["travel"]}`))

	tests := []struct {
		name      string
		body      *req.UpdateProfileReq
		setupMock func(profileRepo *repoMocks.UserProfileRepository)
		wantErr   bool
		errCode   int
	}{
		{
			name: "success all fields",
			body: &req.UpdateProfileReq{
				DisplayName:  strPtr("John Updated"),
				EnglishLevel: strPtr("advanced"),
			},
			setupMock: func(profileRepo *repoMocks.UserProfileRepository) {
				profileRepo.On("GetByUserID", mock.Anything, "user-1").Return(&models.UserProfile{
					UserID:       "user-1",
					DisplayName:  "John",
					EnglishLevel: "intermediate",
					Preferences:  existingPrefs,
				}, nil)
				profileRepo.On("Upsert", mock.Anything, mock.MatchedBy(func(p *models.UserProfile) bool {
					return p.DisplayName == "John Updated" && p.EnglishLevel == "advanced"
				})).Return(nil)
			},
		},
		{
			name: "partial update display name only",
			body: &req.UpdateProfileReq{
				DisplayName: strPtr("John Partial"),
			},
			setupMock: func(profileRepo *repoMocks.UserProfileRepository) {
				profileRepo.On("GetByUserID", mock.Anything, "user-1").Return(&models.UserProfile{
					UserID:       "user-1",
					DisplayName:  "John",
					EnglishLevel: "intermediate",
					Preferences:  existingPrefs,
				}, nil)
				profileRepo.On("Upsert", mock.Anything, mock.MatchedBy(func(p *models.UserProfile) bool {
					return p.DisplayName == "John Partial" && p.EnglishLevel == "intermediate"
				})).Return(nil)
			},
		},
		{
			name: "update preferences",
			body: &req.UpdateProfileReq{
				DisplayName:    strPtr("John"),
				NativeLanguage: strPtr("en"),
			},
			setupMock: func(profileRepo *repoMocks.UserProfileRepository) {
				profileRepo.On("GetByUserID", mock.Anything, "user-1").Return(&models.UserProfile{
					UserID:       "user-1",
					DisplayName:  "John",
					EnglishLevel: "intermediate",
					Preferences:  existingPrefs,
				}, nil)
				profileRepo.On("Upsert", mock.Anything, mock.Anything).Return(nil)
			},
		},
		{
			name: "get profile fails",
			body: &req.UpdateProfileReq{
				DisplayName: strPtr("John"),
			},
			setupMock: func(profileRepo *repoMocks.UserProfileRepository) {
				profileRepo.On("GetByUserID", mock.Anything, "user-1").Return(nil, gorm.ErrRecordNotFound)
			},
			wantErr: true,
			errCode: http.StatusNotFound,
		},
		{
			name: "upsert fails",
			body: &req.UpdateProfileReq{
				DisplayName: strPtr("John"),
			},
			setupMock: func(profileRepo *repoMocks.UserProfileRepository) {
				profileRepo.On("GetByUserID", mock.Anything, "user-1").Return(&models.UserProfile{
					UserID:       "user-1",
					DisplayName:  "John",
					EnglishLevel: "intermediate",
					Preferences:  existingPrefs,
				}, nil)
				profileRepo.On("Upsert", mock.Anything, mock.Anything).Return(gorm.ErrInvalidDB)
			},
			wantErr: true,
			errCode: http.StatusInternalServerError,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			profileRepo := new(repoMocks.UserProfileRepository)
			tt.setupMock(profileRepo)

			uow := new(svcMocks.UnitOfWork)
			clerkClient := new(clerkMocks.ClerkClient)
			svc := services.NewUserProfileService(profileRepo, uow, clerkClient)
			ctx := setupCtx("user-1")
			appErr := svc.Update(ctx, "user-1", tt.body)

			if tt.wantErr {
				require.NotNil(t, appErr)
				assert.Equal(t, tt.errCode, appErr.Code)
			} else {
				require.Nil(t, appErr)
			}
		})
	}
}

func TestUserProfileService_Onboard(t *testing.T) {
	tests := []struct {
		name      string
		body      *req.UpsertProfileReq
		setupMock func(profileRepo *repoMocks.UserProfileRepository, uow *svcMocks.UnitOfWork, provider *svcMocks.Provider, clerkClient *clerkMocks.ClerkClient)
		wantErr   bool
		errCode   int
	}{
		{
			name: "uow error",
			body: &req.UpsertProfileReq{
				DisplayName:  "John",
				EnglishLevel: "intermediate",
			},
			setupMock: func(profileRepo *repoMocks.UserProfileRepository, uow *svcMocks.UnitOfWork, provider *svcMocks.Provider, clerkClient *clerkMocks.ClerkClient) {
				uow.On("Do", mock.Anything, mock.Anything).Return(gorm.ErrInvalidTransaction)
			},
			wantErr: true,
			errCode: http.StatusInternalServerError,
		},
		{
			name: "upsert fails inside transaction",
			body: &req.UpsertProfileReq{
				DisplayName:  "John",
				EnglishLevel: "intermediate",
			},
			setupMock: func(profileRepo *repoMocks.UserProfileRepository, uow *svcMocks.UnitOfWork, provider *svcMocks.Provider, clerkClient *clerkMocks.ClerkClient) {
				profileRepo.On("Upsert", mock.Anything, mock.Anything).Return(gorm.ErrInvalidDB)
				provider.On("UserProfile").Return(profileRepo)
				uow.On("Do", mock.Anything, mock.Anything).Return(nil)
			},
			wantErr: true,
			errCode: http.StatusInternalServerError,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			profileRepo := new(repoMocks.UserProfileRepository)
			uow := new(svcMocks.UnitOfWork)
			provider := new(svcMocks.Provider)
			clerkClient := new(clerkMocks.ClerkClient)

			tt.setupMock(profileRepo, uow, provider, clerkClient)

			if tt.name == "upsert fails inside transaction" {
				uow.SetProvider(provider)
			}

			svc := services.NewUserProfileService(profileRepo, uow, clerkClient)
			ctx := setupCtx("user-1")
			appErr := svc.Onboard(ctx, "user-1", tt.body)

			require.NotNil(t, appErr)
			assert.Equal(t, tt.errCode, appErr.Code)
		})
	}
}

func strPtr(s string) *string { return &s }
