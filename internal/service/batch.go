package service

import (
	"context"
	"fmt"
	"github.com/11DingKing/robot-athlete-village/internal/auth"
	"github.com/11DingKing/robot-athlete-village/internal/domain"
	"github.com/11DingKing/robot-athlete-village/internal/repository"
	"sync"
	"time"
)

type BatchResult struct {
	Key     string
	Success bool
	Message string
}
type Batch struct{ village *Village }

func NewBatch(v *Village) *Batch { return &Batch{village: v} }
func (b *Batch) AdmitMany(ctx context.Context, u domain.User, requests []domain.AdmissionRequest) []BatchResult {
	out := make([]BatchResult, len(requests))
	var wg sync.WaitGroup
	for i, req := range requests {
		wg.Add(1)
		go func(i int, req domain.AdmissionRequest) {
			defer wg.Done()
			if err := req.Validate(); err != nil {
				out[i] = BatchResult{Key: req.IdempotencyKey, Message: err.Error()}
				return
			}
			_, err := b.village.Admit(ctx, u, req.DelegationID, req.RoomID, req.IdempotencyKey)
			out[i] = BatchResult{Key: req.IdempotencyKey, Success: err == nil}
			if err != nil {
				out[i].Message = err.Error()
			}
		}(i, req)
	}
	wg.Wait()
	return out
}
func (b *Batch) BookAndConfirm(ctx context.Context, u domain.User, req domain.BookingRequest) (domain.Booking, error) {
	if err := req.Validate(); err != nil {
		return domain.Booking{}, err
	}
	if err := auth.RequireRole(u, domain.RoleCoach, domain.RoleAdmin); err != nil {
		return domain.Booking{}, err
	}
	booking, err := b.village.ReserveTraining(ctx, u, req.AthleteID, req.SlotID, req.IdempotencyKey)
	if err != nil {
		return booking, err
	}
	confirmed, confirmErr := b.village.ConfirmTraining(ctx, u, booking.ID)
	if confirmErr != nil {
		_ = b.village.CompensateFailedConfirmation(ctx, u, booking.ID)
		return booking, confirmErr
	}
	return confirmed, nil
}
func (b *Batch) CloseExpired(ctx context.Context, u domain.User, ids []int64, now time.Time) int {
	closed := 0
	for _, id := range ids {
		if _, err := b.village.CloseStay(ctx, u, id); err == nil {
			closed++
		}
	}
	return closed
}
func ValidateBatchSize(requests []domain.AdmissionRequest) error {
	if len(requests) == 0 {
		return fmt.Errorf("batch cannot be empty")
	}
	if len(requests) > 100 {
		return fmt.Errorf("batch exceeds 100")
	}
	return nil
}

var _ repository.Store
