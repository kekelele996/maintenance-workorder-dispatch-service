package repository

import (
	"errors"
	"testing"

	"workorder/internal/model"
	"workorder/internal/store"
)

func newRepo() *Repository {
	return New(store.New())
}

func TestFindByIDWrapsNotFound(t *testing.T) {
	r := newRepo()
	_, err := r.FindByID("missing")
	if !errors.Is(err, ErrNotFound) {
		t.Fatalf("errors.Is(err, ErrNotFound)=false, err=%v", err)
	}
}

func TestFindTechnicianWrapsNotFound(t *testing.T) {
	r := newRepo()
	_, err := r.FindTechnician("missing")
	if !errors.Is(err, ErrTechNotFound) {
		t.Fatalf("errors.Is(err, ErrTechNotFound)=false, err=%v", err)
	}
}

func TestCreateDuplicate(t *testing.T) {
	r := newRepo()
	if _, err := r.Create(&model.WorkOrder{ID: "w1"}); err != nil {
		t.Fatal(err)
	}
	if _, err := r.Create(&model.WorkOrder{ID: "w1"}); !errors.Is(err, ErrAlreadyExists) {
		t.Fatalf("dup err=%v want ErrAlreadyExists", err)
	}
}

func TestFindByEquipment(t *testing.T) {
	r := newRepo()
	r.Create(&model.WorkOrder{ID: "w1", EquipmentID: "e1"})
	r.Create(&model.WorkOrder{ID: "w2", EquipmentID: "e2"})
	got, err := r.FindByEquipment("e1")
	if err != nil {
		t.Fatal(err)
	}
	if len(got) != 1 || got[0].ID != "w1" {
		t.Fatalf("got=%+v", got)
	}
}
