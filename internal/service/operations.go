package service

import (
	"context"
	"fmt"
	"github.com/11DingKing/robot-athlete-village/internal/auth"
	"github.com/11DingKing/robot-athlete-village/internal/domain"
	"github.com/11DingKing/robot-athlete-village/internal/repository"
	"time"
)

type Operations struct {
	store  repository.Store
	policy Policy
}

func NewOperations(store repository.Store, policy Policy) *Operations {
	return &Operations{store: store, policy: policy}
}
func (o *Operations) ValidateAthlete(ctx context.Context, id int64) error {
	r, ok := o.store.(*repository.SQLite)
	if !ok {
		return fmt.Errorf("unsupported store")
	}
	a, e := r.FindAthlete(ctx, id)
	if e != nil {
		return e
	}
	return o.policy.ValidateAthlete(a)
}
func (o *Operations) CloseIfEligible(ctx context.Context, u domain.User, stayID int64, checkIn, now time.Time) (domain.Stay, error) {
	if err := auth.RequireRole(u, domain.RoleAdmin); err != nil {
		return domain.Stay{}, err
	}
	if !o.policy.CheckOutAllowed(checkIn, now) {
		return domain.Stay{}, fmt.Errorf("checkout window not reached")
	}
	v := NewVillage(o.store, nil)
	return v.CloseStay(ctx, u, stayID)
}
func (o *Operations) ReserveIfAllowed(ctx context.Context, u domain.User, req domain.BookingRequest, active int) (domain.Booking, error) {
	if !o.policy.CanAddBooking(active) {
		return domain.Booking{}, fmt.Errorf("booking limit reached")
	}
	return NewBatch(NewVillage(o.store, nil)).BookAndConfirm(ctx, u, req)
}
func (o *Operations) IsCategoryAllowed(category string) bool {
	return o.policy.AllowedCategories[category]
}
