package worker

import (
	"testing"
	"time"
)

func TestBackoffSequence(t *testing.T) {
	b := Backoff{Base: time.Second, Max: 8 * time.Second}
	cases := []time.Duration{time.Second, time.Second, 2 * time.Second, 4 * time.Second, 8 * time.Second, 8 * time.Second}
	for i, want := range cases {
		if got := b.Delay(i); got != want {
			t.Fatalf("attempt %d got %v want %v", i, got, want)
		}
	}
}
func TestBackoffMinimum(t *testing.T) {
	if got := (Backoff{Base: time.Second, Max: time.Minute}).Delay(-1); got != time.Second {
		t.Fatal(got)
	}
}
func TestRetryable(t *testing.T) {
	for _, s := range []string{"queued", "retry"} {
		if !Retryable(s) {
			t.Fatal(s)
		}
	}
	for _, s := range []string{"running", "done", "failed"} {
		if Retryable(s) {
			t.Fatal(s)
		}
	}
}
func TestTerminal(t *testing.T) {
	for _, s := range []string{"done", "failed"} {
		if !Terminal(s) {
			t.Fatal(s)
		}
	}
	if Terminal("retry") {
		t.Fatal("retry terminal")
	}
}
func TestNextAttempt(t *testing.T) {
	now := time.Unix(100, 0)
	got := NextAttempt(now, 3, Backoff{Base: time.Second, Max: 10 * time.Second})
	if got.Sub(now) != 4*time.Second {
		t.Fatal(got)
	}
}
