package repositories

import (
	"context"
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

func (r *userQuotaRepository) Reserve(ctx context.Context, userID string, quotaKey string, quotaDate time.Time, dailyLimit int64) (int64, error) {
	var reserved int64
	err := r.db.WithContext(ctx).Raw(`
		WITH upsert AS (
			INSERT INTO user_quotas (user_id, quota_key, quota_date, remaining, created_at, updated_at)
			VALUES (?, ?, ?, ?, NOW(), NOW())
			ON CONFLICT (user_id, quota_key, quota_date)
			DO UPDATE SET remaining = user_quotas.remaining, updated_at = NOW()
			RETURNING remaining
		)
		SELECT remaining FROM upsert
	`, userID, quotaKey, quotaDate, dailyLimit).Scan(&reserved).Error
	if err != nil {
		return 0, err
	}
	if reserved <= 0 {
		return 0, nil
	}
	if reserved > dailyLimit {
		reserved = dailyLimit
	}
	err = r.db.WithContext(ctx).
		Model(&models.UserQuota{}).
		Where("user_id = ? AND quota_key = ? AND quota_date = ?", userID, quotaKey, quotaDate).
		Update("remaining", gorm.Expr("remaining - ?", reserved)).Error
	if err != nil {
		return 0, err
	}
	return reserved, nil
}

func (r *userQuotaRepository) Release(ctx context.Context, userID string, quotaKey string, quotaDate time.Time, amount int64) error {
	if amount <= 0 {
		return nil
	}
	return r.db.WithContext(ctx).
		Model(&models.UserQuota{}).
		Where("user_id = ? AND quota_key = ? AND quota_date = ?", userID, quotaKey, quotaDate).
		Update("remaining", gorm.Expr("remaining + ?", amount)).Error
}
