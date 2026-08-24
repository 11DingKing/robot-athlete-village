package repository

import (
	"context"
	"database/sql"
	"fmt"
	"github.com/11DingKing/robot-athlete-village/internal/domain"
	"github.com/11DingKing/robot-athlete-village/internal/pagination"
	"strings"
)

type SlotView struct {
	ID               int64
	Venue, Sport     string
	StartsAt, EndsAt string
	Status           string
}

func (s *SQLite) ListSlots(ctx context.Context, p pagination.Request, sport string) (pagination.Result[SlotView], error) {
	args := []any{}
	where := "WHERE t.status='open'"
	if strings.TrimSpace(sport) != "" {
		where += " AND v.sport=?"
		args = append(args, sport)
	}
	args = append(args, p.Limit, p.Offset)
	rows, err := s.db.QueryContext(ctx, "SELECT t.id,v.name,v.sport,t.starts_at,t.ends_at,t.status FROM training_slots t JOIN venues v ON v.id=t.venue_id "+where+" ORDER BY t.starts_at LIMIT ? OFFSET ?", args...)
	if err != nil {
		return pagination.Result[SlotView]{}, err
	}
	defer rows.Close()
	out := pagination.Result[SlotView]{Limit: p.Limit, Offset: p.Offset}
	for rows.Next() {
		var x SlotView
		if err = rows.Scan(&x.ID, &x.Venue, &x.Sport, &x.StartsAt, &x.EndsAt, &x.Status); err != nil {
			return out, err
		}
		out.Items = append(out.Items, x)
	}
	if err = rows.Err(); err != nil {
		return out, err
	}
	countArgs := []any{}
	countWhere := "WHERE t.status='open'"
	if strings.TrimSpace(sport) != "" {
		countWhere += " AND v.sport=?"
		countArgs = append(countArgs, sport)
	}
	if err = s.db.QueryRowContext(ctx, "SELECT COUNT(*) FROM training_slots t JOIN venues v ON v.id=t.venue_id "+countWhere, countArgs...).Scan(&out.Total); err != nil {
		return out, fmt.Errorf("count slots: %w", err)
	}
	return out, nil
}
func (s *SQLite) FindAthlete(ctx context.Context, id int64) (domain.Athlete, error) {
	var a domain.Athlete
	var created string
	err := s.db.QueryRowContext(ctx, "SELECT id,delegation_id,display_name,category,status,created_at FROM athletes WHERE id=?", id).Scan(&a.ID, &a.DelegationID, &a.DisplayName, &a.Category, &a.Status, &created)
	if err == sql.ErrNoRows {
		return a, fmt.Errorf("athlete: %w", sql.ErrNoRows)
	}
	if err != nil {
		return a, err
	}
	return a, nil
}
func (s *SQLite) SetSlotStatus(ctx context.Context, id int64, from, to string) error {
	res, err := s.db.ExecContext(ctx, "UPDATE training_slots SET status=? WHERE id=? AND status=?", to, id, from)
	if err != nil {
		return err
	}
	n, _ := res.RowsAffected()
	if n == 0 {
		return fmt.Errorf("slot transition rejected")
	}
	return nil
}
