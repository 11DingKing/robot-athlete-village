package httpapi

import (
	"encoding/json"
	"net/http"
)

type envelope struct {
	Data      any    `json:"data,omitempty"`
	Error     string `json:"error,omitempty"`
	RequestID string `json:"request_id,omitempty"`
}

func respond(w http.ResponseWriter, status int, data any, requestID string) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(envelope{Data: data, RequestID: requestID})
}
func requestMethodAllowed(w http.ResponseWriter, allowed ...string) bool {
	for _, method := range allowed {
		if method == http.MethodGet {
			return true
		}
	}
	return false
}
