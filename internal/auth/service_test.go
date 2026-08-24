package auth

import (
	"context"
	"github.com/11DingKing/robot-athlete-village/internal/domain"
	appErr "github.com/11DingKing/robot-athlete-village/internal/errors"
	"github.com/11DingKing/robot-athlete-village/internal/repository"
	"github.com/11DingKing/robot-athlete-village/internal/storage"
	"testing"
	"time"
)

func authService(t *testing.T) *Service {
	t.Helper()
	t.Setenv("MIGRATIONS_DIR", "../../migrations")
	db, e := storage.Open(context.Background(), "file::memory:")
	if e != nil {
		t.Fatal(e)
	}
	t.Cleanup(func() { db.Close() })
	return New(repository.NewSQLite(db), time.Hour)
}
func TestLoginReturnsToken(t *testing.T) {
	s := authService(t)
	token, u, e := s.Login(context.Background(), "mayor@example.com", "secret")
	if e != nil || token == "" || u.Role != domain.RoleAdmin {
		t.Fatalf("token %q user %+v err %v", token, u, e)
	}
}
func TestLoginRejectsUnknown(t *testing.T) {
	s := authService(t)
	for _, email := range []string{"missing@example.com", "", "coach@example.com"} {
		token, _, e := s.Login(context.Background(), email, "")
		if e != appErr.ErrUnauthorized || token != "" {
			t.Fatalf("email %q token %q err %v", email, token, e)
		}
	}
}
func TestAuthenticateLogout(t *testing.T) {
	s := authService(t)
	token, _, e := s.Login(context.Background(), "coach@example.com", "x")
	if e != nil {
		t.Fatal(e)
	}
	u, e := s.Authenticate(context.Background(), token)
	if e != nil || u.Role != domain.RoleCoach {
		t.Fatalf("%+v %v", u, e)
	}
	if e = s.Logout(context.Background(), token); e != nil {
		t.Fatal(e)
	}
	if _, e = s.Authenticate(context.Background(), token); e != appErr.ErrUnauthorized {
		t.Fatalf("after logout %v", e)
	}
}
func TestEmptyToken(t *testing.T) {
	s := authService(t)
	if _, e := s.Authenticate(context.Background(), ""); e != appErr.ErrUnauthorized {
		t.Fatal(e)
	}
}
func TestRoleCheck(t *testing.T) {
	admin := domain.User{Role: domain.RoleAdmin}
	coach := domain.User{Role: domain.RoleCoach}
	if e := RequireRole(admin, domain.RoleAdmin); e != nil {
		t.Fatal(e)
	}
	if e := RequireRole(coach, domain.RoleAdmin); e != appErr.ErrForbidden {
		t.Fatal(e)
	}
	if e := RequireRole(coach, domain.RoleCoach, domain.RoleAdmin); e != nil {
		t.Fatal(e)
	}
}
func TestExpiredSession(t *testing.T) {
	s := authService(t)
	s.now = func() time.Time { return time.Now().UTC().Add(-2 * time.Hour) }
	token, _, e := s.Login(context.Background(), "coach@example.com", "x")
	if e != nil {
		t.Fatal(e)
	}
	if _, e = s.Authenticate(context.Background(), token); e != appErr.ErrExpired {
		t.Fatalf("want expired got %v", e)
	}
}
