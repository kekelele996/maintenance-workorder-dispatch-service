package service

import (
	"errors"
	"fmt"
	"sync"
	"testing"

	"workorder/internal/config"
	"workorder/internal/model"
	"workorder/internal/repository"
	"workorder/internal/store"
)

func newService() *Service {
	return New(repository.New(store.New()), config.Load())
}

func seedTech(s *Service, id string, skills []string, available bool) {
	s.repo.CreateTechnician(&model.Technician{ID: id, Name: id, Skills: skills, Available: available})
}

func TestCreateAndFind(t *testing.T) {
	s := newService()
	o, err := s.CreateWorkOrder("CNC-01", "主轴异响", model.PriorityHigh)
	if err != nil {
		t.Fatal(err)
	}
	if o.Status != model.StatusPending {
		t.Fatalf("status=%q want pending", o.Status)
	}
	got, err := s.FindWorkOrder(o.ID)
	if err != nil {
		t.Fatal(err)
	}
	if got.EquipmentID != "CNC-01" {
		t.Fatalf("equipment=%q", got.EquipmentID)
	}
}

func TestFindMissingWrapsOrderNotFound(t *testing.T) {
	s := newService()
	_, err := s.FindWorkOrder("missing")
	if !errors.Is(err, ErrOrderNotFound) {
		t.Fatalf("errors.Is(err, ErrOrderNotFound)=false, err=%v", err)
	}
}

func TestDispatchMatchesSkill(t *testing.T) {
	s := newService()
	o, _ := s.CreateWorkOrder("CNC-01", "主轴异响", model.PriorityHigh)
	seedTech(s, "t1", []string{"cnc-repair"}, true)
	got, err := s.DispatchOrder(o.ID, "t1")
	if err != nil {
		t.Fatal(err)
	}
	if got.Status != model.StatusAssigned || got.TechnicianID != "t1" {
		t.Fatalf("got=%+v", got)
	}
}

func TestDispatchRejectsWrongSkill(t *testing.T) {
	s := newService()
	o, _ := s.CreateWorkOrder("CNC-01", "主轴异响", model.PriorityHigh)
	seedTech(s, "t1", []string{"boiler-cert"}, true)
	if _, err := s.DispatchOrder(o.ID, "t1"); !errors.Is(err, ErrNoAvailableTech) {
		t.Fatalf("err=%v want ErrNoAvailableTech", err)
	}
}

func TestExecuteLifecycle(t *testing.T) {
	s := newService()
	o, _ := s.CreateWorkOrder("CVY-01", "传送带检修", model.PriorityMedium)
	seedTech(s, "t1", []string{"conveyor-maintenance"}, true)
	s.DispatchOrder(o.ID, "t1")
	// first execute: assigned -> in_progress
	got, err := s.ExecuteOrder(o.ID)
	if err != nil {
		t.Fatal(err)
	}
	if got.Status != model.StatusCompleted {
		t.Fatalf("status=%q want completed", got.Status)
	}
}

func TestRetryLifecycle(t *testing.T) {
	s := newService()
	o, _ := s.CreateWorkOrder("FAULTY", "故障设备", model.PriorityHigh)
	seedTech(s, "t1", []string{"general"}, true)
	s.DispatchOrder(o.ID, "t1")
	if _, err := s.ExecuteOrder(o.ID); err != nil {
		t.Fatal(err)
	}
	failed, _ := s.FindWorkOrder(o.ID)
	if failed.Status != model.StatusFailed {
		t.Fatalf("status=%q want failed", failed.Status)
	}
	if _, err := s.RetryOrder(o.ID); err != nil {
		t.Fatal(err)
	}
	retried, _ := s.FindWorkOrder(o.ID)
	if retried.Status != model.StatusRetrying {
		t.Fatalf("status=%q want retrying", retried.Status)
	}
}

func TestActiveCount(t *testing.T) {
	s := newService()
	o1, _ := s.CreateWorkOrder("CNC-01", "a", model.PriorityHigh)
	o2, _ := s.CreateWorkOrder("CNC-02", "b", model.PriorityHigh)
	o3, _ := s.CreateWorkOrder("CNC-03", "c", model.PriorityHigh)
	seedTech(s, "t1", []string{"cnc-repair"}, true)
	s.DispatchOrder(o1.ID, "t1")
	s.ExecuteOrder(o2.ID) // pending -> invalid transition, stays pending
	_ = o3
	if got := s.ActiveCount(); got != 3 {
		t.Fatalf("ActiveCount=%d want 3", got)
	}
}

