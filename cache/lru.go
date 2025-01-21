package cache

import (
	"sync"
	"time"
)

// Implements a Least Recently Used cache, bounded by optionally
// number of values and maximum duration since access. When at least
// one of "size of cache" or "maximum age" is reached, items will be
// evicted until both conditions are satisfied. Eviction happens as
// part of reading, or writing, to the cache. For the purposes of the
// LRU cache, both reads and writes are counted as "usage".
type LRU[K comparable, V any] struct {
	lock    sync.Mutex
	m       map[K]V
	keys    *cacheTimeMap[K]
	maxSize int
	maxAge  time.Duration
}

// Return a new Least Recently Used (LRU) cache.
//
// The provided key (k) and value (v) are ONLY used for their type(s).
//
// If a non-positive maxSize is provided, the size of the cache is
// unbounded. If a "zero" time is provided, the "age" is unbounded. If
// both size and age are unbounded, an error is returned.
func NewLRUCache[K comparable, V any](k K, v V, maxSize int, maxAge time.Duration) (*LRU[K, V], error) {
	if (maxSize < 1) && (maxAge == 0) {
		return nil, IncorrectlySpecified
	}
	rv := new(LRU[K, V])
	rv.m = make(map[K]V)
	rv.keys = newCacheTimeMap(k)
	rv.maxAge = maxAge
	rv.maxSize = maxSize

	return rv, nil
}

// Age out oldest entries, until there are (a) bo too-old entries left
// and (b) we are under the max size of the cache.
func lruAge[K comparable, V any](lru *LRU[K, V], now time.Time) {
	if lru.maxAge > 0 {
		var done bool
		for !done {
			since := sinceOldest(lru.keys, now)
			if since < lru.maxAge {
				done = true
				continue
			}

			drop := removeOldest(lru.keys)
			delete(lru.m, drop)

			if len(lru.m) == 0 {
				done = true
			}
		}
	}

	if lru.maxSize > 0 {
		for len(lru.m) > lru.maxSize {
			drop := removeOldest(lru.keys)
			delete(lru.m, drop)
		}
	}
}

// Set cached value for a specific key in an LRU map, uses a
// syncronisation primitive so should be safe for concurrent use.
func SetLRU[K comparable, V any](lru *LRU[K, V], k K, v V) {
	lru.lock.Lock()
	defer lru.lock.Unlock()

	now := time.Now()
	lru.m[k] = v
	updateTimeMap(lru.keys, k, now)
	lruAge(lru, now)
}

// Get cached value for a specific key in an LRU map, uses a
// synchronisation primitive. The returned bool is true if the key
// existed, otherwise false.
func GetLRU[K comparable, V any](lru *LRU[K, V], k K) (V, bool) {
	lru.lock.Lock()
	defer lru.lock.Unlock()

	now := time.Now()
	updateTimeMap(lru.keys, k, now)

	rv, ok := lru.m[k]
	return rv, ok
}
