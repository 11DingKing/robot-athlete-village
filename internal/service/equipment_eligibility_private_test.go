package service

import (
	"context"
	"github.com/11DingKing/robot-athlete-village/internal/domain"
	"github.com/11DingKing/robot-athlete-village/internal/repository"
	"testing"
)

type eligibilityBarrierStore struct {
	repository.Store
	checked chan struct{}
	release <-chan struct{}
}

func (s *eligibilityBarrierStore) FindAthlete(ctx context.Context, athleteID int64) (domain.Athlete, error) {
	athlete, err := s.Store.FindAthlete(ctx, athleteID)
	if err != nil {
		return athlete, err
	}
	s.checked <- struct{}{}
	select {
	case <-s.release:
		return athlete, nil
	case <-ctx.Done():
		return domain.Athlete{}, ctx.Err()
	}
}

func TestConcurrentAthleteWithdrawalPreventsEquipmentAssignment(t *testing.T) {
	_, base := villageService(t)
	release := make(chan struct{})
	store := &eligibilityBarrierStore{Store: base, checked: make(chan struct{}, 1), release: release}
	village := NewVillage(store, nil)
	coach := domain.User{ID: 2, Role: domain.RoleCoach}
	result := make(chan error, 1)

	go func() {
		_, err := village.Assign(context.Background(), coach, 1, 1)
		result <- err
	}()
	<-store.checked
	if _, err := base.DB().Exec("UPDATE athletes SET status='withdrawn' WHERE id=1"); err != nil {
		t.Fatal(err)
	}
	close(release)
	if err := <-result; err == nil {
		t.Fatal("equipment assignment succeeded after athlete eligibility was withdrawn")
	}

	var status string
	var assigned any
	if err := base.DB().QueryRow("SELECT status,assigned_athlete_id FROM equipment WHERE id=1").Scan(&status, &assigned); err != nil {
		t.Fatal(err)
	}
	if status != string(domain.EquipmentReady) || assigned != nil {
		t.Fatalf("withdrawn athlete retained equipment: status=%s assigned=%v", status, assigned)
	}
}
