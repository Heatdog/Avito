package cache

import "context"

type Key struct {
	TagID     int
	FeatureID int
}

type Cache[K comparable, V any] interface {
	Get(ctx context.Context, key K) (value V, ok bool, err error)
	Add(ctx context.Context, key K, value V) (evicated bool, err error)
	Remove(ctx context.Context, key K) (bool, error)
}