func TestListWorkOrdersByStatus(t *testing.T) {
	s := newService()
	o1, _ := s.CreateWorkOrder("CNC-01", "a", model.PriorityLow)
	o2, _ := s.CreateWorkOrder("CNC-02", "b", model.PriorityUrgent)
	status := model.StatusPending
	got, err := s.ListWorkOrders(&status)
	if err != nil {
		t.Fatal(err)
	}
	if len(got) != 2 {
		t.Fatalf("len=%d want 2", len(got))
	}
	if got[0].ID != o2.ID || got[1].ID != o1.ID {
		t.Fatalf("order wrong: %v", got)
	}
}

func TestConcurrentDispatchAndRead(t *testing.T) {
	s := newService()
	for i := 0; i < 120; i++ {
		if _, err := s.CreateWorkOrder(fmt.Sprintf("CNC-%03d", i), "job", model.PriorityHigh); err != nil {
			t.Fatal(err)
		}
	}
	seedTech(s, "t1", []string{"cnc-repair"}, true)

	var wg sync.WaitGroup
	wg.Add(1)
	go func() {
		defer wg.Done()
		orders, _ := s.ListWorkOrders(nil)
		for _, o := range orders {
			if _, err := s.DispatchOrder(o.ID, "t1"); err == nil {
				_, _ = s.ExecuteOrder(o.ID)
			}
		}
	}()
	for i := 0; i < 4; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			for j := 0; j < 60; j++ {
				_ = s.ActiveCount()
			}
		}()
	}
	wg.Wait()
}

func TestCustomSkillRoutes(t *testing.T) {
	t.Setenv("WORKORDER_SKILL_ROUTES", "cnc:special,boiler:boiler-cert")
	st := store.New()
	repo := repository.New(st)
	svc := New(repo, config.Load())

	o, err := svc.CreateWorkOrder("CNC-01", "主轴异响", model.PriorityHigh)
	if err != nil {
		t.Fatal(err)
	}
	repo.CreateTechnician(&model.Technician{ID: "t1", Name: "t1", Skills: []string{"special"}, Available: true})
	got, err := svc.DispatchOrder(o.ID, "t1")
	if err != nil {
		t.Fatal(err)
	}
	if got.TechnicianID != "t1" || got.Status != model.StatusAssigned {
		t.Fatalf("got=%+v", got)
	}
}

func TestSkillRouteFallback(t *testing.T) {
	t.Setenv("WORKORDER_SKILL_ROUTES", "invalid")
	st := store.New()
	repo := repository.New(st)
	svc := New(repo, config.Load())

	o, err := svc.CreateWorkOrder("CNC-01", "主轴异响", model.PriorityHigh)
	if err != nil {
		t.Fatal(err)
	}
	repo.CreateTechnician(&model.Technician{ID: "t1", Name: "t1", Skills: []string{"cnc-repair"}, Available: true})
	got, err := svc.DispatchOrder(o.ID, "t1")
	if err != nil {
		t.Fatal(err)
	}
	if got.TechnicianID != "t1" || got.Status != model.StatusAssigned {
		t.Fatalf("got=%+v", got)
	}
}

func TestStats(t *testing.T) {
	s := newService()
	o1, _ := s.CreateWorkOrder("CNC-01", "a", model.PriorityHigh)
	s.CreateWorkOrder("CNC-02", "b", model.PriorityHigh)
	s.CreateWorkOrder("CNC-03", "c", model.PriorityHigh)
	seedTech(s, "t1", []string{"cnc-repair"}, true)
	s.DispatchOrder(o1.ID, "t1")

	stats := s.Stats()
	if stats[model.StatusPending] != 2 {
		t.Fatalf("pending=%d want 2", stats[model.StatusPending])
	}
	if stats[model.StatusAssigned] != 1 {
		t.Fatalf("assigned=%d want 1", stats[model.StatusAssigned])
	}
}
