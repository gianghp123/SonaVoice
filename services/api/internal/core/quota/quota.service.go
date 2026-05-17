package quota

import (
	"context"
	"fmt"
	"time"

	zapLogger "github.com/gianghp123/SonaVoice/api/internal/core/zap-logger"
	redisClient "github.com/gianghp123/SonaVoice/api/internal/redis-client"
	"github.com/gianghp123/SonaVoice/api/internal/redis-client/scripts"
)

type quotaService struct {
	redisClient      redisClient.IRedisClient
	reserveAllScript *scripts.Script
	releaseScript    *scripts.Script
}

func NewQuotaService(client redisClient.IRedisClient) IQuotaService {
	ctx := context.Background()

	reserveAllScript, err := scripts.New("reserve.lua")
	if err != nil {
		panic(err)
	}
	if err := client.LoadScript(ctx, reserveAllScript); err != nil {
		panic(err)
	}

	releaseScript, err := scripts.New("release.lua")
	if err != nil {
		panic(err)
	}
	if err := client.LoadScript(ctx, releaseScript); err != nil {
		panic(err)
	}

	return &quotaService{
		redisClient:      client,
		reserveAllScript: reserveAllScript,
		releaseScript:    releaseScript,
	}
}

func (s *quotaService) ReserveAll(ctx context.Context, userID string, cfg QuotaConfig) (int64, error) {
	logger := zapLogger.S()
	logger.Debugw("Reserving all remaining quota", "key", cfg.Key, "userId", userID, "dailyQuota", cfg.DailyLimit)

	if userID == "" {
		logger.Errorw("Cannot reserve quota for empty userId")
		return 0, fmt.Errorf("empty userId")
	}

	if cfg.DailyLimit <= 0 {
		logger.Warnw("Daily quota is not positive", "key", cfg.Key, "userId", userID, "dailyQuota", cfg.DailyLimit)
		return 0, nil
	}

	key := quotaKey(cfg.Key, userID, time.Now())
	ttl := secondsUntilMidnight()

	result, err := s.redisClient.RunScript(ctx, s.reserveAllScript, []string{key}, cfg.DailyLimit, ttl)
	if err != nil {
		logger.Errorw("Failed to reserve all remaining quota", "key", cfg.Key, "userId", userID, "error", err)
		return 0, fmt.Errorf("reserve all quota: %w", err)
	}

	reservedAmount, ok := result.(int64)
	if !ok {
		logger.Errorw("Unexpected reserve script result type", "key", cfg.Key, "userId", userID, "result", result)
		return 0, fmt.Errorf("unexpected reserve result type: %T", result)
	}

	return reservedAmount, nil
}

func (s *quotaService) Release(ctx context.Context, userID string, cfg QuotaConfig, reserved, actual int64) error {
	logger := zapLogger.S()
	logger.Debugw("Releasing quota", "key", cfg.Key, "userId", userID, "reservedAmount", reserved, "actualUsage", actual, "dailyQuota", cfg.DailyLimit)

	if userID == "" {
		logger.Errorw("Cannot release quota for empty userId")
		return fmt.Errorf("empty userId")
	}

	if reserved < 0 {
		reserved = 0
	}
	if actual < 0 {
		actual = 0
	}

	if actual > reserved {
		logger.Warnw("Actual usage is greater than reserved amount; clamping", "key", cfg.Key, "userId", userID, "actualUsage", actual, "reservedAmount", reserved)
		actual = reserved
	}

	if cfg.DailyLimit < 0 {
		cfg.DailyLimit = 0
	}

	key := quotaKey(cfg.Key, userID, time.Now())
	ttl := secondsUntilMidnight()

	_, err := s.redisClient.RunScript(ctx, s.releaseScript, []string{key}, reserved, actual, cfg.DailyLimit, ttl)
	if err != nil {
		logger.Errorw("Failed to release quota", "key", cfg.Key, "userId", userID, "error", err)
		return fmt.Errorf("release quota: %w", err)
	}

	return nil
}

func quotaKey(namespace, userID string, now time.Time) string {
	date := now.Format("2006-01-02")
	return fmt.Sprintf("quota:%s:%s:%s", namespace, userID, date)
}

func secondsUntilMidnight() int64 {
	now := time.Now()
	midnight := time.Date(now.Year(), now.Month(), now.Day()+1, 0, 0, 0, 0, now.Location())
	seconds := int64(midnight.Sub(now).Seconds())
	if seconds <= 0 {
		return 1
	}
	return seconds + 1
}
