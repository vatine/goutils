package maputils

import (
	"testing"
)

func TestReceive(t *testing.T) {
	c := Channel(map[int]string{})

	go func() {
		c <- Transit[int, string]{Key: 1, Value: "one"}
		c <- Transit[int, string]{Key: 2, Value: "two"}
		c <- Transit[int, string]{Key: 3, Value: "three"}
		close(c)
	}()

	seen := Receive(c)

	if len(seen) != 3 {
		t.Errorf("Unexpected lengtyh, saw %d, want 3", len(seen))
	}

	v1, ok1 := seen[1]
	if !ok1 {
		t.Errorf("Key 1 not seen")
	}
	if v1 != "one" {
		t.Errorf("Expected one, saw %s", v1)
	}
	v2, ok2 := seen[2]
	if !ok2 {
		t.Errorf("Key 2 not seen")
	}
	if v2 != "two" {
		t.Errorf("Expected two, saw %s", v2)
	}
	v3, ok3 := seen[3]
	if !ok3 {
		t.Errorf("Key 3 not seen")
	}
	if v3 != "three" {
		t.Errorf("Expected three, saw %s", v3)
	}

}

func TestMapEqual(t *testing.T) {
	a := map[int]string{
		1: "one",
		2: "two",
	}
	b := map[int]string{
		1: "one",
		2: "two",
		3: "three",
	}
	c := map[int]string{
		1: "one",
		2: "two",
	}

	if !MapEqual(a, a, t) {
		t.Errorf("Map tested unequal with itself.")
	}

	if MapEqual(a, b, nil) {
		t.Errorf("Map tested equal with a map of a different shape")
	}

	if !MapEqual(a, c, t) {
		t.Errorf("Map tested unequal with a different map that has the same contents and keys.")
	}
}

func TestSend(t *testing.T) {
	cases := []map[int]string{
		map[int]string{},
		map[int]string{
			1: "one",
		},
		map[int]string{
			1: "one",
			2: "one",
			3: "one",
			4: "one",
			5: "one",
		},
	}

	for ix, want := range cases {
		ch := Channel(want)
		go Send(want, ch)
		got := Receive(ch)
		if !MapEqual(got, want, t) {
			t.Errorf("Mismatch in test case %d", ix)
		}
	}
}
