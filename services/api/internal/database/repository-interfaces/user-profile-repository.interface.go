package repository_interfaces

import (
	"context"

	"github.com/gianghp123/SonaVoice/api/internal/database/models"
)

type IUserProfileRepository interface {
	GetByUserID(ctx context.Context, userID string) (*models.UserProfile, error)
	Upsert(ctx context.Context, profile *models.UserProfile) error
}
