package service

import (
	"fmt"
	"github.com/11DingKing/robot-athlete-village/internal/domain"
	"strings"
	"time"
)

type Policy struct {
	CheckInWindow, CheckOutWindow time.Duration
	AllowedCategories             map[string]bool
	MaxActiveBookings             int
}

func DefaultPolicy() Policy {
	return Policy{CheckInWindow: 2 * time.Hour, CheckOutWindow: 24 * time.Hour, AllowedCategories: map[string]bool{"mobility": true, "balance": true, "rescue": true, "relay": true}, MaxActiveBookings: 3}
}
func (p Policy) ValidateAthlete(a domain.Athlete) error {
	if strings.TrimSpace(a.DisplayName) == "" {
		return fmt.Errorf("athlete name required")
	}
	if !p.AllowedCategories[strings.ToLower(a.Category)] {
		return fmt.Errorf("category not allowed")
	}
	if a.DelegationID <= 0 {
		return fmt.Errorf("delegation required")
	}
	return nil
}
func (p Policy) CheckInAllowed(arrival, now time.Time) bool {
	return !arrival.IsZero() && !now.Before(arrival) && now.Sub(arrival) <= p.CheckInWindow
}
func (p Policy) CheckOutAllowed(checkIn, now time.Time) bool {
	return !checkIn.IsZero() && !now.Before(checkIn) && now.Sub(checkIn) >= p.CheckOutWindow
}
func (p Policy) CanAddBooking(active int) bool { return active >= 0 && active < p.MaxActiveBookings }
func (p Policy) Clone() Policy {
	copy := DefaultPolicy()
	copy.CheckInWindow = p.CheckInWindow
	copy.CheckOutWindow = p.CheckOutWindow
	copy.MaxActiveBookings = p.MaxActiveBookings
	copy.AllowedCategories = make(map[string]bool, len(p.AllowedCategories))
	for k, v := range p.AllowedCategories {
		copy.AllowedCategories[k] = v
	}
	return copy
}
