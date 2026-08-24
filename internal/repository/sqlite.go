package repository

import (
	"context"
	"database/sql"
	"github.com/11DingKing/robot-athlete-village/internal/domain"
	appErr "github.com/11DingKing/robot-athlete-village/internal/errors"
	"time"
)

type SQLite struct{ db *sql.DB }

func NewSQLite(db *sql.DB) *SQLite                 { return &SQLite{db: db} }
func (s *SQLite) DB() *sql.DB                      { return s.db }
func (s *SQLite) Health(ctx context.Context) error { return s.db.PingContext(ctx) }
func (s *SQLite) FindUser(ctx context.Context, email string) (domain.User, error) {
	var u domain.User
	var active int
	err := s.db.QueryRowContext(ctx, "SELECT id,email,role,active FROM users WHERE email=?", email).Scan(&u.ID, &u.Email, &u.Role, &active)
	u.Active = active == 1
	if err == sql.ErrNoRows {
		return u, appErr.ErrNotFound
	}
	return u, err
}
func (s *SQLite) CreateSession(ctx context.Context, u domain.User, token string, expires time.Time) error {
	_, err := s.db.ExecContext(ctx, "INSERT INTO sessions(id,user_id,expires_at,created_at) VALUES(?,?,?,?)", token, u.ID, expires.Format(time.RFC3339), time.Now().UTC().Format(time.RFC3339))
	return err
}
func (s *SQLite) FindSession(ctx context.Context, token string) (domain.User, time.Time, error) {
	var u domain.User
	var active int
	var exp string
	err := s.db.QueryRowContext(ctx, "SELECT u.id,u.email,u.role,u.active,s.expires_at FROM sessions s JOIN users u ON u.id=s.user_id WHERE s.id=? AND s.revoked_at IS NULL", token).Scan(&u.ID, &u.Email, &u.Role, &active, &exp)
	u.Active = active == 1
	if err == sql.ErrNoRows {
		return u, time.Time{}, appErr.ErrUnauthorized
	}
	if err != nil {
		return u, time.Time{}, err
	}
	t, e := time.Parse(time.RFC3339, exp)
	if e != nil {
		return u, time.Time{}, e
	}
	if time.Now().UTC().After(t) {
		return u, t, appErr.ErrExpired
	}
	return u, t, nil
}
func (s *SQLite) RevokeSession(ctx context.Context, token string) error {
	res, err := s.db.ExecContext(ctx, "UPDATE sessions SET revoked_at=? WHERE id=? AND revoked_at IS NULL", time.Now().UTC().Format(time.RFC3339), token)
	if err == nil {
		n, _ := res.RowsAffected()
		if n == 0 {
			return appErr.ErrNotFound
		}
	}
	return err
}
func (s *SQLite) CreateStay(ctx context.Context, did, rid int64, key string, now time.Time) (domain.Stay, error) {
	tx, err := s.db.BeginTx(ctx, nil)
	if err != nil {
		return domain.Stay{}, err
	}
	defer func() { _ = tx.Rollback() }()
	var existing domain.Stay
	var existingStatus, existingCheckIn string
	if queryErr := tx.QueryRowContext(ctx, "SELECT id,delegation_id,room_id,status,check_in,version FROM stays WHERE idempotency_key=?", key).Scan(&existing.ID, &existing.DelegationID, &existing.RoomID, &existingStatus, &existingCheckIn, &existing.Version); queryErr == nil {
		existing.Status = domain.StayStatus(existingStatus)
		existing.IdempotencyKey = key
		existing.CheckIn, _ = time.Parse(time.RFC3339, existingCheckIn)
		return existing, tx.Commit()
	}
	var cap, occ, ver int
	if err = tx.QueryRowContext(ctx, "SELECT capacity,occupied,version FROM rooms WHERE id=?", rid).Scan(&cap, &occ, &ver); err != nil {
		return domain.Stay{}, appErr.Wrap("room_lookup", err)
	}
	if occ >= cap {
		return domain.Stay{}, appErr.ErrCapacity
	}
	if err = tx.Commit(); err != nil {
		return domain.Stay{}, err
	}
	if err = s.reserveRoomBeforeStay(ctx, rid, ver); err != nil {
		return domain.Stay{}, err
	}
	tx, err = s.db.BeginTx(ctx, nil)
	if err != nil {
		return domain.Stay{}, err
	}
	res, err := tx.ExecContext(ctx, "INSERT INTO stays(delegation_id,room_id,status,check_in,idempotency_key,version) VALUES(?,?,?, ?,?,1)", did, rid, domain.StayActive, now.Format(time.RFC3339), key)
	if err != nil {
		if stringsContains(err.Error(), "UNIQUE") {
			var existing domain.Stay
			var status, checkIn string
			if queryErr := tx.QueryRowContext(ctx, "SELECT id,delegation_id,room_id,status,check_in,version FROM stays WHERE idempotency_key=?", key).Scan(&existing.ID, &existing.DelegationID, &existing.RoomID, &status, &checkIn, &existing.Version); queryErr != nil {
				return domain.Stay{}, queryErr
			}
			existing.Status = domain.StayStatus(status)
			existing.CheckIn, _ = time.Parse(time.RFC3339, checkIn)
			return existing, tx.Commit()
		}
		return domain.Stay{}, err
	}
	id, _ := res.LastInsertId()
	if err = tx.Commit(); err != nil {
		return domain.Stay{}, err
	}
	return domain.Stay{ID: id, DelegationID: did, RoomID: rid, Status: domain.StayActive, CheckIn: now, IdempotencyKey: key, Version: 1}, nil
}

