package maputils

// A small collection of utility functions handy for dealing with maps.

import (
	"testing"
)

type Transit[K comparable, V any] struct {
	Key   K
	Value V
}

// Create a channel suitable for sending the contents of a map across.
func Channel[K comparable, V any](m map[K]V) chan Transit[K, V] {
	rv := make(chan Transit[K, V])

	return rv
}

// Send the contents of a map (paired up in Transit structures, to
// allow reconstruction on the other end). Once the map has been sent
// across, close the passed-in channel to signal the end of
// transmission.
//
// Returns the count of items transmitted.
func Send[K comparable, V any](m map[K]V, c chan Transit[K, V]) int {
	count := 0

	for k, v := range m {
		item := Transit[K, V]{Value: v, Key: k}
		c <- item
		count++
	}

	close(c)

	return count
}

// Construct a map from the items send across a Transit channel,
// then return that map.
func Receive[K comparable, V any](c chan Transit[K, V]) map[K]V {
	m := make(map[K]V)

	for item := range c {
		m[item.Key] = item.Value
	}

	return m
}

// Create a copy of a map. This is the shallowest-possible copy,
// simply enough to ensure that new[key]=newvalue does not modify
// old[key].
func Clone[K comparable, V any](source map[K]V) map[K]V {
	out := make(map[K]V)

	for k, v := range source {
		out[k] = v
	}

	return out
}

// Compare two maps for key/value equality. If you pass in a
// *testing.T, the function will signal an error for any
// mismatches. Either way, it will return true if all key/value pairs
// are present in both maps and match.
func MapEqual[K comparable, V comparable](seen, want map[K]V, t *testing.T) bool {
	both := make(map[K]bool)
	seenOnly := make(map[K]bool)
	wantOnly := make(map[K]bool)

	rv := true

	for seenK, seenV := range seen {
		wantV, wantOK := want[seenK]

		if !wantOK {
			seenOnly[seenK] = true
			rv = false
		} else {
			both[seenK] = true
			rv = rv && wantV == seenV
			if t != nil && seenV != wantV {
				t.Errorf("seen[%v] != want[%v]", seenV, wantV)
			}
		}
	}

	for wantK, _ := range want {
		if !both[wantK] {
			wantOnly[wantK] = true
			rv = false
		}
	}

	if t != nil {
		for k, _ := range seenOnly {
			t.Errorf("key %v present only in seen map", k)
		}
		for k, _ := range wantOnly {
			t.Errorf("key %v present only in want map", k)
		}
	}

	return rv
}
