package quota

import "context"

type QuotaConfig struct {
	Key        string
	DailyLimit int64
}

type IQuotaService interface {
	ReserveAll(ctx context.Context, userID string, cfg QuotaConfig) (int64, error)

	Release(ctx context.Context, userID string, cfg QuotaConfig, reserved, actual int64) error
}
