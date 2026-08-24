package domain

import "time"

type Role string

const (
	RoleAdmin Role = "village_admin"
	RoleCoach Role = "coach"
)

type StayStatus string

const (
	StayPending StayStatus = "pending"
	StayActive  StayStatus = "active"
	StayClosed  StayStatus = "closed"
)

type BookingStatus string

const (
	BookingHeld      BookingStatus = "held"
	BookingConfirmed BookingStatus = "confirmed"
	BookingCancelled BookingStatus = "cancelled"
)

type AthleteStatus string

const (
	AthleteReady      AthleteStatus = "ready"
	AthleteWithdrawn  AthleteStatus = "withdrawn"
)

type EquipmentStatus string

const (
	EquipmentReady       EquipmentStatus = "ready"
	EquipmentInUse       EquipmentStatus = "in_use"
	EquipmentMaintenance EquipmentStatus = "maintenance"
)

type User struct {
	ID     int64
	Email  string
	Role   Role
	Active bool
}
type Delegation struct {
	ID                        int64
	Name, CountryCode, Status string
	CreatedAt                 time.Time
}
type Athlete struct {
	ID, DelegationID              int64
	DisplayName, Category, Status string
	CreatedAt                     time.Time
}
type Room struct {
	ID                          int64
	Code                        string
	Capacity, Occupied, Version int
}
type Stay struct {
	ID, DelegationID, RoomID int64
	Status                   StayStatus
	CheckIn                  time.Time
	CheckOut                 *time.Time
	IdempotencyKey           string
	Version                  int
}
type Venue struct {
	ID          int64
	Name, Sport string
	Capacity    int
	Active      bool
}
type TrainingSlot struct {
	ID, VenueID      int64
	StartsAt, EndsAt time.Time
	Status           string
}
type Booking struct {
	ID, AthleteID, SlotID, CoachID int64
	Status                         BookingStatus
	IdempotencyKey                 string
	Version                        int
}
type Equipment struct {
	ID                int64
	Serial, Kind      string
	Status            EquipmentStatus
	AssignedAthleteID *int64
	Version           int
}
type MaintenanceJob struct {
	ID, EquipmentID int64
	Status          string
	Attempts        int
	NextRunAt       time.Time
	LastError       string
}
type AuditEvent struct {
	ID, ActorUserID                                 int64
	EntityType, EntityID, Action, Result, RequestID string
	CreatedAt                                       time.Time
}

func (s StayStatus) CanTransition(to StayStatus) bool {
	return (s == StayPending && to == StayActive) || (s == StayActive && to == StayClosed)
}
func (b BookingStatus) CanTransition(to BookingStatus) bool {
	return (b == BookingHeld && to == BookingConfirmed) || (b == BookingHeld && to == BookingCancelled) || (b == BookingConfirmed && to == BookingCancelled)
}
func (e EquipmentStatus) CanTransition(to EquipmentStatus) bool {
	return (e == EquipmentReady && to == EquipmentInUse) || (e == EquipmentInUse && to == EquipmentReady) || (e == EquipmentReady && to == EquipmentMaintenance) || (e == EquipmentMaintenance && to == EquipmentReady)
}
