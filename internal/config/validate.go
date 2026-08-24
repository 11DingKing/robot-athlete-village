package config

import (
	"fmt"
	"net/url"
	"strings"
	"time"
)

func (c Config) Validate() error {
	if strings.TrimSpace(c.Addr) == "" {
		return fmt.Errorf("address required")
	}
	if c.SessionTTL < time.Minute {
		return fmt.Errorf("session ttl too short")
	}
	if c.WorkerInterval <= 0 {
		return fmt.Errorf("worker interval must be positive")
	}
	if strings.TrimSpace(c.DatabaseURL) == "" {
		return fmt.Errorf("database url required")
	}
	return nil
}
func (c Config) IsMemoryDatabase() bool {
	return strings.Contains(c.DatabaseURL, "mode=memory") || strings.HasPrefix(c.DatabaseURL, "file::memory:")
}
func (c Config) DatabaseScheme() string {
	if u, e := url.Parse(c.DatabaseURL); e == nil && u.Scheme != "" {
		return u.Scheme
	}
	return "sqlite"
}
func (c Config) WithDefaults() Config {
	out := c
	if out.Addr == "" {
		out.Addr = ":8080"
	}
	if out.DatabaseURL == "" {
		out.DatabaseURL = "file:athlete-village.db"
	}
	if out.SessionTTL == 0 {
		out.SessionTTL = 12 * time.Hour
	}
	if out.WorkerInterval == 0 {
		out.WorkerInterval = 2 * time.Second
	}
	return out
}
func (c Config) PublicAddress() string {
	if strings.HasPrefix(c.Addr, ":") {
		return "http://127.0.0.1" + c.Addr
	}
	return "http://" + c.Addr
}
