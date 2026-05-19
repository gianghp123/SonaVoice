package repository_interfaces

import (
	"context"
	"time"
)

type IUserQuotaRepository interface {
	GetOrCreate(ctx context.Context, userID string, quotaKey string, quotaDate time.Time, initialAmount int64) (int64, error)
	Deduct(ctx context.Context, userID string, quotaKey string, quotaDate time.Time, amount int64) error
	GetRemaining(ctx context.Context, userID string, quotaKey string, quotaDate time.Time) (int64, error)
}
