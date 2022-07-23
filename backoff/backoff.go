package backoff

// A package to provide a well-tested implementation of an exponential
// backoff helper.

import (
	"errors"
	"math/rand"
	"time"
)

// The concrete implementation of an exponential backoff helper
type Exponential struct {
	initialDelay time.Duration
	nextDelay    time.Duration
	jitter       time.Duration
	scale        float64
	maxTries     int32
	currentTries int32
}

type BackoffHelper interface {
	Again() bool
}

type BackoffError string

const StopBackoff BackoffError = "backoff intentionally terminated"
const RetriesExhausted BackoffError = "backoff maximum attempts done"

func (e BackoffError) Error() string {
	return string(e)
}

func randomDuration(max time.Duration) time.Duration {
	next := rand.Int63n(int64(max))

	return time.Duration(next)
}

// This creates an exponential backoff helper with somewhat sane default values
func NewExponential() *Exponential {
	helper := Exponential{
		initialDelay: 100 * time.Millisecond,
		jitter:       50 * time.Millisecond,
		scale:        2.0,
		maxTries:     5,
		currentTries: 0,
	}

	helper.nextDelay = helper.initialDelay + randomDuration(helper.jitter)

	return &helper
}

// Compute the next delay, this is essentially the current delay,
// multiplied by the scale factor and some random jitter added.
func (e *Exponential) updateDelay() {
	jitter := randomDuration(e.jitter)
	scaled := float64(e.nextDelay)
	scaled = scaled * e.scale
	e.nextDelay = time.Duration(scaled) + jitter
}

// Try another backoff step. If the maximum number of attempts haev
// been made, this wil lreturn false and not do anthing else. If there
// are still attempts left, this will sleep the requisite time and
// return true.
func (e *Exponential) Again() bool {
	if e.currentTries >= e.maxTries {
		return false
	}
	e.currentTries++
	delta := e.nextDelay
	e.updateDelay()

	time.Sleep(delta)
	return true
}

// Set the maximum number of tries for a helper.
func (e *Exponential) SetRetries(n int32) *Exponential {
	e.maxTries = n

	return e
}

// Set the initial delay for a helper. If there have, so far, been no
// attempts, it will as a side-effect compute a new "next delay" based
// on the initial delay and some random jitter.
func (e *Exponential) SetInitialDelay(dt time.Duration) *Exponential {
	e.initialDelay = dt
	if e.currentTries == 0 {
		e.nextDelay = dt + randomDuration(e.jitter)
	}

	return e
}

// Set the maximum amount of jitter to use. If ethre have been no
// attempts with this helper, also recompute the next delay to be the
// initial delay plus some random jitter.
func (e *Exponential) SetJitter(dt time.Duration) *Exponential {
	e.jitter = dt
	if e.currentTries == 0 {
		e.nextDelay = e.initialDelay + randomDuration(e.jitter)
	}

	return e
}

// Reset a helper to "no tries, next delay will be initial delay plus
// some random jitter".
func (e *Exponential) Reset() *Exponential {
	e.currentTries = 0
	e.nextDelay = e.initialDelay + randomDuration(e.jitter)

	return e
}

// Set the scaling factor. Will only work on a helper not currently
// going through a backoff session (that is, either not used or
// received a Reset after use).
//
// Will do nothing unless the scaling factor is > 1.0
func (e *Exponential) SetScale(s float64) *Exponential {
	if s <= 1.0 {
		return e
	}
	if e.currentTries != 0 {
		return e
	}

	e.scale = s

	return e
}

// Call a function f, with repeated calls using a backoff helper.
//
// The function will be called. If it returns an error, the backup
// helper's Again method will be called. This will continue until f
// returns no error, the backoff helper has run its maximum number of
// tries, or the error returned is the StopBackoff error from this
// package.
//
// Whatever the last return value was from the function will be returned. If the
func CallWithHelper[T any](h BackoffHelper, f func() (T, error)) (T, error) {
	rv, err := f()

	for err != nil {
		if errors.Is(err, StopBackoff) {
			return rv, err
		}
		if !h.Again() {
			return rv, RetriesExhausted
		}
		rv, err = f()
	}

	return rv, nil
}
