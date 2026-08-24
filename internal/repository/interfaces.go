package repository

import (
	"context"
	"database/sql"
	"github.com/11DingKing/robot-athlete-village/internal/domain"
	"time"
)

type Store interface {
	DB() *sql.DB
	FindUser(ctx context.Context, email string) (domain.User, error)
	CreateSession(ctx context.Context, s domain.User, token string, expires time.Time) error
	FindSession(ctx context.Context, token string) (domain.User, time.Time, error)
	RevokeSession(ctx context.Context, token string) error
	CreateStay(ctx context.Context, delegationID, roomID int64, key string, now time.Time) (domain.Stay, error)
	TransitionStay(ctx context.Context, id int64, from, to domain.StayStatus, now time.Time) (domain.Stay, error)
	CreateBooking(ctx context.Context, athleteID, slotID, coachID int64, key string) (domain.Booking, error)
	TransitionBooking(ctx context.Context, id int64, from, to domain.BookingStatus) (domain.Booking, error)
	AssignEquipment(ctx context.Context, equipmentID, athleteID int64) (domain.Equipment, error)
	QueueMaintenance(ctx context.Context, equipmentID int64, now time.Time) (domain.MaintenanceJob, error)
	ClaimMaintenance(ctx context.Context, now time.Time) (domain.MaintenanceJob, error)
	CompleteMaintenance(ctx context.Context, id int64, success bool, errText string, now time.Time) error
	RestoreEquipment(ctx context.Context, equipmentID int64) error
	CompleteMaintenanceWithRestore(ctx context.Context, jobID, equipmentID int64, now time.Time) error
	RecordCheckin(ctx context.Context, athleteID int64, event, key string, now time.Time) error
	AddAudit(ctx context.Context, event domain.AuditEvent) error
	Health(ctx context.Context) error
}
