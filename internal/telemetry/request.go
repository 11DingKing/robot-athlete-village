package telemetry

import (
	"context"
	"crypto/rand"
	"encoding/hex"
)

type key struct{}

func WithRequestID(ctx context.Context, id string) context.Context {
	return context.WithValue(ctx, key{}, id)
}
func RequestID(ctx context.Context) string {
	if id, ok := ctx.Value(key{}).(string); ok && id != "" {
		return id
	}
	b := make([]byte, 8)
	_, _ = rand.Read(b)
	return hex.EncodeToString(b)
}
