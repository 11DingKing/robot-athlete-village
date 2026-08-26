package domain

import "strings"

type AthleteFilter struct{ Status, Category, Search string }

func (f AthleteFilter) Matches(a Athlete) bool {
	if f.Status != "" && a.Status != f.Status {
		return false
	}
	if f.Category != "" && a.Category != f.Category {
		return false
	}
	if f.Search != "" && !strings.Contains(strings.ToLower(a.DisplayName), strings.ToLower(f.Search)) {
		return false
	}
	return true
}
func (f AthleteFilter) Empty() bool { return f.Status == "" && f.Category == "" && f.Search == "" }
func (f AthleteFilter) Normalize() AthleteFilter {
	return AthleteFilter{Status: strings.ToLower(strings.TrimSpace(f.Status)), Category: strings.ToLower(strings.TrimSpace(f.Category)), Search: strings.TrimSpace(f.Search)}
}
