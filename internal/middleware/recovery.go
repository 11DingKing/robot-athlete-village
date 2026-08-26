package middleware

import (
	"github.com/11DingKing/robot-athlete-village/internal/telemetry"
	"log/slog"
	"net/http"
	"runtime/debug"
)

func Recovery(log *slog.Logger, next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		defer func() {
			if v := recover(); v != nil {
				log.Error("panic recovered", "panic", v, "stack", string(debug.Stack()))
				writeError(w, http.StatusInternalServerError, "internal_error", telemetry.RequestID(r.Context()))
			}
		}()
		next.ServeHTTP(w, r)
	})
}
func writeError(w http.ResponseWriter, status int, code, requestID string) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	_, _ = w.Write([]byte(`{"error":"` + code + `","request_id":"` + requestID + `"}`))
}
