package services

import (
	"context"
	"encoding/json"

	clerkuser "github.com/clerk/clerk-sdk-go/v2/user"
	"github.com/getsentry/sentry-go"
	"github.com/gianghp123/SonaVoice/api/internal/core/enums"
	"github.com/gianghp123/SonaVoice/api/internal/core/errors"
	zapLogger "github.com/gianghp123/SonaVoice/api/internal/core/zap-logger"
	"github.com/gianghp123/SonaVoice/api/internal/database/models"
	repository_interfaces "github.com/gianghp123/SonaVoice/api/internal/database/repository-interfaces"
	"github.com/gianghp123/SonaVoice/api/internal/database/transaction"
	"github.com/gianghp123/SonaVoice/api/internal/modules/user_profile/dtos/req"
	"gorm.io/datatypes"
)

type IUserProfileService interface {
	GetByUserID(ctx context.Context, userID string) (*models.UserProfile, *errors.AppError)
	Upsert(ctx context.Context, userID string, body *req.UpsertProfileReq) (*models.UserProfile, *errors.AppError)
}

type userProfileService struct {
	profileRepo repository_interfaces.IUserProfileRepository
	uow         transaction.UnitOfWork
}

func NewUserProfileService(
	profileRepo repository_interfaces.IUserProfileRepository,
	uow transaction.UnitOfWork,
) IUserProfileService {
	return &userProfileService{
		profileRepo: profileRepo,
		uow:         uow,
	}
}

func (s *userProfileService) GetByUserID(ctx context.Context, userID string) (*models.UserProfile, *errors.AppError) {
	logger := zapLogger.S()

	profile, err := s.profileRepo.GetByUserID(ctx, userID)
	if err != nil {
		logger.Errorw("Failed to get user profile", "userId", userID, "error", err)
		sentry.CaptureException(err)
		return nil, errors.MapRepoError(err)
	}

	return profile, nil
}

func (s *userProfileService) Upsert(ctx context.Context, userID string, body *req.UpsertProfileReq) (*models.UserProfile, *errors.AppError) {
	logger := zapLogger.S()

	preferences := map[string]interface{}{
		"native_language":        body.NativeLanguage,
		"improvement_goals":      body.ImprovementGoals,
		"topics":                 body.Topics,
		"custom_topics":          body.CustomTopics,
		"learning_reason":        body.LearningReason,
		"custom_learning_reason": body.CustomLearningReason,
	}

	prefsJSON, err := json.Marshal(preferences)
	if err != nil {
		logger.Errorw("Failed to marshal preferences", "userId", userID, "error", err)
		sentry.CaptureException(err)
		return nil, errors.Internal("failed to marshal preferences")
	}

	profile := &models.UserProfile{
		UserID:       userID,
		DisplayName:  body.DisplayName,
		EnglishLevel: body.EnglishLevel,
		Preferences:  datatypes.JSON(prefsJSON),
	}

	var savedProfile *models.UserProfile

	err = s.uow.Do(ctx, func(ctx context.Context, p transaction.IProvider) error {
		if err := p.UserProfile().Upsert(ctx, profile); err != nil {
			logger.Errorw("Failed to upsert profile in DB", "userId", userID, "error", err)
			sentry.CaptureException(err)
			return err
		}

		metadata := map[string]interface{}{
			"onboarding_completed": true,
			"role":                 enums.UserRoleUser,
		}
		metadataJSON, _ := json.Marshal(metadata)
		rawMetadata := json.RawMessage(metadataJSON)

		_, err := clerkuser.UpdateMetadata(ctx, userID, &clerkuser.UpdateMetadataParams{
			PublicMetadata: &rawMetadata,
		})
		if err != nil {
			logger.Errorw("Failed to update Clerk metadata", "userId", userID, "error", err)
			sentry.CaptureException(err)
			return err
		}

		return nil
	})

	if err != nil {
		logger.Errorw("Failed to upsert user profile", "userId", userID, "error", err)
		sentry.CaptureException(err)
		return nil, errors.Internal("failed to upsert user profile")
	}

	return savedProfile, nil
}
