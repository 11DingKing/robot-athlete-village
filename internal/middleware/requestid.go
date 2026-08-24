package middleware

import (
	"github.com/11DingKing/robot-athlete-village/internal/telemetry"
	"net/http"
)

func RequestID(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		id := r.Header.Get("X-Request-ID")
		ctx := telemetry.WithRequestID(r.Context(), id)
		w.Header().Set("X-Request-ID", telemetry.RequestID(ctx))
		next.ServeHTTP(w, r.WithContext(ctx))
	})
}
