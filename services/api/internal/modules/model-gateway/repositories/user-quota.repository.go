package repositories

import (
	"context"
	"time"

	"github.com/gianghp123/SonaVoice/api/internal/database/models"
	repository_interfaces "github.com/gianghp123/SonaVoice/api/internal/database/repository-interfaces"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

type userQuotaRepository struct {
	db *gorm.DB
}

func NewUserQuotaRepository(db *gorm.DB) repository_interfaces.IUserQuotaRepository {
	return &userQuotaRepository{db: db}
}

func (r *userQuotaRepository) ReserveAll(ctx context.Context, userID string, quotaKey string, quotaDate time.Time, dailyLimit int64) (int64, error) {
	quota := &models.UserQuota{
		UserID:    userID,
		QuotaKey:  quotaKey,
		QuotaDate: quotaDate,
		Remaining: dailyLimit,
	}
	if err := r.db.WithContext(ctx).Clauses(clause.OnConflict{
		Columns:   []clause.Column{{Name: "user_id"}, {Name: "quota_key"}, {Name: "quota_date"}},
		DoUpdates: clause.AssignmentColumns([]string{"updated_at"}),
	}).Create(quota).Error; err != nil {
		return 0, err
	}

	var current int64
	if err := r.db.WithContext(ctx).
		Model(&models.UserQuota{}).
		Where("user_id = ? AND quota_key = ? AND quota_date = ?", userID, quotaKey, quotaDate).
		Select("remaining").
		Scan(&current).Error; err != nil {
		return 0, err
	}
	if current <= 0 {
		return 0, nil
	}

	reserve := current
	if reserve > dailyLimit {
		reserve = dailyLimit
	}

	if err := r.db.WithContext(ctx).
		Model(&models.UserQuota{}).
		Where("user_id = ? AND quota_key = ? AND quota_date = ?", userID, quotaKey, quotaDate).
		Update("remaining", gorm.Expr("remaining - ?", reserve)).Error; err != nil {
		return 0, err
	}
	return reserve, nil
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
