package clock

import (
	"testing"
	"time"
)

func TestFixedClock(t *testing.T) {
	want := time.Date(2026, 8, 24, 0, 0, 0, 0, time.UTC)
	if got := (Fixed{Value: want}).Now(); !got.Equal(want) {
		t.Fatalf("got %v", got)
	}
}
func TestRealClockUTC(t *testing.T) {
	got := Real{}.Now()
	if got.Location() != time.UTC {
		t.Fatalf("location %v", got.Location())
	}
	if time.Since(got) > time.Second {
		t.Fatalf("clock stale")
	}
}
func TestFixedClockDoesNotChange(t *testing.T) {
	want := time.Unix(1, 2).UTC()
	c := Fixed{Value: want}
	for i := 0; i < 5; i++ {
		if !c.Now().Equal(want) {
			t.Fatal("changed")
		}
	}
}
