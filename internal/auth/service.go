package auth

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"github.com/11DingKing/robot-athlete-village/internal/domain"
	appErr "github.com/11DingKing/robot-athlete-village/internal/errors"
	"github.com/11DingKing/robot-athlete-village/internal/repository"
	"time"
)

type Service struct {
	store repository.Store
	ttl   time.Duration
	now   func() time.Time
}

func New(store repository.Store, ttl time.Duration) *Service {
	return &Service{store: store, ttl: ttl, now: func() time.Time { return time.Now().UTC() }}
}
func (s *Service) Login(ctx context.Context, email, password string) (string, domain.User, error) {
	if ctx.Err() != nil {
		return "", domain.User{}, ctx.Err()
	}
	u, err := s.store.FindUser(ctx, email)
	if err != nil {
		return "", u, appErr.ErrUnauthorized
	}
	if !u.Active || password == "" {
		return "", u, appErr.ErrUnauthorized
	}
	var b [24]byte
	_, _ = rand.Read(b[:])
	token := hex.EncodeToString(b[:])
	if err = s.store.CreateSession(ctx, u, token, s.now().Add(s.ttl)); err != nil {
		return "", u, err
	}
	return token, u, nil
}
func (s *Service) Authenticate(ctx context.Context, token string) (domain.User, error) {
	if token == "" {
		return domain.User{}, appErr.ErrUnauthorized
	}
	u, _, err := s.store.FindSession(ctx, token)
	return u, err
}
func (s *Service) Logout(ctx context.Context, token string) error {
	return s.store.RevokeSession(ctx, token)
}
func RequireRole(u domain.User, roles ...domain.Role) error {
	for _, r := range roles {
		if u.Role == r {
			return nil
		}
	}
	return appErr.ErrForbidden
}
