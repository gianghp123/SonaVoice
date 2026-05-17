package services

import (
	"context"
	"fmt"
	"time"

	"github.com/gianghp123/SonaVoice/api/internal/core/errors"
	zapLogger "github.com/gianghp123/SonaVoice/api/internal/core/zap-logger"
	redisClient "github.com/gianghp123/SonaVoice/api/internal/redis-client"
	"github.com/gianghp123/SonaVoice/api/internal/redis-client/scripts"
)

type ISessionLockService interface {
	Acquire(ctx context.Context, userID string, ttl time.Duration) (string, *errors.AppError)
	Release(ctx context.Context, userID, lockValue string) *errors.AppError
}

type sessionLockService struct {
	redisClient       redisClient.IRedisClient
	lockReleaseScript *scripts.Script
}

func NewSessionLockService(client redisClient.IRedisClient) ISessionLockService {
	ctx := context.Background()

	lockReleaseScript, err := scripts.New("lock_release.lua")
	if err != nil {
		panic(err)
	}
	if err := client.LoadScript(ctx, lockReleaseScript); err != nil {
		panic(err)
	}

	return &sessionLockService{
		redisClient:       client,
		lockReleaseScript: lockReleaseScript,
	}
}

func (s *sessionLockService) Acquire(ctx context.Context, userID string, ttl time.Duration) (string, *errors.AppError) {
	logger := zapLogger.S()
	logger.Debugw("Acquiring session lock", "userId", userID, "ttl", ttl)

	if userID == "" {
		logger.Errorw("Cannot acquire session lock for empty userId")
		return "", errors.Internal()
	}
	if ttl <= 0 {
		logger.Errorw("Cannot acquire session lock with invalid ttl", "userId", userID, "ttl", ttl)
		return "", errors.Internal()
	}

	lockKey := fmt.Sprintf("session_lock:%s", userID)
	lockValue := fmt.Sprintf("lock:%s:%d", userID, time.Now().UnixNano())

	ok, err := s.redisClient.SetNX(ctx, lockKey, lockValue, ttl)
	if err != nil {
		logger.Errorw("Failed to acquire session lock", "userId", userID, "error", err)
		return "", errors.Internal()
	}
	if !ok {
		logger.Warnw("Session lock already exists", "userId", userID)
		return "", errors.Conflict("session already active")
	}

	return lockValue, nil
}

func (s *sessionLockService) Release(ctx context.Context, userID, lockValue string) *errors.AppError {
	logger := zapLogger.S()
	logger.Debugw("Releasing session lock", "userId", userID, "lockValue", lockValue)

	if userID == "" {
		logger.Errorw("Cannot release session lock for empty userId")
		return errors.Internal()
	}
	if lockValue == "" {
		logger.Errorw("Cannot release session lock with empty lockValue", "userId", userID)
		return errors.Internal()
	}

	lockKey := fmt.Sprintf("session_lock:%s", userID)
	_, err := s.redisClient.RunScript(ctx, s.lockReleaseScript, []string{lockKey}, lockValue)
	if err != nil {
		logger.Errorw("Failed to release session lock", "userId", userID, "error", err)
		return errors.Internal()
	}

	return nil
}
