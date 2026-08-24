package service

import (
	"context"
	"database/sql"
	"fmt"
	"github.com/11DingKing/robot-athlete-village/internal/pagination"
)

type Report struct{ db *sql.DB }

func NewReport(db *sql.DB) *Report { return &Report{db: db} }

type Occupancy struct {
	RoomCode string `json:"room_code"`
	Occupied int    `json:"occupied"`
	Capacity int    `json:"capacity"`
}

func (r *Report) Occupancy(ctx context.Context, p pagination.Request) (pagination.Result[Occupancy], error) {
	rows, err := r.db.QueryContext(ctx, "SELECT code,occupied,capacity FROM rooms ORDER BY code LIMIT ? OFFSET ?", p.Limit, p.Offset)
	if err != nil {
		return pagination.Result[Occupancy]{}, err
	}
	defer rows.Close()
	out := pagination.Result[Occupancy]{Limit: p.Limit, Offset: p.Offset}
	for rows.Next() {
		var x Occupancy
		if err = rows.Scan(&x.RoomCode, &x.Occupied, &x.Capacity); err != nil {
			return out, err
		}
		out.Items = append(out.Items, x)
	}
	if err = rows.Err(); err != nil {
		return out, err
	}
	if err = r.db.QueryRowContext(ctx, "SELECT COUNT(*) FROM rooms").Scan(&out.Total); err != nil {
		return out, fmt.Errorf("count occupancy: %w", err)
	}
	return out, nil
}
