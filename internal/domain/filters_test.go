package domain

import "testing"

func TestAthleteFilter(t *testing.T) {
	a := Athlete{DisplayName: "小虎机器人", Category: "mobility", Status: "ready"}
	cases := []struct {
		f  AthleteFilter
		ok bool
	}{{AthleteFilter{}, true}, {AthleteFilter{Category: "mobility"}, true}, {AthleteFilter{Status: "active"}, false}, {AthleteFilter{Search: "虎机"}, true}, {AthleteFilter{Search: "缺席"}, false}}
	for _, tc := range cases {
		if got := tc.f.Matches(a); got != tc.ok {
			t.Fatalf("%+v got %v", tc.f, got)
		}
	}
}
func TestAthleteFilterNormalize(t *testing.T) {
	f := AthleteFilter{Status: " READY ", Category: " MOBILITY ", Search: "  虎  "}.Normalize()
	if f.Status != "ready" || f.Category != "mobility" || f.Search != "虎" {
		t.Fatalf("%+v", f)
	}
	if !(AthleteFilter{}).Empty() {
		t.Fatal("empty")
	}
}
