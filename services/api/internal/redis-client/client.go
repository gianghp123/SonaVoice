package redisClient

import (
	"context"
	"time"

	"github.com/gianghp123/SonaVoice/api/internal/redis-client/scripts"
	"github.com/redis/go-redis/v9"
)

// RedisClient provides a generic Redis interface used across the whole application.
type IRedisClient interface {
	// Basic key/value
	Get(ctx context.Context, key string) (string, error)
	Set(ctx context.Context, key string, value interface{}, ttl time.Duration) error
	Del(ctx context.Context, keys ...string) error
	Exists(ctx context.Context, keys ...string) (int64, error)

	// Atomic operations
	SetNX(ctx context.Context, key string, value interface{}, ttl time.Duration) (bool, error)
	IncrBy(ctx context.Context, key string, value int64) (int64, error)
	DecrBy(ctx context.Context, key string, value int64) (int64, error)

	// Lua scripts (for atomic multi‑step logic)
	RunScript(ctx context.Context, script *scripts.Script, keys []string, args ...interface{}) (interface{}, error)
	LoadScript(ctx context.Context, script *scripts.Script) error
}

type redisClient struct {
	client *redis.Client
}

func NewClient(client *redis.Client) IRedisClient {
	return &redisClient{
		client: client,
	}
}

func (r *redisClient) Get(ctx context.Context, key string) (string, error) {
	return r.client.Get(ctx, key).Result()
}

func (r *redisClient) Set(ctx context.Context, key string, value interface{}, ttl time.Duration) error {
	return r.client.Set(ctx, key, value, ttl).Err()
}

func (r *redisClient) Del(ctx context.Context, keys ...string) error {
	return r.client.Del(ctx, keys...).Err()
}

func (r *redisClient) Exists(ctx context.Context, keys ...string) (int64, error) {
	return r.client.Exists(ctx, keys...).Result()
}

func (r *redisClient) SetNX(ctx context.Context, key string, value interface{}, ttl time.Duration) (bool, error) {
	return r.client.SetNX(ctx, key, value, ttl).Result()
}

func (r *redisClient) IncrBy(ctx context.Context, key string, value int64) (int64, error) {
	return r.client.IncrBy(ctx, key, value).Result()
}

func (r *redisClient) DecrBy(ctx context.Context, key string, value int64) (int64, error) {
	return r.client.DecrBy(ctx, key, value).Result()
}

func (r *redisClient) RunScript(ctx context.Context, script *scripts.Script, keys []string, args ...interface{}) (interface{}, error) {
	if script.SHA != "" {
		return r.client.EvalSha(ctx, script.SHA, keys, args...).Result()
	}

	return r.client.Eval(ctx, script.Src, keys, args...).Result()
}

func (r *redisClient) LoadScript(ctx context.Context, script *scripts.Script) error {
	sha, err := r.client.ScriptLoad(ctx, script.Src).Result()
	if err != nil {
		return err
	}

	script.SHA = sha
	return nil
}
