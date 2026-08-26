package httpapi

import (
	"encoding/json"
	"errors"
	"github.com/11DingKing/robot-athlete-village/internal/auth"
	"github.com/11DingKing/robot-athlete-village/internal/domain"
	appErr "github.com/11DingKing/robot-athlete-village/internal/errors"
	"github.com/11DingKing/robot-athlete-village/internal/pagination"
	"github.com/11DingKing/robot-athlete-village/internal/repository"
	"github.com/11DingKing/robot-athlete-village/internal/service"
	"net/http"
	"strconv"
	"strings"
)

type Server struct {
	auth    *auth.Service
	village *service.Village
	report  *service.Report
	store   repository.Store
}

func New(a *auth.Service, v *service.Village, r *service.Report, s repository.Store) *Server {
	return &Server{auth: a, village: v, report: r, store: s}
}
func (s *Server) Routes() http.Handler {
	mux := http.NewServeMux()
	mux.HandleFunc("/healthz", s.health)
	mux.HandleFunc("/readyz", s.ready)
	mux.HandleFunc("/v1/login", s.login)
	mux.HandleFunc("/v1/logout", s.logout)
	mux.HandleFunc("/v1/stays", s.stays)
	mux.HandleFunc("/v1/training/bookings", s.bookings)
	mux.HandleFunc("/v1/checkins", s.checkins)
	mux.HandleFunc("/v1/reports/occupancy", s.occupancy)
	return mux
}
func (s *Server) health(w http.ResponseWriter, _ *http.Request) {
	w.WriteHeader(http.StatusOK)
	_, _ = w.Write([]byte(`{"status":"ok"}`))
}
func (s *Server) ready(w http.ResponseWriter, r *http.Request) {
	if err := s.store.Health(r.Context()); err != nil {
		http.Error(w, `{"status":"not_ready"}`, 503)
		return
	}
	w.WriteHeader(200)
	_, _ = w.Write([]byte(`{"status":"ready"}`))
}

type loginRequest struct{ Email, Password string }

func (s *Server) login(w http.ResponseWriter, r *http.Request) {
	var in loginRequest
	if json.NewDecoder(r.Body).Decode(&in) != nil {
		errorJSON(w, 400, "invalid_json")
		return
	}
	token, u, err := s.auth.Login(r.Context(), in.Email, in.Password)
	if err != nil {
		errorJSON(w, 401, "unauthorized")
		return
	}
	jsonResponse(w, 200, map[string]any{"token": token, "user": u})
}
func (s *Server) logout(w http.ResponseWriter, r *http.Request) {
	token := bearer(r)
	if err := s.auth.Logout(r.Context(), token); err != nil {
		errorJSON(w, 404, "session_not_found")
		return
	}
	jsonResponse(w, 200, map[string]string{"status": "revoked"})
}
func (s *Server) stays(w http.ResponseWriter, r *http.Request) {
	u, ok := s.user(w, r)
	if !ok {
		return
	}
	if r.Method == http.MethodPost {
		var in struct {
			DelegationID, RoomID int64
			IdempotencyKey       string
		}
		if json.NewDecoder(r.Body).Decode(&in) != nil {
			errorJSON(w, 400, "invalid_json")
			return
		}
		st, err := s.village.Admit(r.Context(), u, in.DelegationID, in.RoomID, in.IdempotencyKey)
		if err != nil {
			mapError(w, err)
			return
		}
		jsonResponse(w, 201, st)
		return
	}
	if r.Method == http.MethodPatch {
		id, _ := strconv.ParseInt(r.URL.Query().Get("id"), 10, 64)
		st, err := s.village.CloseStay(r.Context(), u, id)
		if err != nil {
			mapError(w, err)
			return
		}
		jsonResponse(w, 200, st)
		return
	}
	errorJSON(w, 405, "method_not_allowed")
}
func (s *Server) bookings(w http.ResponseWriter, r *http.Request) {
	u, ok := s.user(w, r)
	if !ok {
		return
	}
	if r.Method == http.MethodPost {
		var in struct {
			AthleteID, SlotID int64
			IdempotencyKey    string
		}
		if json.NewDecoder(r.Body).Decode(&in) != nil {
			errorJSON(w, 400, "invalid_json")
			return
		}
		b, err := s.village.ReserveTraining(r.Context(), u, in.AthleteID, in.SlotID, in.IdempotencyKey)
		if err != nil {
			mapError(w, err)
			return
		}
		jsonResponse(w, 201, b)
		return
	}
	if r.Method == http.MethodPut {
		id, _ := strconv.ParseInt(r.URL.Query().Get("id"), 10, 64)
		b, err := s.village.ConfirmTraining(r.Context(), u, id)
		if err != nil {
			mapError(w, err)
			return
		}
		jsonResponse(w, 200, b)
		return
	}
	errorJSON(w, 405, "method_not_allowed")
}
func (s *Server) checkins(w http.ResponseWriter, r *http.Request) {
	u, ok := s.user(w, r)
	if !ok {
		return
	}
	var in struct {
		AthleteID                 int64
		EventCode, IdempotencyKey string
	}
	if json.NewDecoder(r.Body).Decode(&in) != nil {
		errorJSON(w, 400, "invalid_json")
		return
	}
	if err := s.village.Checkin(r.Context(), u, in.AthleteID, in.EventCode, in.IdempotencyKey); err != nil {
		mapError(w, err)
		return
	}
	jsonResponse(w, 201, map[string]string{"status": "checked_in"})
}
func (s *Server) occupancy(w http.ResponseWriter, r *http.Request) {
	if _, ok := s.user(w, r); !ok {
		return
	}
	p := pagination.Normalize(parseInt(r.URL.Query().Get("limit")), parseInt(r.URL.Query().Get("offset")))
	v, err := s.report.Occupancy(r.Context(), p)
	if err != nil {
		mapError(w, err)
		return
	}
	jsonResponse(w, 200, v)
}
func (s *Server) user(w http.ResponseWriter, r *http.Request) (domain.User, bool) {
	u, err := s.auth.Authenticate(r.Context(), bearer(r))
	if err != nil {
		mapError(w, err)
		return domain.User{}, false
	}
	return u, true
}
func bearer(r *http.Request) string {
	return strings.TrimPrefix(r.Header.Get("Authorization"), "Bearer ")
}
func parseInt(v string) int { n, _ := strconv.Atoi(v); return n }
func jsonResponse(w http.ResponseWriter, status int, v any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(v)
}
func errorJSON(w http.ResponseWriter, status int, code string) {
	jsonResponse(w, status, map[string]string{"error": code})
}
func mapError(w http.ResponseWriter, err error) {
	status := 500
	code := "internal_error"
	switch {
	case errors.Is(err, appErr.ErrUnauthorized), errors.Is(err, appErr.ErrExpired):
		status = 401
		code = "unauthorized"
	case errors.Is(err, appErr.ErrForbidden):
		status = 403
		code = "forbidden"
	case errors.Is(err, appErr.ErrConflict), errors.Is(err, appErr.ErrCapacity):
		status = 409
		code = "conflict"
	case errors.Is(err, appErr.ErrNotFound):
		status = 404
		code = "not_found"
	case errors.Is(err, appErr.ErrInvalidState):
		status = 422
		code = "invalid_state"
	}
	errorJSON(w, status, code)
}
