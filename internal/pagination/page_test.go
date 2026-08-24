package pagination

import "testing"

func TestNormalizeDefaults(t *testing.T) {
	p := Normalize(0, -1)
	if p.Limit != 20 || p.Offset != 0 || p.Sort == "" {
		t.Fatalf("%+v", p)
	}
}
func TestNormalizeUpperBound(t *testing.T) {
	if p := Normalize(1000, 4); p.Limit != 20 || p.Offset != 4 {
		t.Fatalf("%+v", p)
	}
}
func TestNormalizeValid(t *testing.T) {
	if p := Normalize(50, 10); p.Limit != 50 || p.Offset != 10 {
		t.Fatalf("%+v", p)
	}
}
