package mocks

import (
	"context"
	"time"

	"github.com/gianghp123/SonaVoice/api/internal/clients/redis-client/scripts"
	"github.com/stretchr/testify/mock"
)

type RedisClient struct {
	mock.Mock
}

func (m *RedisClient) Get(ctx context.Context, key string) (string, error) {
	args := m.Called(ctx, key)
	return args.String(0), args.Error(1)
}

func (m *RedisClient) Set(ctx context.Context, key string, value interface{}, ttl time.Duration) error {
	args := m.Called(ctx, key, value, ttl)
	return args.Error(0)
}

func (m *RedisClient) Del(ctx context.Context, keys ...string) error {
	args := m.Called(ctx, keys)
	return args.Error(0)
}

func (m *RedisClient) Exists(ctx context.Context, keys ...string) (int64, error) {
	args := m.Called(ctx, keys)
	return args.Get(0).(int64), args.Error(1)
}

func (m *RedisClient) SetNX(ctx context.Context, key string, value interface{}, ttl time.Duration) (bool, error) {
	args := m.Called(ctx, key, value, ttl)
	return args.Bool(0), args.Error(1)
}

func (m *RedisClient) IncrBy(ctx context.Context, key string, value int64) (int64, error) {
	args := m.Called(ctx, key, value)
	return args.Get(0).(int64), args.Error(1)
}

func (m *RedisClient) DecrBy(ctx context.Context, key string, value int64) (int64, error) {
	args := m.Called(ctx, key, value)
	return args.Get(0).(int64), args.Error(1)
}

func (m *RedisClient) RunScript(ctx context.Context, script *scripts.Script, keys []string, args ...interface{}) (interface{}, error) {
	callArgs := m.Called(ctx, script, keys, args)
	return callArgs.Get(0), callArgs.Error(1)
}

func (m *RedisClient) LoadScript(ctx context.Context, script *scripts.Script) error {
	callArgs := m.Called(ctx, script)
	return callArgs.Error(0)
}
