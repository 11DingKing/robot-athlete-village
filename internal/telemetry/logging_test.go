package telemetry

import (
	"errors"
	"testing"
)

func TestErrorAttrs(t *testing.T) {
	if ErrorAttrs(nil) != nil {
		t.Fatal("nil attrs")
	}
	attrs := ErrorAttrs(errors.New("boom"))
	if len(attrs) != 2 || attrs[1] != "boom" {
		t.Fatalf("%v", attrs)
	}
}
func TestLoggerExists(t *testing.T) {
	if Logger() == nil {
		t.Fatal("nil logger")
	}
}
