package store

import (
	"testing"

	"workorder/internal/model"
)

func TestPutGetOrder(t *testing.T) {
	s := New()
	o := &model.WorkOrder{ID: "w1", EquipmentID: "e1", Status: model.StatusPending}
	if err := s.PutOrder(o); err != nil {
		t.Fatal(err)
	}
	got, err := s.GetOrder("w1")
	if err != nil {
		t.Fatal(err)
	}
	if got.ID != "w1" || got.Status != model.StatusPending {
		t.Fatalf("got=%+v", got)
	}
	if _, err := s.GetOrder("missing"); err != ErrNotFound {
		t.Fatalf("err=%v want ErrNotFound", err)
	}
	if err := s.PutOrder(o); err != ErrAlreadyExists {
		t.Fatalf("dup err=%v want ErrAlreadyExists", err)
	}
}

func TestGetOrderReturnsCopy(t *testing.T) {
	s := New()
	if err := s.PutOrder(&model.WorkOrder{ID: "w1", EquipmentID: "e1"}); err != nil {
		t.Fatal(err)
	}
	got, _ := s.GetOrder("w1")
	got.EquipmentID = "MUTATED"
	again, _ := s.GetOrder("w1")
	if again.EquipmentID != "e1" {
		t.Fatalf("GetOrder returned internal reference: EquipmentID=%q", again.EquipmentID)
	}
}

func TestListOrdersReturnsCopies(t *testing.T) {
	s := New()
	for _, id := range []string{"b", "a", "c"} {
		if err := s.PutOrder(&model.WorkOrder{ID: id}); err != nil {
			t.Fatal(err)
		}
	}
	first := s.ListOrders()
	for i := range first {
		first[i].Title = "MUTATED"
	}
	again := s.ListOrders()
	for _, o := range again {
		if o.Title == "MUTATED" {
			t.Fatalf("ListOrders returned internal reference for %s", o.ID)
		}
	}
}

func TestUpdateOrder(t *testing.T) {
	s := New()
	if err := s.PutOrder(&model.WorkOrder{ID: "w1", Status: model.StatusPending}); err != nil {
		t.Fatal(err)
	}
	_, err := s.UpdateOrder("w1", func(o *model.WorkOrder) {
		o.Status = model.StatusAssigned
	})
	if err != nil {
		t.Fatal(err)
	}
	got, _ := s.GetOrder("w1")
	if got.Status != model.StatusAssigned {
		t.Fatalf("Status=%q want assigned", got.Status)
	}
}

func TestTechnicians(t *testing.T) {
	s := New()
	if err := s.PutTechnician(&model.Technician{ID: "t1", Name: "Tom", Skills: []string{"cnc-repair"}}); err != nil {
		t.Fatal(err)
	}
	got, err := s.GetTechnician("t1")
	if err != nil {
		t.Fatal(err)
	}
	if got.Name != "Tom" {
		t.Fatalf("Name=%q", got.Name)
	}
	got.Skills[0] = "MUTATED"
	again, _ := s.GetTechnician("t1")
	if again.Skills[0] != "cnc-repair" {
		t.Fatal("GetTechnician returned internal reference")
	}
	if got := len(s.TechnicianIDs()); got != 1 {
		t.Fatalf("TechnicianIDs len=%d want 1", got)
	}
}
