package audit

import (
	"context"
	"database/sql"
	"fmt"
	"github.com/11DingKing/robot-athlete-village/internal/pagination"
)

type Record struct {
	ID                                                         int64
	EntityType, EntityID, Action, Result, RequestID, CreatedAt string
}

func List(ctx context.Context, db *sql.DB, p pagination.Request, entity string) (pagination.Result[Record], error) {
	where := ""
	args := []any{}
	if entity != "" {
		where = "WHERE entity_type=?"
		args = append(args, entity)
	}
	args = append(args, p.Limit, p.Offset)
	rows, err := db.QueryContext(ctx, "SELECT id,entity_type,entity_id,action,result,request_id,created_at FROM audit_events "+where+" ORDER BY id DESC LIMIT ? OFFSET ?", args...)
	if err != nil {
		return pagination.Result[Record]{}, err
	}
	defer rows.Close()
	out := pagination.Result[Record]{Limit: p.Limit, Offset: p.Offset}
	for rows.Next() {
		var r Record
		if err = rows.Scan(&r.ID, &r.EntityType, &r.EntityID, &r.Action, &r.Result, &r.RequestID, &r.CreatedAt); err != nil {
			return out, err
		}
		out.Items = append(out.Items, r)
	}
	if err = rows.Err(); err != nil {
		return out, err
	}
	countArgs := []any{}
	if entity != "" {
		countArgs = append(countArgs, entity)
	}
	if err = db.QueryRowContext(ctx, "SELECT COUNT(*) FROM audit_events "+where, countArgs...).Scan(&out.Total); err != nil {
		return out, fmt.Errorf("audit count: %w", err)
	}
	return out, nil
}
