package cache

type CacheKey struct {
	TagID     int
	FeatureID int
}

type Cache[K comparable, V any] interface {
	Get(key K) (value V, ok bool)
	Add(key K, value V) (evicated bool)
}
