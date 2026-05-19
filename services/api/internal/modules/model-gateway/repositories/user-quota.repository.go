package repositories

import (
	"context"
	"errors"
	"time"

	"github.com/gianghp123/SonaVoice/api/internal/database/models"
	repository_interfaces "github.com/gianghp123/SonaVoice/api/internal/database/repository-interfaces"
	"gorm.io/gorm"
)

type userQuotaRepository struct {
	db *gorm.DB
}

func NewUserQuotaRepository(db *gorm.DB) repository_interfaces.IUserQuotaRepository {
	return &userQuotaRepository{db: db}
}

func (r *userQuotaRepository) GetOrCreate(ctx context.Context, userID string, quotaKey string, quotaDate time.Time, initialAmount int64) (int64, error) {
	var remaining int64
	err := r.db.WithContext(ctx).Raw(`
		INSERT INTO user_quotas (user_id, quota_key, quota_date, remaining, created_at, updated_at)
		VALUES (?, ?, ?, ?, NOW(), NOW())
		ON CONFLICT (user_id, quota_key, quota_date)
		DO UPDATE SET updated_at = user_quotas.updated_at
		RETURNING remaining
	`, userID, quotaKey, quotaDate, initialAmount).Scan(&remaining).Error
	if err != nil {
		return 0, err
	}
	return remaining, nil
}

func (r *userQuotaRepository) Deduct(ctx context.Context, userID string, quotaKey string, quotaDate time.Time, amount int64) error {
	if amount <= 0 {
		return nil
	}
	result := r.db.WithContext(ctx).
		Model(&models.UserQuota{}).
		Where("user_id = ? AND quota_key = ? AND quota_date = ? AND remaining >= ?", userID, quotaKey, quotaDate, amount).
		Update("remaining", gorm.Expr("remaining - ?", amount))
	if result.Error != nil {
		return result.Error
	}
	if result.RowsAffected == 0 {
		return errors.New("insufficient quota")
	}
	return nil
}

func (r *userQuotaRepository) GetRemaining(ctx context.Context, userID string, quotaKey string, quotaDate time.Time) (int64, error) {
	var model models.UserQuota
	err := r.db.WithContext(ctx).
		First(&model, "user_id = ? AND quota_key = ? AND quota_date = ?", userID, quotaKey, quotaDate).Error
	if err != nil {
		return 0, err
	}
	return model.Remaining, nil
}
