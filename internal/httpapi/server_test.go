package httpapi

import (
	"bytes"
	"context"
	"encoding/json"
	"github.com/11DingKing/robot-athlete-village/internal/audit"
	"github.com/11DingKing/robot-athlete-village/internal/auth"
	"github.com/11DingKing/robot-athlete-village/internal/middleware"
	"github.com/11DingKing/robot-athlete-village/internal/repository"
	"github.com/11DingKing/robot-athlete-village/internal/service"
	"github.com/11DingKing/robot-athlete-village/internal/storage"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"
)

func testServer(t *testing.T) (http.Handler, repository.Store) {
	t.Helper()
	t.Setenv("MIGRATIONS_DIR", "../../migrations")
	db, e := storage.Open(context.Background(), "file::memory:")
	if e != nil {
		t.Fatal(e)
	}
	t.Cleanup(func() { db.Close() })
	store := repository.NewSQLite(db)
	a := auth.New(store, time.Hour)
	v := service.NewVillage(store, audit.New(store))
	h := middleware.RequestID(New(a, v, service.NewReport(db), store).Routes())
	return h, store
}
func request(t *testing.T, h http.Handler, method, path string, body any, token string) *httptest.ResponseRecorder {
	t.Helper()
	var b bytes.Buffer
	if body != nil {
		_ = json.NewEncoder(&b).Encode(body)
	}
	r := httptest.NewRequest(method, path, &b)
	if token != "" {
		r.Header.Set("Authorization", "Bearer "+token)
	}
	w := httptest.NewRecorder()
	h.ServeHTTP(w, r)
	return w
}
func loginHTTP(t *testing.T, h http.Handler) string {
	w := request(t, h, http.MethodPost, "/v1/login", map[string]string{"Email": "mayor@example.com", "Password": "x"}, "")
	if w.Code != 200 {
		t.Fatalf("login %d %s", w.Code, w.Body.String())
	}
	var out struct{ Token string }
	if e := json.Unmarshal(w.Body.Bytes(), &out); e != nil {
		t.Fatal(e)
	}
	return out.Token
}
func TestHealthReady(t *testing.T) {
	h, _ := testServer(t)
	for _, path := range []string{"/healthz", "/readyz"} {
		w := request(t, h, http.MethodGet, path, nil, "")
		if w.Code != 200 {
			t.Fatalf("%s %d", path, w.Code)
		}
	}
}
func TestLoginLogoutHTTP(t *testing.T) {
	h, _ := testServer(t)
	token := loginHTTP(t, h)
	w := request(t, h, http.MethodPost, "/v1/logout", nil, token)
	if w.Code != 200 {
		t.Fatalf("logout %d", w.Code)
	}
	w = request(t, h, http.MethodGet, "/v1/reports/occupancy", nil, token)
	if w.Code != 401 {
		t.Fatalf("revoked %d", w.Code)
	}
}
func TestProtectedEndpoint(t *testing.T) {
	h, _ := testServer(t)
	w := request(t, h, http.MethodGet, "/v1/reports/occupancy", nil, "")
	if w.Code != 401 {
		t.Fatalf("want 401 got %d", w.Code)
	}
}
func TestStayAndBookingHTTP(t *testing.T) {
	h, _ := testServer(t)
	token := loginHTTP(t, h)
	w := request(t, h, http.MethodPost, "/v1/stays", map[string]any{"DelegationID": 1, "RoomID": 1, "IdempotencyKey": "http-stay"}, token)
	if w.Code != 201 {
		t.Fatalf("stay %d %s", w.Code, w.Body.String())
	}
	w = request(t, h, http.MethodPost, "/v1/training/bookings", map[string]any{"AthleteID": 1, "SlotID": 1, "IdempotencyKey": "http-book"}, token)
	if w.Code != 201 {
		t.Fatalf("booking %d %s", w.Code, w.Body.String())
	}
}
func TestBadJSON(t *testing.T) {
	h, _ := testServer(t)
	token := loginHTTP(t, h)
	r := httptest.NewRequest(http.MethodPost, "/v1/stays", bytes.NewBufferString("not-json"))
	r.Header.Set("Authorization", "Bearer "+token)
	w := httptest.NewRecorder()
	h.ServeHTTP(w, r)
	if w.Code != 400 {
		t.Fatalf("bad json %d", w.Code)
	}
}
func TestRequestID(t *testing.T) {
	h, _ := testServer(t)
	r := httptest.NewRequest(http.MethodGet, "/healthz", nil)
	r.Header.Set("X-Request-ID", "req-fixed")
	w := httptest.NewRecorder()
	h.ServeHTTP(w, r)
	if w.Header().Get("X-Request-ID") != "req-fixed" {
		t.Fatalf("request id %q", w.Header().Get("X-Request-ID"))
	}
}
