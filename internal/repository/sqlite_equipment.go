package repository

import (
	"context"
	"database/sql"
	"github.com/11DingKing/robot-athlete-village/internal/domain"
	appErr "github.com/11DingKing/robot-athlete-village/internal/errors"
	"time"
)

func (s *SQLite) AssignEquipment(ctx context.Context, eid, aid int64) (domain.Equipment, error) {
	tx, err := s.db.BeginTx(ctx, nil)
	if err != nil {
		return domain.Equipment{}, err
	}
	defer tx.Rollback()
	var status string
	var ver int
	if err = tx.QueryRowContext(ctx, "SELECT status,version FROM equipment WHERE id=?", eid).Scan(&status, &ver); err != nil {
		return domain.Equipment{}, appErr.ErrNotFound
	}
	if status != "ready" {
		return domain.Equipment{}, appErr.ErrConflict
	}
	if _, err = tx.ExecContext(ctx, "UPDATE equipment SET status='in_use',assigned_athlete_id=?,version=version+1 WHERE id=? AND version=?", aid, eid, ver); err != nil {
		return domain.Equipment{}, err
	}
	if err = tx.Commit(); err != nil {
		return domain.Equipment{}, err
	}
	return domain.Equipment{ID: eid, AssignedAthleteID: &aid, Status: domain.EquipmentInUse, Version: ver + 1}, nil
}
func (s *SQLite) QueueMaintenance(ctx context.Context, eid int64, now time.Time) (domain.MaintenanceJob, error) {
	tx, err := s.db.BeginTx(ctx, nil)
	if err != nil {
		return domain.MaintenanceJob{}, err
	}
	defer tx.Rollback()
	result, err := tx.ExecContext(ctx, "UPDATE equipment SET status='maintenance',assigned_athlete_id=NULL,version=version+1 WHERE id=? AND status!='maintenance'", eid)
	if err != nil {
		return domain.MaintenanceJob{}, err
	}
	changed, _ := result.RowsAffected()
	if changed != 1 {
		return domain.MaintenanceJob{}, appErr.ErrConflict
	}
	res, err := tx.ExecContext(ctx, "INSERT INTO maintenance_jobs(equipment_id,status,attempts,next_run_at) VALUES(?,?,0,?)", eid, "queued", now.Format(time.RFC3339))
	if err != nil {
		return domain.MaintenanceJob{}, err
	}
	id, _ := res.LastInsertId()
	if err = tx.Commit(); err != nil {
		return domain.MaintenanceJob{}, err
	}
	return domain.MaintenanceJob{ID: id, EquipmentID: eid, Status: "queued", NextRunAt: now}, nil
}
func (s *SQLite) ClaimMaintenance(ctx context.Context, now time.Time) (domain.MaintenanceJob, error) {
	tx, err := s.db.BeginTx(ctx, nil)
	if err != nil {
		return domain.MaintenanceJob{}, err
	}
	defer tx.Rollback()
	var j domain.MaintenanceJob
	var run string
	err = tx.QueryRowContext(ctx, "SELECT id,equipment_id,attempts,next_run_at FROM maintenance_jobs WHERE status IN ('queued','retry') AND next_run_at<=? ORDER BY id LIMIT 1", now.Format(time.RFC3339)).Scan(&j.ID, &j.EquipmentID, &j.Attempts, &run)
	if err == sql.ErrNoRows {
		return j, appErr.ErrNotFound
	}
	if err != nil {
		return j, err
	}
	if _, err = tx.ExecContext(ctx, "UPDATE maintenance_jobs SET status='running',attempts=attempts+1 WHERE id=?", j.ID); err != nil {
		return j, err
	}
	if err = tx.Commit(); err != nil {
		return j, err
	}
	j.Status = "running"
	j.Attempts++
	j.NextRunAt, _ = time.Parse(time.RFC3339, run)
	return j, nil
}
func (s *SQLite) CompleteMaintenance(ctx context.Context, id int64, success bool, errText string, now time.Time) error {
	status := "retry"
	next := now.Add(time.Minute)
	if success {
		status = "done"
		next = now
	}
	_, err := s.db.ExecContext(ctx, "UPDATE maintenance_jobs SET status=?,last_error=?,next_run_at=? WHERE id=? AND status='running'", status, errText, next.Format(time.RFC3339), id)
	return err
}
func (s *SQLite) RestoreEquipment(ctx context.Context, equipmentID int64) error {
	result, err := s.db.ExecContext(ctx, "UPDATE equipment SET status='ready',assigned_athlete_id=NULL,version=version+1 WHERE id=? AND status='maintenance'", equipmentID)
	if err != nil {
		return err
	}
	changed, _ := result.RowsAffected()
	if changed != 1 {
		return appErr.ErrConflict
	}
	return nil
}
func (s *SQLite) CompleteMaintenanceWithRestore(ctx context.Context, jobID, equipmentID int64, now time.Time) error {
	tx, err := s.db.BeginTx(ctx, nil)
	if err != nil {
		return err
	}
	defer tx.Rollback()
	restored, err := tx.ExecContext(ctx, "UPDATE equipment SET status='ready',assigned_athlete_id=NULL,version=version+1 WHERE id=? AND status='maintenance'", equipmentID)
	if err != nil {
		return err
	}
	restoredRows, _ := restored.RowsAffected()
	if restoredRows != 1 {
		return appErr.ErrConflict
	}
	completed, err := tx.ExecContext(ctx, "UPDATE maintenance_jobs SET status='done',last_error='',next_run_at=? WHERE id=? AND status='running'", now.Format(time.RFC3339), jobID)
	if err != nil {
		return err
	}
	completedRows, _ := completed.RowsAffected()
	if completedRows != 1 {
		return appErr.ErrConflict
	}
	return tx.Commit()
}
func (s *SQLite) RecordCheckin(ctx context.Context, aid int64, event, key string, now time.Time) error {
	_, err := s.db.ExecContext(ctx, "INSERT OR IGNORE INTO checkins(athlete_id,event_code,checked_at,idempotency_key) VALUES(?,?,?,?)", aid, event, now.Format(time.RFC3339), key)
	return err
}
func (s *SQLite) AddAudit(ctx context.Context, e domain.AuditEvent) error {
	_, err := s.db.ExecContext(ctx, "INSERT INTO audit_events(actor_user_id,entity_type,entity_id,action,result,request_id,created_at) VALUES(?,?,?,?,?,?,?)", e.ActorUserID, e.EntityType, e.EntityID, e.Action, e.Result, e.RequestID, e.CreatedAt.Format(time.RFC3339))
	return err
}
