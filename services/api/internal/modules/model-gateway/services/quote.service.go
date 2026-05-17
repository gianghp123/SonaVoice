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

type IQuoteService interface {
	// ReserveAllRemaining reserves all currently remaining quota for a live session.
	// It returns the amount reserved, which should be used as maxDuration.
	ReserveAllRemaining(ctx context.Context, userID string, dailyQuota int64) (int64, error)

	// Release returns the unused portion of a previous reservation.
	Release(ctx context.Context, userID string, reservedAmount, actualUsage, dailyQuota int64) error

	// AcquireSessionLock returns a lock value that must be passed to ReleaseSessionLock.
	AcquireSessionLock(ctx context.Context, userID string, ttl time.Duration) (string, error)

	// ReleaseSessionLock safely releases the lock only if lockValue matches.
	ReleaseSessionLock(ctx context.Context, userID, lockValue string) error

	ReserveQuota(ctx context.Context, userID string, dailyQuota int64) (reserved int64, cleanup func(bool), err *errors.AppError)
}

type quoteService struct {
	redisClient               redisClient.IRedisClient
	reserveAllRemainingScript *scripts.Script
	releaseScript             *scripts.Script
	lockReleaseScript         *scripts.Script
}

func NewQuoteService(client redisClient.IRedisClient) IQuoteService {
	ctx := context.Background()

	reserveAllRemainingScript, err := scripts.New("reserve.lua")
	if err != nil {
		panic(err)
	}
	if err := client.LoadScript(ctx, reserveAllRemainingScript); err != nil {
		panic(err)
	}

	releaseScript, err := scripts.New("release.lua")
	if err != nil {
		panic(err)
	}
	if err := client.LoadScript(ctx, releaseScript); err != nil {
		panic(err)
	}

	lockReleaseScript, err := scripts.New("lock_release.lua")
	if err != nil {
		panic(err)
	}
	if err := client.LoadScript(ctx, lockReleaseScript); err != nil {
		panic(err)
	}

	return &quoteService{
		redisClient:               client,
		reserveAllRemainingScript: reserveAllRemainingScript,
		releaseScript:             releaseScript,
		lockReleaseScript:         lockReleaseScript,
	}
}

func (s *quoteService) ReserveAllRemaining(
	ctx context.Context,
	userID string,
	dailyQuota int64,
) (int64, error) {
	logger := zapLogger.S()

	logger.Debugw(
		"Reserving all remaining quota",
		"userId", userID,
		"dailyQuota", dailyQuota,
	)

	if userID == "" {
		logger.Errorw("Cannot reserve quota for empty userId")
		return 0, errors.Internal()
	}

	if dailyQuota <= 0 {
		logger.Warnw(
			"Daily quota is not positive",
			"userId", userID,
			"dailyQuota", dailyQuota,
		)
		return 0, nil
	}

	key := quotaKey(userID, time.Now())
	ttl := secondsUntilMidnight()

	result, err := s.redisClient.RunScript(
		ctx,
		s.reserveAllRemainingScript,
		[]string{key},
		dailyQuota,
		ttl,
	)
	if err != nil {
		logger.Errorw(
			"Failed to reserve all remaining quota",
			"userId", userID,
			"error", err,
		)
		return 0, errors.Internal()
	}

	reservedAmount, ok := result.(int64)
	if !ok {
		logger.Errorw(
			"Unexpected reserve script result type",
			"userId", userID,
			"result", result,
		)
		return 0, errors.Internal()
	}

	return reservedAmount, nil
}

func (s *quoteService) Release(
	ctx context.Context,
	userID string,
	reservedAmount,
	actualUsage,
	dailyQuota int64,
) error {
	logger := zapLogger.S()

	logger.Debugw(
		"Releasing quota",
		"userId", userID,
		"reservedAmount", reservedAmount,
		"actualUsage", actualUsage,
		"dailyQuota", dailyQuota,
	)

	if userID == "" {
		logger.Errorw("Cannot release quota for empty userId")
		return errors.Internal()
	}

	if reservedAmount < 0 {
		reservedAmount = 0
	}

	if actualUsage < 0 {
		actualUsage = 0
	}

	if actualUsage > reservedAmount {
		logger.Warnw(
			"Actual usage is greater than reserved amount; clamping",
			"userId", userID,
			"actualUsage", actualUsage,
			"reservedAmount", reservedAmount,
		)
		actualUsage = reservedAmount
	}

	if dailyQuota < 0 {
		dailyQuota = 0
	}

	key := quotaKey(userID, time.Now())
	ttl := secondsUntilMidnight()

	_, err := s.redisClient.RunScript(
		ctx,
		s.releaseScript,
		[]string{key},
		reservedAmount,
		actualUsage,
		dailyQuota,
		ttl,
	)
	if err != nil {
		logger.Errorw(
			"Failed to release quota",
			"userId", userID,
			"reservedAmount", reservedAmount,
			"actualUsage", actualUsage,
			"error", err,
		)
		return errors.Internal()
	}

	return nil
}

