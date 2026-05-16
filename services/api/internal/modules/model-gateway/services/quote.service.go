package services

import (
	"context"
	"time"

	redisClient "github.com/gianghp123/SonaVoice/api/internal/redis-client"
	"github.com/gianghp123/SonaVoice/api/internal/redis-client/scripts"
)

type IQuoteService interface {
	Reserve(ctx context.Context, userID string, maxDuration, dailyQuota int64) (bool, error)

	// Release returns the unused portion of a previous reservation.
	Release(ctx context.Context, userID string, reservedAmount, actualUsage, dailyQuota int64) error

	// AcquireSessionLock tries to grab a short-lived lock for session creation.
	// Returns true if acquired, false if another process holds it.
	AcquireSessionLock(ctx context.Context, userID string, ttl time.Duration) (bool, error)

	// ReleaseSessionLock releases the session creation lock early.
	ReleaseSessionLock(ctx context.Context, userID string) error
}

type quoteService struct {
	redisClient   redisClient.IRedisClient
	reverseScript *scripts.Script
	releaseScript *scripts.Script
}

func NewQuoteService(client redisClient.IRedisClient) IQuoteService {
	ctx := context.Background()
	reverseScript, err := scripts.New("reverse.lua")

	if err != nil {
		panic(err)
	}

	if err := client.LoadScript(ctx, reverseScript); err != nil {
		panic(err)
	}

	releaseScript, err := scripts.New("release.lua")

	if err != nil {
		panic(err)
	}

	if err := client.LoadScript(ctx, releaseScript); err != nil {
		panic(err)
	}

	return &quoteService{
		redisClient:   client,
		reverseScript: reverseScript,
		releaseScript: releaseScript,
	}
}

func (s *quoteService) Reserve(
	ctx context.Context,
	userID string,
	maxDuration,
	dailyQuota int64,
) (bool, error) {
	return false, nil
}

func (s *quoteService) Release(
	ctx context.Context,
	userID string,
	reservedAmount,
	actualUsage,
	dailyQuota int64,
) error {
	return nil
}

func (s *quoteService) AcquireSessionLock(
	ctx context.Context,
	userID string,
	ttl time.Duration,
) (bool, error) {
	return false, nil
}

func (s *quoteService) ReleaseSessionLock(
	ctx context.Context,
	userID string,
) error {
	return nil
}
