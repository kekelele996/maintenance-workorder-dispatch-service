package worker

import (
	"context"
	"testing"
	"time"

	"workorder/internal/config"
	"workorder/internal/model"
	"workorder/internal/repository"
	"workorder/internal/service"
	"workorder/internal/store"
)

func setup() (*Scheduler, *service.Service, *repository.Repository) {
	st := store.New()
	repo := repository.New(st)
	cfg := config.Load()
	svc := service.New(repo, cfg)
	sch := New(repo, svc, 500*time.Millisecond)
	return sch, svc, repo
}

func TestTickRetriesAndExecutes(t *testing.T) {
	sch, svc, repo := setup()
	repo.CreateTechnician(&model.Technician{ID: "t1", Name: "t1", Skills: []string{"general"}, Available: true})
	o, _ := svc.CreateWorkOrder("FAULTY", "fault", model.PriorityHigh)
	svc.DispatchOrder(o.ID, "t1")
	svc.ExecuteOrder(o.ID) // -> failed

	retried, executed := sch.Tick(context.Background())
	if retried != 1 {
		t.Fatalf("retried=%d want 1", retried)
	}
	if executed != 1 {
		t.Fatalf("executed=%d want 1", executed)
	}

	failed, _ := svc.FindWorkOrder(o.ID)
	if failed.Status != model.StatusFailed {
		t.Fatalf("status=%q want failed", failed.Status)
	}
	if failed.Attempts < 1 {
		t.Fatalf("attempts=%d want >=1", failed.Attempts)
	}
}

func TestTickStopsWhenCancelled(t *testing.T) {
	st := store.New()
	repo := repository.New(st)
	svc := service.New(repo, config.Load())
	sch := New(repo, svc, 500*time.Millisecond)
	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	retried, executed := sch.Tick(ctx)
	if retried != 0 || executed != 0 {
		t.Fatalf("retried=%d executed=%d want 0/0", retried, executed)
	}
}
