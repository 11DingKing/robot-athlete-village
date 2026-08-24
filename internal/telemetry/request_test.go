package telemetry

import (
	"context"
	"testing"
)

func TestRequestIDRoundTrip(t *testing.T) {
	ctx := WithRequestID(context.Background(), "abc")
	if RequestID(ctx) != "abc" {
		t.Fatal("id")
	}
}
func TestRequestIDGenerated(t *testing.T) {
	id := RequestID(context.Background())
	if len(id) != 16 {
		t.Fatalf("len %d", len(id))
	}
}
func TestRequestIDOverride(t *testing.T) {
	ctx := WithRequestID(context.Background(), "one")
	ctx = WithRequestID(ctx, "two")
	if RequestID(ctx) != "two" {
		t.Fatal("override")
	}
}
