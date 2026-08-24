package config

import (
	"testing"
	"time"
)

func TestLoadDefaults(t *testing.T) {
	t.Setenv("APP_ADDR", "")
	t.Setenv("DATABASE_URL", "")
	t.Setenv("SESSION_TTL", "")
	t.Setenv("WORKER_INTERVAL", "")
	c := Load()
	if c.Addr != ":8080" || c.DatabaseURL == "" || c.SessionTTL <= 0 || c.WorkerInterval <= 0 {
		t.Fatalf("invalid defaults %+v", c)
	}
}
func TestLoadEnvironment(t *testing.T) {
	t.Setenv("APP_ADDR", ":9090")
	t.Setenv("DATABASE_URL", "file:test.db")
	t.Setenv("SESSION_TTL", "3h")
	t.Setenv("WORKER_INTERVAL", "250ms")
	c := Load()
	if c.Addr != ":9090" || c.DatabaseURL != "file:test.db" || c.SessionTTL != 3*time.Hour || c.WorkerInterval != 250*time.Millisecond {
		t.Fatalf("got %+v", c)
	}
}
func TestInvalidDurationFallback(t *testing.T) {
	t.Setenv("SESSION_TTL", "oops")
	if Load().SessionTTL != 12*time.Hour {
		t.Fatal("fallback")
	}
}
