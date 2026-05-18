package repositories

import (
	"context"
	"time"

	zapLogger "github.com/gianghp123/SonaVoice/api/internal/core/zap-logger"
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

func (r *userQuotaRepository) Reserve(
	ctx context.Context,
	userID string,
	quotaKey string,
	quotaDate time.Time,
	dailyLimit int64,
) (int64, error) {
	logger := zapLogger.S()

	logger.Debugw(
		"reserve quota called",
		"userId", userID,
		"quotaKey", quotaKey,
		"quotaDate", quotaDate,
		"dailyLimit", dailyLimit,
	)

	if dailyLimit <= 0 {
		logger.Warnw(
			"daily limit is less than or equal to zero",
			"userId", userID,
			"quotaKey", quotaKey,
			"quotaDate", quotaDate,
			"dailyLimit", dailyLimit,
		)
		return 0, nil
	}

	type upsertResult struct {
		ID          string    `gorm:"column:id"`
		UserID      string    `gorm:"column:user_id"`
		QuotaKey    string    `gorm:"column:quota_key"`
		QuotaDate   time.Time `gorm:"column:quota_date"`
		Remaining   int64     `gorm:"column:remaining"`
		WasInserted bool      `gorm:"column:was_inserted"`
		CreatedAt   time.Time `gorm:"column:created_at"`
		UpdatedAt   time.Time `gorm:"column:updated_at"`
	}

	var upserted upsertResult

	err := r.db.WithContext(ctx).Raw(`
		INSERT INTO user_quotas (
			user_id,
			quota_key,
			quota_date,
			remaining,
			created_at,
			updated_at
		)
		VALUES (?, ?, ?, ?, NOW(), NOW())
		ON CONFLICT (user_id, quota_key, quota_date)
		DO UPDATE SET
			updated_at = user_quotas.updated_at
		RETURNING
			id,
			user_id,
			quota_key,
			quota_date,
			remaining,
			(xmax = 0) AS was_inserted,
			created_at,
			updated_at
	`, userID, quotaKey, quotaDate, dailyLimit).Scan(&upserted).Error

	if err != nil {
		logger.Errorw(
			"failed to upsert quota row",
			"userId", userID,
			"quotaKey", quotaKey,
			"quotaDate", quotaDate,
			"dailyLimit", dailyLimit,
			"error", err,
		)
		return 0, err
	}

	logger.Debugw(
		"quota row after upsert",
		"id", upserted.ID,
		"userId", upserted.UserID,
		"quotaKey", upserted.QuotaKey,
		"quotaDate", upserted.QuotaDate,
		"remaining", upserted.Remaining,
		"wasInserted", upserted.WasInserted,
		"createdAt", upserted.CreatedAt,
		"updatedAt", upserted.UpdatedAt,
	)

	type reserveResult struct {
		Reserved       int64 `gorm:"column:reserved"`
		RemainingAfter int64 `gorm:"column:remaining_after"`
	}

	var result reserveResult

	err = r.db.WithContext(ctx).Raw(`
		WITH reservation AS (
			SELECT
				uq.user_id,
				uq.quota_key,
				uq.quota_date,
				LEAST(uq.remaining, ?) AS amount
			FROM user_quotas uq
			WHERE uq.user_id = ?
			  AND uq.quota_key = ?
			  AND uq.quota_date = ?
			FOR UPDATE
		)
		UPDATE user_quotas uq
		SET
			remaining = uq.remaining - reservation.amount,
			updated_at = NOW()
		FROM reservation
		WHERE uq.user_id = reservation.user_id
		  AND uq.quota_key = reservation.quota_key
		  AND uq.quota_date = reservation.quota_date
		RETURNING
			reservation.amount AS reserved,
			uq.remaining AS remaining_after
	`,
		dailyLimit,
		userID,
		quotaKey,
		quotaDate,
	).Scan(&result).Error

	if err != nil {
		logger.Errorw(
			"failed to reserve quota",
			"userId", userID,
			"quotaKey", quotaKey,
			"quotaDate", quotaDate,
			"dailyLimit", dailyLimit,
			"error", err,
		)
		return 0, err
	}

	logger.Debugw(
		"quota reservation result",
		"userId", userID,
		"quotaKey", quotaKey,
		"quotaDate", quotaDate,
		"dailyLimit", dailyLimit,
		"reserved", result.Reserved,
		"remainingAfter", result.RemainingAfter,
	)

	return result.Reserved, nil
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

func (r *userQuotaRepository) GetRemaining(ctx context.Context, userID string, quotaKey string, quotaDate time.Time) (int64, error) {
	var model models.UserQuota
	err := r.db.WithContext(ctx).
		First(&model, "user_id = ? AND quota_key = ? AND quota_date = ?", userID, quotaKey, quotaDate).Error
	if err != nil {
		return 0, err
	}
	return model.Remaining, nil
}
