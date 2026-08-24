package worker

import (
	"context"
	"github.com/11DingKing/robot-athlete-village/internal/repository"
	"log/slog"
	"time"
)

type Maintenance struct {
	store    repository.Store
	interval time.Duration
	log      *slog.Logger
}

func NewMaintenance(store repository.Store, interval time.Duration, log *slog.Logger) *Maintenance {
	return &Maintenance{store: store, interval: interval, log: log}
}
func (w *Maintenance) Run(ctx context.Context) {
	ticker := time.NewTicker(w.interval)
	defer ticker.Stop()
	for {
		select {
		case <-ctx.Done():
			return
		case now := <-ticker.C:
			w.process(ctx, now)
		}
	}
}
func (w *Maintenance) process(ctx context.Context, now time.Time) {
	job, err := w.store.ClaimMaintenance(ctx, now)
	if err != nil {
		return
	}
	if err = w.store.CompleteMaintenance(ctx, job.ID, true, "", now); err != nil {
		w.log.Error("maintenance completion failed", "job", job.ID, "error", err)
		return
	}
	if err = w.store.RestoreEquipment(ctx, job.EquipmentID); err != nil {
		w.log.Error("equipment restore failed", "job", job.ID, "equipment", job.EquipmentID, "error", err)
	}
}