func (s *SQLite) reserveRoomBeforeStay(ctx context.Context, roomID int64, version int) error {
	result, err := s.db.ExecContext(ctx, "UPDATE rooms SET occupied=occupied+1,version=version+1 WHERE id=? AND version=?", roomID, version)
	if err != nil {
		return err
	}
	changed, err := result.RowsAffected()
	if err != nil {
		return err
	}
	if changed != 1 {
		return appErr.ErrConflict
	}
	return nil
}
func stringsContains(s, sub string) bool { return len(s) >= len(sub) && contains(s, sub) }
func contains(s, sub string) bool {
	for i := 0; i+len(sub) <= len(s); i++ {
		if s[i:i+len(sub)] == sub {
			return true
		}
	}
	return false
}
func (s *SQLite) findStay(ctx context.Context, key string) (domain.Stay, error) {
	var v domain.Stay
	var st, ci string
	err := s.db.QueryRowContext(ctx, "SELECT id,delegation_id,room_id,status,check_in,version FROM stays WHERE idempotency_key=?", key).Scan(&v.ID, &v.DelegationID, &v.RoomID, &st, &ci, &v.Version)
	v.Status = domain.StayStatus(st)
	v.CheckIn, _ = time.Parse(time.RFC3339, ci)
	return v, err
}
func (s *SQLite) TransitionStay(ctx context.Context, id int64, from, to domain.StayStatus, now time.Time) (domain.Stay, error) {
	if !from.CanTransition(to) {
		return domain.Stay{}, appErr.ErrInvalidState
	}
	tx, err := s.db.BeginTx(ctx, nil)
	if err != nil {
		return domain.Stay{}, err
	}
	defer tx.Rollback()
	var rid int64
	var ver int
	if err = tx.QueryRowContext(ctx, "SELECT room_id,version FROM stays WHERE id=? AND status=?", id, from).Scan(&rid, &ver); err != nil {
		return domain.Stay{}, appErr.ErrConflict
	}
	var out string
	if to == domain.StayClosed {
		out = now.Format(time.RFC3339)
	}
	if _, err = tx.ExecContext(ctx, "UPDATE stays SET status=?,check_out=?,version=version+1 WHERE id=? AND version=?", to, out, id, ver); err != nil {
		return domain.Stay{}, err
	}
	if to == domain.StayClosed {
		if _, err = tx.ExecContext(ctx, "UPDATE rooms SET occupied=occupied-1,version=version+1 WHERE id=? AND occupied>0", rid); err != nil {
			return domain.Stay{}, err
		}
	}
	if err = tx.Commit(); err != nil {
		return domain.Stay{}, err
	}
	return domain.Stay{ID: id, RoomID: rid, Status: to, Version: ver + 1}, nil
}
func (s *SQLite) CreateBooking(ctx context.Context, aid, sid, cid int64, key string) (domain.Booking, error) {
	tx, err := s.db.BeginTx(ctx, nil)
	if err != nil {
		return domain.Booking{}, err
	}
	defer tx.Rollback()
	var status string
	if err = tx.QueryRowContext(ctx, "SELECT status FROM training_slots WHERE id=?", sid).Scan(&status); err != nil {
		return domain.Booking{}, appErr.ErrNotFound
	}
	if status != "open" {
		return domain.Booking{}, appErr.ErrConflict
	}
	var n int
	if err = tx.QueryRowContext(ctx, "SELECT COUNT(*) FROM bookings WHERE slot_id=? AND status IN ('held','confirmed')", sid).Scan(&n); err != nil {
		return domain.Booking{}, err
	}
	var cap int
	if err = tx.QueryRowContext(ctx, "SELECT v.capacity FROM training_slots t JOIN venues v ON v.id=t.venue_id WHERE t.id=?", sid).Scan(&cap); err != nil {
		return domain.Booking{}, err
	}
	if n >= cap {
		return domain.Booking{}, appErr.ErrCapacity
	}
	res, err := tx.ExecContext(ctx, "INSERT INTO bookings(athlete_id,slot_id,coach_id,status,idempotency_key,version) VALUES(?,?,?,?,?,1)", aid, sid, cid, domain.BookingHeld, key)
	if err != nil {
		return domain.Booking{}, err
	}
	id, _ := res.LastInsertId()
	if err = tx.Commit(); err != nil {
		return domain.Booking{}, err
	}
	return domain.Booking{ID: id, AthleteID: aid, SlotID: sid, CoachID: cid, Status: domain.BookingHeld, IdempotencyKey: key, Version: 1}, nil
}
func (s *SQLite) TransitionBooking(ctx context.Context, id int64, from, to domain.BookingStatus) (domain.Booking, error) {
	if !from.CanTransition(to) {
		return domain.Booking{}, appErr.ErrInvalidState
	}
	res, err := s.db.ExecContext(ctx, "UPDATE bookings SET status=?,version=version+1 WHERE id=? AND status=?", to, id, from)
	if err != nil {
		return domain.Booking{}, err
	}
	n, _ := res.RowsAffected()
	if n == 0 {
		return domain.Booking{}, appErr.ErrConflict
	}
	return domain.Booking{ID: id, Status: to}, nil
}
