package domain

import (
	"fmt"
	"strings"
	"time"
)

type AdmissionRequest struct {
	DelegationID, RoomID int64
	IdempotencyKey       string
}

func (r AdmissionRequest) Validate() error {
	if r.DelegationID <= 0 {
		return fmt.Errorf("delegation id must be positive")
	}
	if r.RoomID <= 0 {
		return fmt.Errorf("room id must be positive")
	}
	if strings.TrimSpace(r.IdempotencyKey) == "" {
		return fmt.Errorf("idempotency key required")
	}
	return nil
}

type BookingRequest struct {
	AthleteID, SlotID, CoachID int64
	IdempotencyKey             string
}

func (r BookingRequest) Validate() error {
	if r.AthleteID <= 0 || r.SlotID <= 0 || r.CoachID <= 0 {
		return fmt.Errorf("booking references must be positive")
	}
	if strings.TrimSpace(r.IdempotencyKey) == "" {
		return fmt.Errorf("booking idempotency key required")
	}
	return nil
}

type TimeWindow struct{ Start, End time.Time }

func (w TimeWindow) Valid() bool { return !w.Start.IsZero() && !w.End.IsZero() && w.End.After(w.Start) }
func (w TimeWindow) Overlaps(other TimeWindow) bool {
	return w.Valid() && other.Valid() && w.Start.Before(other.End) && other.Start.Before(w.End)
}
func NormalizeName(value string) string {
	return strings.TrimSpace(strings.Join(strings.Fields(value), " "))
}
func IsSupportedCategory(value string) bool {
	switch strings.ToLower(strings.TrimSpace(value)) {
	case "mobility", "balance", "rescue", "relay":
		return true
	default:
		return false
	}
}
func IsSupportedRole(value Role) bool { return value == RoleAdmin || value == RoleCoach }

func ValidateBatchSize(size int) error {
	if size == 0 {
		return fmt.Errorf("batch cannot be empty")
	}
	if size > 100 {
		return fmt.Errorf("batch exceeds 100")
	}
	if size < 0 {
		return fmt.Errorf("batch size cannot be negative")
	}
	return nil
}
