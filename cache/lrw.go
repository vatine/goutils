package cache

import (
	"sync"
	"time"
)

// Implements a Least Recently Written cache, bounded by optionally
// number of values and maximum duration since access. When at least
// one of "size of cache" or "maximum age" is reached, items will be
// evicted until both conditions are satisfied. Eviction happens as
// part of reading, or writing, to the cache. For the purposes of the
// LRW cache, both reads and writes are counted as "usage".
type LRW[K comparable, V any] struct {
	lock    sync.Mutex
	m       map[K]V
	keys    *cacheTimeMap[K]
	maxSize int
	maxAge  time.Duration
}

// Return a new Least Recently Written (LRW) cache.
//
// The provided key (k) and value (v) are ONLY used for their type(s).
//
// If a non-positive maxSize is provided, the size of the cache is
// unbounded. If a "zero" time is provided, the "age" is unbounded. If
// both size and age are unbounded, an error is returned.
func NewLRWCache[K comparable, V any](k K, v V, maxSize int, maxAge time.Duration) (*LRW[K, V], error) {
	if (maxSize < 1) && (maxAge == 0) {
		return nil, IncorrectlySpecified
	}
	rv := new(LRW[K, V])
	rv.m = make(map[K]V)
	rv.keys = newCacheTimeMap(k)
	rv.maxAge = maxAge
	rv.maxSize = maxSize

	return rv, nil
}

// Age out oldest entries, until there are (a) no too-old entries left
// and (b) we are under the max size of the cache.
func lrwAge[K comparable, V any](lrw *LRW[K, V], now time.Time) {
	if lrw.maxAge > 0 {
		var done bool
		for !done {
			since := sinceOldest(lrw.keys, now)
			if since < lrw.maxAge {
				done = true
				continue
			}

			drop := removeOldest(lrw.keys)
			delete(lrw.m, drop)

			if len(lrw.m) == 0 {
				done = true
			}
		}
	}

	if lrw.maxSize > 0 {
		for len(lrw.m) > lrw.maxSize {
			drop := removeOldest(lrw.keys)
			delete(lrw.m, drop)
		}
	}
}

// Set cached value for a specific key in an LRW map, uses a
// syncronisation primitive so should be safe for concurrent use.
func SetLRW[K comparable, V any](lrw *LRW[K, V], k K, v V) {
	lrw.lock.Lock()
	defer lrw.lock.Unlock()

	now := time.Now()
	lrw.m[k] = v
	updateTimeMap(lrw.keys, k, now)
	lrwAge(lrw, now)
}

// Get cached value for a specific key in an LRW map, uses a
// synchronisation primitive. The returned bool is true if the key
// existed, otherwise false.
func GetLRW[K comparable, V any](lrw *LRW[K, V], k K) (V, bool) {
	lrw.lock.Lock()
	defer lrw.lock.Unlock()

	rv, ok := lrw.m[k]
	return rv, ok
}
