package repository

import (
	"context"
	"github.com/11DingKing/robot-athlete-village/internal/domain"
	"github.com/11DingKing/robot-athlete-village/internal/pagination"
	"testing"
)

func TestListAthletesFilters(t *testing.T) {
	r := newRepo(t)
	for _, f := range []domain.AthleteFilter{{}, {Category: "mobility"}, {Search: "虎"}, {Status: "missing"}} {
		out, e := r.ListAthletes(context.Background(), pagination.Normalize(10, 0), f)
		if e != nil {
			t.Fatal(e)
		}
		if f.Status == "missing" && out.Total != 0 {
			t.Fatalf("%+v", out)
		}
	}
}
