package repository_interfaces

import (
	"context"
	"time"
)

type IUserQuotaRepository interface {
	ReserveAll(ctx context.Context, userID string, quotaKey string, quotaDate time.Time, dailyLimit int64) (int64, error)
	Release(ctx context.Context, userID string, quotaKey string, quotaDate time.Time, amount int64) error
}
