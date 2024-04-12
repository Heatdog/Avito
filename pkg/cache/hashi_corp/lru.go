package hashicorplru

import (
	"context"
	"log/slog"

	cash "github.com/Heatdog/Avito/pkg/cache"
	"github.com/hashicorp/golang-lru/v2/expirable"
)

type LRU[K comparable, V any] struct {
	logger *slog.Logger
	cache  *expirable.LRU[K, V]
}

func NewLRU[K comparable, V any](logger *slog.Logger, cache *expirable.LRU[K, V]) cash.Cache[K, V] {
	return &LRU[K, V]{
		logger: logger,
		cache:  cache,
	}
}

func (lru LRU[K, V]) Get(_ context.Context, key K) (V, bool, error) {
	lru.logger.Debug("get", slog.Any("key", key))
	val, ok := lru.cache.Get(key)
	lru.logger.Debug("get result", slog.Any("value", val), slog.Any("ok", ok))

	return val, ok, nil
}

func (lru LRU[K, V]) Add(_ context.Context, key K, value V) (bool, error) {
	lru.logger.Debug("add", slog.Any("key", key), slog.Any("value", value))
	evicated := lru.cache.Add(key, value)

	return evicated, nil
}

func (lru LRU[K, V]) Remove(_ context.Context, key K) (bool, error) {
	lru.logger.Debug("delete", slog.Any("key", key))
	return lru.cache.Remove(key), nil
}
