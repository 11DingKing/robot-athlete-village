package httpapi

import (
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestContentType(t *testing.T) {
	for _, value := range []string{"", "application/json", "application/json; charset=utf-8"} {
		r := httptest.NewRequest(http.MethodPost, "/", nil)
		r.Header.Set("Content-Type", value)
		if !contentTypeJSON(r) {
			t.Fatalf("reject %q", value)
		}
	}
	r := httptest.NewRequest(http.MethodPost, "/", nil)
	r.Header.Set("Content-Type", "text/plain")
	if contentTypeJSON(r) {
		t.Fatal("plain accepted")
	}
}
func TestRequiredHeader(t *testing.T) {
	r := httptest.NewRequest(http.MethodGet, "/", nil)
	if requireHeader(r, "X-Test") {
		t.Fatal("empty accepted")
	}
	r.Header.Set("X-Test", "value")
	if !requireHeader(r, "X-Test") {
		t.Fatal("value rejected")
	}
}
func TestMethodAndNoStore(t *testing.T) {
	r := httptest.NewRequest(http.MethodGet, "/", nil)
	if !methodIs(r, http.MethodPost, http.MethodGet) {
		t.Fatal("method")
	}
	w := httptest.NewRecorder()
	noStore(w)
	if w.Header().Get("Cache-Control") != "no-store" {
		t.Fatal("cache")
	}
}
