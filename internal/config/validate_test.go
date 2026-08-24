package config

import (
	"testing"
	"time"
)

func TestConfigValidate(t *testing.T) {
	base := Config{Addr: ":8080", DatabaseURL: "file:test.db", SessionTTL: time.Hour, WorkerInterval: time.Second}
	if e := base.Validate(); e != nil {
		t.Fatal(e)
	}
	cases := []Config{{Addr: "", DatabaseURL: base.DatabaseURL, SessionTTL: base.SessionTTL, WorkerInterval: base.WorkerInterval}, {Addr: base.Addr, DatabaseURL: "", SessionTTL: base.SessionTTL, WorkerInterval: base.WorkerInterval}, {Addr: base.Addr, DatabaseURL: base.DatabaseURL, SessionTTL: time.Second, WorkerInterval: base.WorkerInterval}, {Addr: base.Addr, DatabaseURL: base.DatabaseURL, SessionTTL: base.SessionTTL, WorkerInterval: 0}}
	for _, c := range cases {
		if c.Validate() == nil {
			t.Fatalf("accepted %+v", c)
		}
	}
}
func TestConfigHelpers(t *testing.T) {
	c := Config{Addr: ":8080", DatabaseURL: "file::memory:?mode=memory", SessionTTL: time.Hour, WorkerInterval: time.Second}
	if !c.IsMemoryDatabase() || c.DatabaseScheme() != "file" || c.PublicAddress() != "http://127.0.0.1:8080" {
		t.Fatalf("helpers %+v", c)
	}
	d := Config{}.WithDefaults()
	if d.Addr == "" || d.DatabaseURL == "" || d.SessionTTL == 0 || d.WorkerInterval == 0 {
		t.Fatal(d)
	}
}
func TestConfigWithDefaultsKeepsValues(t *testing.T) {
	c := Config{Addr: ":9000", DatabaseURL: "file:x", SessionTTL: 2 * time.Hour, WorkerInterval: 3 * time.Second}
	d := c.WithDefaults()
	if d.Addr != c.Addr || d.DatabaseURL != c.DatabaseURL || d.SessionTTL != c.SessionTTL || d.WorkerInterval != c.WorkerInterval {
		t.Fatalf("%+v", d)
	}
}
