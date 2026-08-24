package audit

import (
	"context"
	"github.com/11DingKing/robot-athlete-village/internal/domain"
	"github.com/11DingKing/robot-athlete-village/internal/repository"
	"github.com/11DingKing/robot-athlete-village/internal/telemetry"
	"time"
)

type Logger struct {
	store repository.Store
	now   func() time.Time
}

func New(store repository.Store) *Logger {
	return &Logger{store: store, now: func() time.Time { return time.Now().UTC() }}
}
func (l *Logger) Record(ctx context.Context, actor int64, entity, id, action, result string) error {
	return l.store.AddAudit(ctx, domain.AuditEvent{ActorUserID: actor, EntityType: entity, EntityID: id, Action: action, Result: result, RequestID: telemetry.RequestID(ctx), CreatedAt: l.now()})
}
