package config

import (
	"os"
	"time"
)

type Config struct {
	Addr, DatabaseURL          string
	SessionTTL, WorkerInterval time.Duration
}

func Load() Config {
	return Config{Addr: env("APP_ADDR", ":8080"), DatabaseURL: env("DATABASE_URL", "file:athlete-village.db"), SessionTTL: duration("SESSION_TTL", 12*time.Hour), WorkerInterval: duration("WORKER_INTERVAL", 2*time.Second)}
}
func env(key, fallback string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return fallback
}
func duration(key string, fallback time.Duration) time.Duration {
	if v := os.Getenv(key); v != "" {
		if d, e := time.ParseDuration(v); e == nil {
			return d
		}
	}
	return fallback
}
