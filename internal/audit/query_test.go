package audit

import (
	"context"
	"github.com/11DingKing/robot-athlete-village/internal/domain"
	"github.com/11DingKing/robot-athlete-village/internal/pagination"
	"github.com/11DingKing/robot-athlete-village/internal/repository"
	"github.com/11DingKing/robot-athlete-village/internal/storage"
	"testing"
	"time"
)

func TestListAudit(t *testing.T) {
	t.Setenv("MIGRATIONS_DIR", "../../migrations")
	db, e := storage.Open(context.Background(), "file::memory:")
	if e != nil {
		t.Fatal(e)
	}
	defer db.Close()
	store := repository.NewSQLite(db)
	for i := 0; i < 4; i++ {
		if e := store.AddAudit(context.Background(), domain.AuditEvent{ActorUserID: 1, EntityType: "stay", EntityID: "1", Action: "step", Result: "success", RequestID: string(rune('a' + i)), CreatedAt: time.Now().UTC()}); e != nil {
			t.Fatal(e)
		}
	}
	all, e := List(context.Background(), store.DB(), pagination.Normalize(10, 0), "")
	if e != nil || all.Total != 4 || len(all.Items) != 4 {
		t.Fatalf("%+v %v", all, e)
	}
	filtered, e := List(context.Background(), store.DB(), pagination.Normalize(10, 0), "stay")
	if e != nil || filtered.Total != 4 {
		t.Fatalf("%+v %v", filtered, e)
	}
}
