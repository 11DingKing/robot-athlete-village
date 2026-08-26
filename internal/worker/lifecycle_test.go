package worker

import (
	"context"
	"testing"
	"time"
)

func TestLifecycleTransitions(t *testing.T) {
	l := &Lifecycle{}
	now := time.Now()
	if !l.Start(now) || l.Start(now) {
		t.Fatal("start")
	}
	if !l.Running() {
		t.Fatal("not running")
	}
	if !l.Stop(now.Add(time.Second)) || l.Stop(now) {
		t.Fatal("stop")
	}
	if l.Running() {
		t.Fatal("still running")
	}
	if !l.Start(now.Add(2 * time.Second)) {
		t.Fatal("restart")
	}
}
func TestRunUntilCancel(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	count := 0
	go func() { time.Sleep(8 * time.Millisecond); cancel() }()
	err := RunUntilCancel(ctx, time.Millisecond, func(context.Context) { count++ })
	if err != context.Canceled || count == 0 {
		t.Fatalf("err %v count %d", err, count)
	}
}
func TestRunUntilCancelInvalid(t *testing.T) {
	if err := RunUntilCancel(context.Background(), 0, func(context.Context) {}); err != context.Canceled {
		t.Fatal(err)
	}
}
