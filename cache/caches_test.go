package cache

import (
	"testing"

	"time"
)

type keyTestCase[K comparable] struct {
	key      K
	expected *cacheTimeMap[K]
}

func checkCacheKey[K comparable](ix int, key K, k, e cacheKey[K], t *testing.T) {
	if k.prev != e.prev {
		t.Errorf("Case #%d, key %v, want prev %v, saw prev %v", ix, key, e.prev, k.prev)
	}
	if k.next != e.next {
		t.Errorf("Case #%d, key %v, want next %v, saw next %v", ix, key, e.next, k.next)
	}
}

func checkTimeMap[K comparable](ix int, want, saw *cacheTimeMap[K], t *testing.T) {

	if len(want.m) != len(saw.m) {
		t.Errorf("Case #%d, want map with %d keys, saw map with %d keys", ix, len(want.m), len(saw.m))
	}

	if saw.first != want.first {
		t.Errorf("Case #%d, want first %v, saw first %v", ix, want.first, saw.first)
	}
	if saw.last != want.last {
		t.Errorf("Case #%d, want last %v, saw last %v", ix, want.last, saw.last)
	}

	for k, v := range want.m {
		vS, ok := saw.m[k]
		if ok {
			checkCacheKey(ix, k, v, vS, t)
		} else {
			t.Errorf("Case #%d, key %v is not in the seen map", ix, k)
		}
	}
}

func TestCacheKeyedWithBytes(t *testing.T) {
	cases := []keyTestCase[byte]{
		{ // 0
			key: 30,
			expected: &cacheTimeMap[byte]{
				m: map[byte]cacheKey[byte]{
					30: cacheKey[byte]{prev: 30, next: 30},
				},
				first: 30,
				last:  30,
			},
		},
		{ // 1
			key: 30,
			expected: &cacheTimeMap[byte]{
				m: map[byte]cacheKey[byte]{
					30: cacheKey[byte]{prev: 30, next: 30},
				},
				first: 30,
				last:  30,
			},
		},
		{ // 2
			key: 20,
			expected: &cacheTimeMap[byte]{
				m: map[byte]cacheKey[byte]{
					30: cacheKey[byte]{prev: 20, next: 30},
					20: cacheKey[byte]{prev: 20, next: 30},
				},
				first: 20,
				last:  30,
			},
		},
		{ // 3
			key: 20,
			expected: &cacheTimeMap[byte]{
				m: map[byte]cacheKey[byte]{
					30: cacheKey[byte]{prev: 20, next: 30},
					20: cacheKey[byte]{prev: 20, next: 30},
				},
				first: 20,
				last:  30,
			},
		},
		{ // 4
			key: 30,
			expected: &cacheTimeMap[byte]{
				m: map[byte]cacheKey[byte]{
					30: cacheKey[byte]{prev: 30, next: 20},
					20: cacheKey[byte]{prev: 30, next: 20},
				},
				first: 30,
				last:  20,
			},
		},
		{ // 5
			key: 10,
			expected: &cacheTimeMap[byte]{
				m: map[byte]cacheKey[byte]{
					30: cacheKey[byte]{prev: 10, next: 20},
					20: cacheKey[byte]{prev: 30, next: 20},
					10: cacheKey[byte]{prev: 10, next: 30},
				},
				first: 10,
				last:  20,
			},
		},
		{ // 6
			key: 10,
			expected: &cacheTimeMap[byte]{
				m: map[byte]cacheKey[byte]{
					30: cacheKey[byte]{prev: 10, next: 20},
					20: cacheKey[byte]{prev: 30, next: 20},
					10: cacheKey[byte]{prev: 10, next: 30},
				},
				first: 10,
				last:  20,
			},
		},
		{ // 7
			key: 20,
			expected: &cacheTimeMap[byte]{
				m: map[byte]cacheKey[byte]{
					30: cacheKey[byte]{prev: 10, next: 30},
					20: cacheKey[byte]{prev: 20, next: 10},
					10: cacheKey[byte]{prev: 20, next: 30},
				},
				first: 20,
				last:  30,
			},
		},
		{ // 8
			key: 10,
			expected: &cacheTimeMap[byte]{
				m: map[byte]cacheKey[byte]{
					30: cacheKey[byte]{prev: 20, next: 30},
					20: cacheKey[byte]{prev: 10, next: 30},
					10: cacheKey[byte]{prev: 10, next: 20},
				},
				first: 10,
				last:  30,
			},
		},
	}

	underTest := newCacheTimeMap(byte(0))

	for ix, tc := range cases {
		now := time.Unix(int64(ix), 0)
		updateTimeMap(underTest, tc.key, now)
		checkTimeMap(ix, tc.expected, underTest, t)
	}

}

func TestDeletOldestKeyByte(t *testing.T) {
	underTest := &cacheTimeMap[byte]{
		m: map[byte]cacheKey[byte]{
			30: cacheKey[byte]{prev: 20, next: 30},
			20: cacheKey[byte]{prev: 10, next: 30},
			10: cacheKey[byte]{prev: 10, next: 20},
		},
		first: 10,
		last:  30,
	}

	cases := []keyTestCase[byte]{
		{ // 0
			key: 30,
			expected: &cacheTimeMap[byte]{
				m: map[byte]cacheKey[byte]{
					20: cacheKey[byte]{prev: 10, next: 20},
					10: cacheKey[byte]{prev: 10, next: 20},
				},
				first: 10,
				last:  20,
			},
		},
		{ // 1
			key: 20,
			expected: &cacheTimeMap[byte]{
				m: map[byte]cacheKey[byte]{
					10: cacheKey[byte]{prev: 10, next: 10},
				},
				first: 10,
				last:  10,
			},
		},
		{ // 2
			key: 10,
			expected: &cacheTimeMap[byte]{
				m:     map[byte]cacheKey[byte]{},
				first: 0,
				last:  0,
			},
		},
		{ // 3
			key: 0,
			expected: &cacheTimeMap[byte]{
				m:     map[byte]cacheKey[byte]{},
				first: 0,
				last:  0,
			},
		},
	}

	for ix, tc := range cases {
		removed := removeOldest(underTest)

		if removed != tc.key {
			t.Errorf("Case #%d, expected key %d, got %d", ix, tc.key, removed)
		}
		checkTimeMap(ix, tc.expected, underTest, t)
	}
}
