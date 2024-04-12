package rediscache

import (
	"context"
	"encoding/json"
	"fmt"
	"log/slog"
	"time"

	"github.com/Heatdog/Avito/internal/config"
	"github.com/Heatdog/Avito/pkg/cache"
	"github.com/redis/go-redis/v9"
)

type redisCache[K comparable, V any] struct {
	client *redis.Client
	logger *slog.Logger
	expire time.Duration
}

func (cache redisCache[K, V]) Add(ctx context.Context, key K, value V) (evicated bool, err error) {
	keyStr, err := json.Marshal(key)
	if err != nil {
		cache.logger.Warn(err.Error())
		return false, err
	}

	valStr, err := json.Marshal(value)
	if err != nil {
		cache.logger.Warn(err.Error())
		return false, err
	}

	cache.logger.Debug("add", slog.String("key", string(keyStr)), slog.String("value", string(valStr)))

	if err := cache.client.Set(ctx, string(keyStr), string(valStr), cache.expire).Err(); err != nil {
		cache.logger.Warn(err.Error())
		return false, err
	}

	return false, nil
}

func (cache redisCache[K, V]) Get(ctx context.Context, key K) (value V, ok bool, err error) {
	strKey, err := json.Marshal(key)
	if err != nil {
		cache.logger.Warn(err.Error())
		return value, false, err
	}

	cache.logger.Debug("get", slog.String("key", string(strKey)))

	val, err := cache.client.Get(ctx, string(strKey)).Result()
	if err == redis.Nil {
		return value, false, nil
	}

	if err != nil {
		return value, false, err
	}

	if err := json.Unmarshal([]byte(val), &value); err != nil {
		cache.logger.Warn(err.Error())
		return value, false, err
	}

	cache.logger.Debug("get result", slog.Any("value", value), slog.Any("ok", ok))

	return value, true, nil
}

func (cache redisCache[K, V]) Remove(ctx context.Context, key K) (bool, error) {
	strKey, err := json.Marshal(key)
	if err != nil {
		cache.logger.Warn(err.Error())
		return false, err
	}

	cache.logger.Debug("delete", slog.String("key", string(strKey)))

	num, err := cache.client.Del(ctx, string(strKey)).Result()
	if err != nil {
		cache.logger.Warn(err.Error())
		return false, err
	}

	return num > 0, nil
}

func NewRedisClient[K comparable, V any](ctx context.Context, redisCfg *config.RedisSettings,
	cacheCfg *config.CacheSettings, logger *slog.Logger) (cache.Cache[K, V], error) {
	time.Sleep(time.Duration(redisCfg.TimePrepare) * time.Second)
	host := fmt.Sprintf("%s:%d", redisCfg.Host, redisCfg.Port)
	client := redis.NewClient(&redis.Options{
		Addr:     host,
		Password: redisCfg.Password,
		DB:       0,
	})

	if _, err := client.Ping(ctx).Result(); err != nil {
		return nil, err
	}

	return redisCache[K, V]{
		client: client,
		logger: logger,
		expire: time.Minute * time.Duration(cacheCfg.TTL),
	}, nil
}
