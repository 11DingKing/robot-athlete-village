package repository

import (
	"context"
	"github.com/11DingKing/robot-athlete-village/internal/domain"
	"github.com/11DingKing/robot-athlete-village/internal/pagination"
	"strings"
)

func (s *SQLite) ListAthletes(ctx context.Context, p pagination.Request, f domain.AthleteFilter) (pagination.Result[domain.Athlete], error) {
	f = f.Normalize()
	where := "WHERE 1=1"
	args := []any{}
	if f.Status != "" {
		where += " AND status=?"
		args = append(args, f.Status)
	}
	if f.Category != "" {
		where += " AND category=?"
		args = append(args, f.Category)
	}
	if f.Search != "" {
		where += " AND lower(display_name) LIKE ?"
		args = append(args, "%"+strings.ToLower(f.Search)+"%")
	}
	args = append(args, p.Limit, p.Offset)
	rows, e := s.db.QueryContext(ctx, "SELECT id,delegation_id,display_name,category,status,created_at FROM athletes "+where+" ORDER BY display_name LIMIT ? OFFSET ?", args...)
	if e != nil {
		return pagination.Result[domain.Athlete]{}, e
	}
	defer rows.Close()
	out := pagination.Result[domain.Athlete]{Limit: p.Limit, Offset: p.Offset}
	for rows.Next() {
		var a domain.Athlete
		var created string
		if e = rows.Scan(&a.ID, &a.DelegationID, &a.DisplayName, &a.Category, &a.Status, &created); e != nil {
			return out, e
		}
		out.Items = append(out.Items, a)
	}
	var total int
	countArgs := args[:len(args)-2]
	_ = s.db.QueryRowContext(ctx, "SELECT COUNT(*) FROM athletes "+where, countArgs...).Scan(&total)
	out.Total = total
	return out, rows.Err()
}
