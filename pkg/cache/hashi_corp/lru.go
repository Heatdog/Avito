package hashicorp_lru

import (
	"log/slog"

	cash "github.com/Heatdog/Avito/pkg/cache"
	"github.com/hashicorp/golang-lru/v2/expirable"
)

type LRU[K comparable, V any] struct {
	logger *slog.Logger
	cash   *expirable.LRU[K, V]
}

func NewLRU[K comparable, V any](logger *slog.Logger, cash *expirable.LRU[K, V]) cash.Cache[K, V] {
	return &LRU[K, V]{
		logger: logger,
		cash:   cash,
	}
}

func (lru LRU[K, V]) Get(key K) (V, bool) {
	lru.logger.Debug("get", slog.Any("key", key))
	val, ok := lru.cash.Get(key)
	lru.logger.Debug("get result", slog.Any("value", val), slog.Any("ok", ok))
	return val, ok
}

func (lru LRU[K, V]) Add(key K, value V) bool {
	lru.logger.Debug("add", slog.Any("key", key), slog.Any("value", value))
	evicated := lru.cash.Add(key, value)
	return evicated
}
