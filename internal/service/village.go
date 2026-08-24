package service

import (
	"context"
	"fmt"
	"github.com/11DingKing/robot-athlete-village/internal/audit"
	"github.com/11DingKing/robot-athlete-village/internal/auth"
	"github.com/11DingKing/robot-athlete-village/internal/domain"
	appErr "github.com/11DingKing/robot-athlete-village/internal/errors"
	"github.com/11DingKing/robot-athlete-village/internal/repository"
	"time"
)

type Village struct {
	store repository.Store
	audit *audit.Logger
	now   func() time.Time
}

func (v *Village) record(ctx context.Context, actor int64, entity, id, action, result string) {
	if v.audit != nil {
		_ = v.audit.Record(ctx, actor, entity, id, action, result)
	}
}

func NewVillage(store repository.Store, a *audit.Logger) *Village {
	return &Village{store: store, audit: a, now: func() time.Time { return time.Now().UTC() }}
}
func (v *Village) Admit(ctx context.Context, u domain.User, did, rid int64, key string) (domain.Stay, error) {
	if err := auth.RequireRole(u, domain.RoleAdmin); err != nil {
		return domain.Stay{}, err
	}
	st, err := v.store.CreateStay(ctx, did, rid, key, v.now())
	if err != nil {
		v.record(ctx, u.ID, "stay", fmt.Sprint(did), "admit", "failed")
		return st, err
	}
	v.record(ctx, u.ID, "stay", fmt.Sprint(st.ID), "admit", "success")
	return st, nil
}
func (v *Village) CloseStay(ctx context.Context, u domain.User, id int64) (domain.Stay, error) {
	if err := auth.RequireRole(u, domain.RoleAdmin); err != nil {
		return domain.Stay{}, err
	}
	st, err := v.store.TransitionStay(ctx, id, domain.StayActive, domain.StayClosed, v.now())
	if err != nil {
		return st, err
	}
	v.record(ctx, u.ID, "stay", fmt.Sprint(id), "close", "success")
	return st, nil
}
func (v *Village) ReserveTraining(ctx context.Context, u domain.User, aid, sid int64, key string) (domain.Booking, error) {
	if err := auth.RequireRole(u, domain.RoleCoach, domain.RoleAdmin); err != nil {
		return domain.Booking{}, err
	}
	b, err := v.store.CreateBooking(ctx, aid, sid, u.ID, key)
	if err != nil {
		return b, err
	}
	v.record(ctx, u.ID, "booking", fmt.Sprint(b.ID), "hold", "success")
	return b, nil
}
func (v *Village) ConfirmTraining(ctx context.Context, u domain.User, id int64) (domain.Booking, error) {
	if err := auth.RequireRole(u, domain.RoleCoach, domain.RoleAdmin); err != nil {
		return domain.Booking{}, err
	}
	b, err := v.store.TransitionBooking(ctx, id, domain.BookingHeld, domain.BookingConfirmed)
	if err != nil {
		return b, err
	}
	v.record(ctx, u.ID, "booking", fmt.Sprint(id), "confirm", "success")
	return b, nil
}
func (v *Village) CancelTraining(ctx context.Context, u domain.User, id int64) (domain.Booking, error) {
	if err := auth.RequireRole(u, domain.RoleCoach, domain.RoleAdmin); err != nil {
		return domain.Booking{}, err
	}
	return v.store.TransitionBooking(ctx, id, domain.BookingConfirmed, domain.BookingCancelled)
}
func (v *Village) Assign(ctx context.Context, u domain.User, eid, aid int64) (domain.Equipment, error) {
	if err := auth.RequireRole(u, domain.RoleCoach, domain.RoleAdmin); err != nil {
		return domain.Equipment{}, err
	}
	athlete, err := v.store.FindAthlete(ctx, aid)
	if err != nil {
		return domain.Equipment{}, err
	}
	if athlete.Status != string(domain.AthleteReady) {
		return domain.Equipment{}, appErr.ErrInvalidState
	}
	return v.store.AssignEquipment(ctx, eid, aid)
}
func (v *Village) Checkin(ctx context.Context, u domain.User, aid int64, event, key string) error {
	if err := auth.RequireRole(u, domain.RoleCoach, domain.RoleAdmin); err != nil {
		return err
	}
	return v.store.RecordCheckin(ctx, aid, event, key, v.now())
}
func (v *Village) StartMaintenance(ctx context.Context, u domain.User, eid int64) error {
	if err := auth.RequireRole(u, domain.RoleAdmin); err != nil {
		return err
	}
	_, err := v.store.QueueMaintenance(ctx, eid, v.now())
	return err
}

var _ = appErr.ErrNotFound
