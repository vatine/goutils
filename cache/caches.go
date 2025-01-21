package cache

// Top-level package for various caches, this contains all "common"
// functionality (mainly the code to deal with ageing out keys).

import (
	"errors"
	"time"
)

var IncorrectlySpecified = errors.New("Incorrectly specified cache")

type cacheKey[K comparable] struct {
	prev, next K
	timestamp  time.Time
}

type cacheTimeMap[K comparable] struct {
	m           map[K]cacheKey[K]
	first, last K
}

func newCacheTimeMap[K comparable](k K) *cacheTimeMap[K] {
	var rv cacheTimeMap[K]

	rv.m = make(map[K]cacheKey[K])

	return &rv
}

// Update (or insert) time for a key
func updateTimeMap[K comparable](ctm *cacheTimeMap[K], k K, t time.Time) {
	if len(ctm.m) == 0 {
		ctm.first = k
		ctm.last = k
		entry := cacheKey[K]{
			prev:      k,
			next:      k,
			timestamp: t,
		}
		ctm.m[k] = entry
		return
	}
	entry, ok := ctm.m[k]

	switch {
	case !ok:
		// Entirely new key
		entry = cacheKey[K]{
			prev:      k,
			next:      ctm.first,
			timestamp: t,
		}
		first := ctm.m[ctm.first]
		first.prev = k
		ctm.m[ctm.first] = first
		ctm.first = k
		ctm.m[k] = entry
	case ctm.first == k:
		entry.timestamp = t
		ctm.m[k] = entry
	case ctm.last == k:
		newLastKey := entry.prev
		entry.timestamp = t
		entry.prev = k
		entry.next = ctm.first
		oldFirst := ctm.m[ctm.first]
		oldFirst.prev = k
		ctm.m[ctm.first] = oldFirst
		ctm.first = k
		ctm.m[k] = entry
		newLast := ctm.m[newLastKey]
		newLast.next = newLastKey
		ctm.last = newLastKey
		ctm.m[newLastKey] = newLast
	default:
		oldPrev := ctm.m[entry.prev]
		oldNext := ctm.m[entry.next]
		oldNext.prev = entry.prev
		oldPrev.next = entry.next
		ctm.m[entry.prev] = oldPrev
		ctm.m[entry.next] = oldNext
		entry.timestamp = t
		entry.prev = k
		entry.next = ctm.first
		oldFirst := ctm.m[ctm.first]
		oldFirst.prev = k
		ctm.m[ctm.first] = oldFirst
		ctm.first = k
		ctm.m[k] = entry
	}

}

// Remove the oldest key and update things that need updated. Return
// the key that was removed.
// If no key was removeable, return teh key type zero value.
func removeOldest[K comparable](ctm *cacheTimeMap[K]) K {
	var zero K

	if len(ctm.m) == 0 {
		ctm.first = zero
		ctm.last = zero
		return zero
	}

	oldest := ctm.last
	oldEntry := ctm.m[oldest]
	newOldest := oldEntry.prev
	newOldEntry := ctm.m[newOldest]
	newOldEntry.next = newOldest
	ctm.m[newOldest] = newOldEntry
	ctm.last = newOldest
	delete(ctm.m, oldest)

	if len(ctm.m) == 0 {
		ctm.first = zero
		ctm.last = zero
	}

	return oldest
}

func sinceOldest[K comparable](ctm *cacheTimeMap[K], now time.Time) time.Duration {
	if len(ctm.m) == 0 {
		return 0 * time.Second
	}

	return now.Sub(ctm.m[ctm.last].timestamp)
}
