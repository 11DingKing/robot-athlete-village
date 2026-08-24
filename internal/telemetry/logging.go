package telemetry

import (
	"log/slog"
	"os"
)

func Logger() *slog.Logger {
	return slog.New(slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{Level: slog.LevelInfo}))
}
func ErrorAttrs(err error) []any {
	if err == nil {
		return nil
	}
	return []any{"error", err.Error()}
}
