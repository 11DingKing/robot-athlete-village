package httpapi

import (
	"net/http"
	"strings"
)

func contentTypeJSON(r *http.Request) bool {
	value := r.Header.Get("Content-Type")
	if value == "" {
		return true
	}
	return strings.HasPrefix(strings.ToLower(value), "application/json")
}
func requireHeader(r *http.Request, name string) bool {
	return strings.TrimSpace(r.Header.Get(name)) != ""
}
func methodIs(r *http.Request, methods ...string) bool {
	for _, method := range methods {
		if r.Method == method {
			return true
		}
	}
	return false
}
func noStore(w http.ResponseWriter) { w.Header().Set("Cache-Control", "no-store") }
