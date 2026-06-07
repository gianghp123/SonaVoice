package services

import (
	"context"
	"encoding/json"

	clerkuser "github.com/clerk/clerk-sdk-go/v2/user"
	"github.com/getsentry/sentry-go"
	clerkclient "github.com/gianghp123/SonaVoice/api/internal/clients/clerk-client"
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
	Onboard(ctx context.Context, userID string, body *req.UpsertProfileReq) *errors.AppError
	Update(ctx context.Context, userID string, body *req.UpdateProfileReq) *errors.AppError
}

type userProfileService struct {
	profileRepo repository_interfaces.IUserProfileRepository
	uow         transaction.UnitOfWork
	clerkClient clerkclient.IClerkClient
}

func NewUserProfileService(
	profileRepo repository_interfaces.IUserProfileRepository,
	uow transaction.UnitOfWork,
	clerkClient clerkclient.IClerkClient,
) IUserProfileService {
	return &userProfileService{
		profileRepo: profileRepo,
		uow:         uow,
		clerkClient: clerkClient,
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

func (s *userProfileService) Onboard(ctx context.Context, userID string, body *req.UpsertProfileReq) *errors.AppError {
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
		return errors.Internal("failed to marshal preferences")
	}

	profile := &models.UserProfile{
		UserID:       userID,
		DisplayName:  body.DisplayName,
		EnglishLevel: body.EnglishLevel,
		Preferences:  datatypes.JSON(prefsJSON),
	}

	err = s.uow.Do(ctx, func(ctx context.Context, p transaction.IProvider) error {
		if err := p.UserProfile().Upsert(ctx, profile); err != nil {
			logger.Errorw("Failed to upsert profile in DB", "userId", userID, "error", err)
			sentry.CaptureException(err)
			return err
		}

		metadata := map[string]interface{}{
			"onboardingCompleted": true,
			"role":                enums.UserRoleUser,
		}
		metadataJSON, _ := json.Marshal(metadata)
		rawMetadata := json.RawMessage(metadataJSON)

		_, err := s.clerkClient.UpdateMetadata(ctx, userID, &clerkuser.UpdateMetadataParams{
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
		return errors.Internal("failed to upsert user profile")
	}

	return nil
}

func (s *userProfileService) Update(ctx context.Context, userID string, body *req.UpdateProfileReq) *errors.AppError {
	logger := zapLogger.S()

	existing, err := s.profileRepo.GetByUserID(ctx, userID)
	if err != nil {
		logger.Errorw("Failed to get user profile for update", "userId", userID, "error", err)
		sentry.CaptureException(err)
		return errors.MapRepoError(err)
	}

	if body.DisplayName != nil {
		existing.DisplayName = *body.DisplayName
	}
	if body.EnglishLevel != nil {
		existing.EnglishLevel = *body.EnglishLevel
	}

	var prefs map[string]interface{}
	if err := json.Unmarshal(existing.Preferences, &prefs); err != nil {
		prefs = map[string]interface{}{}
	}

	if body.NativeLanguage != nil {
		prefs["native_language"] = *body.NativeLanguage
	}
	if body.ImprovementGoals != nil {
		prefs["improvement_goals"] = *body.ImprovementGoals
	}
	if body.Topics != nil {
		prefs["topics"] = *body.Topics
	}
	if body.CustomTopics != nil {
		prefs["custom_topics"] = *body.CustomTopics
	}
	if body.LearningReason != nil {
		prefs["learning_reason"] = *body.LearningReason
	}
	if body.CustomLearningReason != nil {
		prefs["custom_learning_reason"] = *body.CustomLearningReason
	}

	prefsJSON, err := json.Marshal(prefs)
	if err != nil {
		logger.Errorw("Failed to marshal preferences", "userId", userID, "error", err)
		sentry.CaptureException(err)
		return errors.Internal("failed to marshal preferences")
	}

	existing.Preferences = datatypes.JSON(prefsJSON)

	if err := s.profileRepo.Upsert(ctx, existing); err != nil {
		logger.Errorw("Failed to update user profile", "userId", userID, "error", err)
		sentry.CaptureException(err)
		return errors.Internal("failed to update user profile")
	}

	return nil
}
