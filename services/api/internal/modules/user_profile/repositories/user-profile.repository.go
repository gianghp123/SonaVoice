package repositories

import (
	"context"

	"github.com/gianghp123/SonaVoice/api/internal/database/models"
	repository_interfaces "github.com/gianghp123/SonaVoice/api/internal/database/repository-interfaces"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

var _ repository_interfaces.IUserProfileRepository = (*userProfileRepository)(nil)

type userProfileRepository struct {
	db *gorm.DB
}

func NewUserProfileRepository(db *gorm.DB) repository_interfaces.IUserProfileRepository {
	return &userProfileRepository{db: db}
}

func (r *userProfileRepository) GetByUserID(ctx context.Context, userID string) (*models.UserProfile, error) {
	var profile models.UserProfile
	err := r.db.WithContext(ctx).
		Where("user_id = ?", userID).
		First(&profile).Error
	if err != nil {
		return nil, err
	}
	return &profile, nil
}

func (r *userProfileRepository) Upsert(ctx context.Context, profile *models.UserProfile) error {
	return r.db.WithContext(ctx).
		Clauses(clause.OnConflict{
			Columns:   []clause.Column{{Name: "user_id"}},
			DoUpdates: clause.AssignmentColumns([]string{"display_name", "english_level", "preferences", "updated_at"}),
		}).
		Create(profile).Error
}
