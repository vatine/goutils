package cache

import (
	"testing"

	"time"
)

func TestAgeLRW(t *testing.T) {
	baseLRW := func() *LRW[int, int] {
		epoch := time.Unix(0, 0)
		ctm := &cacheTimeMap[int]{
			m: map[int]cacheKey[int]{
				10: cacheKey[int]{10, 20, epoch.Add(6 * time.Second)},
				20: cacheKey[int]{10, 30, epoch.Add(5 * time.Second)},
				30: cacheKey[int]{20, 40, epoch.Add(4 * time.Second)},
				40: cacheKey[int]{30, 50, epoch.Add(3 * time.Second)},
				50: cacheKey[int]{40, 60, epoch.Add(2 * time.Second)},
				60: cacheKey[int]{50, 60, epoch.Add(1 * time.Second)},
			},
			first: 10,
			last:  60,
		}
		return &LRW[int, int]{
			m: map[int]int{
				10: 11,
				20: 21,
				30: 31,
				40: 41,
				50: 51,
				60: 61,
			},
			keys:    ctm,
			maxSize: 4,
			maxAge:  5 * time.Second,
		}
	}

	cases := []struct {
		now     time.Time
		expLeft int
	}{
		{time.Unix(6, 0), 4},
		{time.Unix(7, 0), 4},
		{time.Unix(8, 0), 3},
		{time.Unix(9, 0), 2},
		{time.Unix(10, 0), 1},
		{time.Unix(11, 0), 0},
		{time.Unix(12, 0), 0},
	}

	for ix, tc := range cases {
		lrw := baseLRW()
		lrwAge(lrw, tc.now)
		got := len(lrw.m)
		want := tc.expLeft

		if got != want {
			t.Errorf("Case #%d, want %d left, have %d left", ix, want, got)
		}
	}
}

func TestLRWSetAndGet(t *testing.T) {
	lrw, _ := NewLRWCache(0, "", 5, time.Second)

	cases := []struct {
		k       int
		v       string
		expSize int
		set     bool
		ok      bool
	}{
		{10, "ten", 1, true, false},
		{10, "ten", 1, false, true},
		{20, "", 1, false, false},
		{20, "twenty", 2, true, false},
		{20, "twenty", 2, false, true},
		{30, "thirty", 3, true, false},
		{40, "forty", 4, true, false},
		{50, "fifty", 5, true, false},
		{60, "sixty", 5, true, false},
		{10, "", 5, false, false},
		{20, "twenty", 5, false, true},
	}

	for ix, tc := range cases {
		if tc.set {
			SetLRW(lrw, tc.k, tc.v)
			want := len(lrw.m)
			if want != tc.expSize {
				t.Errorf("Case #%d, want size %d, saw %d", ix, tc.expSize, want)
			}
		} else {
			want, ok := GetLRW(lrw, tc.k)
			size := len(lrw.m)
			if want != tc.v {
				t.Errorf("Case #%d, want value «%s», got «%s»", ix, tc.v, want)
			}
			if ok != tc.ok {
				t.Errorf("Case #%d, want ok %v, got %v", ix, tc.ok, ok)
			}
			if size != tc.expSize {
				t.Errorf("Case #%d want size %d, got %d", ix, tc.expSize, size)
			}
		}
	}
}
