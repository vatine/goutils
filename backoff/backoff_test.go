package backoff

import (
	"errors"
	"time"

	"testing"
)

func TestScaleSetting(t *testing.T) {
	e := NewExponential()

	if e.scale <= 1.0 {
		t.Errorf("Initial scale is <= 1.0, %f", e.scale)
	}

	before := e.scale
	e.SetScale(1.0)
	if e.scale != before {
		t.Errorf("Managed to set scale factor to 1")
	}

	before = e.scale
	e.SetScale(-3.5)
	if e.scale != before {
		t.Errorf("Managed to set scale factor to a negative number")
	}

	before = e.scale
	e.SetScale(0.5)
	if e.scale != before {
		t.Errorf("Managed to set scale factor to 0.5")
	}

	before = e.scale
	e.SetScale(0)
	if e.scale != before {
		t.Errorf("Managed to set scale factor to zero")
	}

	e.SetScale(1.75)
	if e.scale != 1.75 {
		t.Errorf("Failed to set scale factor to 1.75, it is now %f", e.scale)
	}
}

func checkInterval(delay, low, high time.Duration, t *testing.T) {
	if delay < low {
		t.Errorf("Delay %v is lower than expected minimum %v", delay, low)
	}
	if delay > high {
		t.Errorf("Delay %v is higher than expected maximum %v", delay, high)
	}
}

func TestSettingInitialAndJitter(t *testing.T) {
	e := NewExponential()

	checkInterval(e.nextDelay, 100*time.Millisecond, 150*time.Millisecond, t)

	e.SetInitialDelay(30 * time.Millisecond)
	checkInterval(e.nextDelay, 30*time.Millisecond, 80*time.Millisecond, t)

	e.SetJitter(100 * time.Millisecond)
	checkInterval(e.nextDelay, 30*time.Millisecond, 130*time.Millisecond, t)

	oldNext := e.nextDelay
	e.currentTries = 1

	e.SetInitialDelay(time.Second)
	if oldNext != e.nextDelay {
		t.Errorf("NextDelay changed when setting initial delay, despite a backoff process in progress.")
	}
	e.SetJitter(time.Second)
	if oldNext != e.nextDelay {
		t.Errorf("NextDelay changed when setting jitter, despite a backoff process in progress.")
	}
}

func TestExtending(t *testing.T) {
	e := NewExponential().SetInitialDelay(100 * time.Millisecond).SetJitter(50 * time.Millisecond).SetScale(2.0)

	for i := 0; i < 10; i++ {
		prevNext := e.nextDelay
		e.updateDelay()
		checkInterval(e.nextDelay, 2*prevNext, 50*time.Millisecond+2*prevNext, t)
	}
}

func TestAgain(t *testing.T) {
	e := NewExponential().SetInitialDelay(10 * time.Millisecond).SetJitter(10 * time.Millisecond).SetScale(2.0)

	before := time.Now()
	e.maxTries = 2
	checkOne := e.Again()
	afterOne := time.Now()
	checkTwo := e.Again()
	afterTwo := time.Now()
	checkThree := e.Again()
	// afterThree := time.Now()

	if !checkOne {
		t.Errorf("Failed the first delay")
	}
	if !checkTwo {
		t.Errorf("Failed the second delay")
	}
	if checkThree {
		t.Errorf("Did not fail the third delay")
	}

	checkInterval(afterOne.Sub(before), 10*time.Millisecond, 20*time.Millisecond, t)
	checkInterval(afterTwo.Sub(afterOne), 20*time.Millisecond, 50*time.Millisecond, t)
}

type tester struct {
	okAfter   int
	stopAfter int
	attempts  int
}

func (t *tester) call() (bool, error) {
	t.attempts++
	if t.attempts >= t.okAfter {
		return true, nil
	}
	if t.attempts >= t.stopAfter {
		return false, StopBackoff
	}

	return false, errors.New("blah")
}

func TestCallWithBackoff(t *testing.T) {
	helper := NewExponential().SetInitialDelay(time.Millisecond).SetJitter(time.Millisecond).SetRetries(5)

	t1 := &tester{
		okAfter:   3,
		stopAfter: 10,
	}
	v, err := CallWithHelper(helper, t1.call)
	if !v {
		t.Errorf("Unexpected return value")
	}
	if t1.attempts != 3 {
		t.Errorf("Unexpected number of calls, saw %d want 3", t1.attempts)
	}
	if err != nil {
		t.Errorf("Unexpected error %v", err)
	}

	helper.Reset()
	t2 := &tester{
		okAfter:   10,
		stopAfter: 10,
	}

	v, err = CallWithHelper(helper, t2.call)
	if v {
		t.Errorf("unexpected return value")
	}
	if t2.attempts != 6 {
		t.Errorf("Unexpected number of calls, saw %d want 6", t2.attempts)
	}
	if err != RetriesExhausted {
		t.Errorf("Unexpected error, %v, expected RetriesExhausted", err)
	}

	helper.Reset()
	t3 := &tester{
		okAfter:   10,
		stopAfter: 2,
	}

	v, err = CallWithHelper(helper, t3.call)
	if v {
		t.Errorf("unexpected return value")
	}
	if t3.attempts != 2 {
		t.Errorf("Unexpected number of calls, saw %d want 2", t3.attempts)
	}
	if err != StopBackoff {
		t.Errorf("Unexpected error, %v, expected StopBackoff", err)
	}

}
