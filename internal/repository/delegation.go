package repository

import (
	"context"
	"database/sql"
	"fmt"
	"github.com/11DingKing/robot-athlete-village/internal/domain"
	"github.com/11DingKing/robot-athlete-village/internal/pagination"
	"strings"
)

func (s *SQLite) ListDelegations(ctx context.Context, p pagination.Request, status string) (pagination.Result[domain.Delegation], error) {
	where := ""
	args := []any{}
	if strings.TrimSpace(status) != "" {
		where = "WHERE status=?"
		args = append(args, status)
	}
	args = append(args, p.Limit, p.Offset)
	rows, err := s.db.QueryContext(ctx, "SELECT id,name,country_code,status,created_at FROM delegations "+where+" ORDER BY name LIMIT ? OFFSET ?", args...)
	if err != nil {
		return pagination.Result[domain.Delegation]{}, err
	}
	defer rows.Close()
	out := pagination.Result[domain.Delegation]{Limit: p.Limit, Offset: p.Offset}
	for rows.Next() {
		var d domain.Delegation
		var created string
		if err = rows.Scan(&d.ID, &d.Name, &d.CountryCode, &d.Status, &created); err != nil {
			return out, err
		}
		out.Items = append(out.Items, d)
	}
	if err = rows.Err(); err != nil {
		return out, err
	}
	countArgs := []any{}
	if strings.TrimSpace(status) != "" {
		countArgs = append(countArgs, status)
	}
	if err = s.db.QueryRowContext(ctx, "SELECT COUNT(*) FROM delegations "+where, countArgs...).Scan(&out.Total); err != nil {
		return out, err
	}
	return out, nil
}
func (s *SQLite) UpdateDelegationStatus(ctx context.Context, id int64, from, to string) error {
	if from == to {
		return fmt.Errorf("status unchanged")
	}
	res, err := s.db.ExecContext(ctx, "UPDATE delegations SET status=? WHERE id=? AND status=?", to, id, from)
	if err != nil {
		return err
	}
	n, _ := res.RowsAffected()
	if n == 0 {
		return sql.ErrNoRows
	}
	return nil
}
func (s *SQLite) CountAthletes(ctx context.Context, delegationID int64) int {
	var n int
	_ = s.db.QueryRowContext(ctx, "SELECT COUNT(*) FROM athletes WHERE delegation_id=?", delegationID).Scan(&n)
	return n
}
func (s *SQLite) DelegationReady(ctx context.Context, id int64) bool {
	var status string
	if e := s.db.QueryRowContext(ctx, "SELECT status FROM delegations WHERE id=?", id).Scan(&status); e != nil {
		return false
	}
	return status == "approved"
}