func (s *quoteService) AcquireSessionLock(
	ctx context.Context,
	userID string,
	ttl time.Duration,
) (string, error) {
	logger := zapLogger.S()

	logger.Debugw(
		"Acquiring session lock",
		"userId", userID,
		"ttl", ttl,
	)

	if userID == "" {
		logger.Errorw("Cannot acquire session lock for empty userId")
		return "", errors.Internal()
	}

	if ttl <= 0 {
		logger.Errorw(
			"Cannot acquire session lock with invalid ttl",
			"userId", userID,
			"ttl", ttl,
		)
		return "", errors.Internal()
	}

	lockKey := sessionLockKey(userID)
	lockValue := fmt.Sprintf("lock:%s:%d", userID, time.Now().UnixNano())

	ok, err := s.redisClient.SetNX(ctx, lockKey, lockValue, ttl)
	if err != nil {
		logger.Errorw(
			"Failed to acquire session lock",
			"userId", userID,
			"error", err,
		)
		return "", errors.Internal()
	}

	if !ok {
		logger.Warnw(
			"Session lock already exists",
			"userId", userID,
		)
		return "", errors.Conflict("session already active")
	}

	return lockValue, nil
}

func (s *quoteService) ReleaseSessionLock(
	ctx context.Context,
	userID, lockValue string,
) error {
	logger := zapLogger.S()

	logger.Debugw(
		"Releasing session lock",
		"userId", userID,
		"lockValue", lockValue,
	)

	if userID == "" {
		logger.Errorw("Cannot release session lock for empty userId")
		return errors.Internal()
	}

	if lockValue == "" {
		logger.Errorw("Cannot release session lock with empty lockValue", "userId", userID)
		return errors.Internal()
	}

	lockKey := sessionLockKey(userID)

	_, err := s.redisClient.RunScript(
		ctx,
		s.lockReleaseScript,
		[]string{lockKey},
		lockValue,
	)
	if err != nil {
		logger.Errorw(
			"Failed to release session lock",
			"userId", userID,
			"error", err,
		)
		return errors.Internal()
	}

	return nil
}

func (s *quoteService) ReserveQuota(
	ctx context.Context,
	userID string,
	dailyQuota int64,
) (int64, func(bool), *errors.AppError) {
	logger := zapLogger.S()

	reservedAmount, err := s.ReserveAllRemaining(ctx, userID, dailyQuota)
	if err != nil {
		return 0, nil, errors.Internal()
	}

	if reservedAmount <= 0 {
		return 0, nil, errors.Forbidden("quota exceeded")
	}

	var committed bool

	cleanup := func(commit bool) {
		if commit {
			committed = true
			return
		}
		if committed {
			return
		}
		releaseCtx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
		defer cancel()
		if err := s.Release(releaseCtx, userID, reservedAmount, 0, dailyQuota); err != nil {
			logger.Errorw(
				"Failed to rollback reserved quota",
				"userId", userID,
				"reservedAmount", reservedAmount,
				"error", err,
			)
		}
	}

	return reservedAmount, cleanup, nil
}

func quotaKey(userID string, now time.Time) string {
	date := now.Format("2006-01-02")
	return fmt.Sprintf("user_quota:%s:%s", userID, date)
}

func sessionLockKey(userID string) string {
	return fmt.Sprintf("session_lock:%s", userID)
}

func secondsUntilMidnight() int64 {
	now := time.Now()
	midnight := time.Date(
		now.Year(),
		now.Month(),
		now.Day()+1,
		0,
		0,
		0,
		0,
		now.Location(),
	)

	seconds := int64(midnight.Sub(now).Seconds())
	if seconds <= 0 {
		return 1
	}

	return seconds + 1
}
