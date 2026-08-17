package util

import (
	"testing"

	"workorder/internal/model"
)

func mk(id string, status model.Status, p model.Priority) *model.WorkOrder {
	return &model.WorkOrder{ID: id, Status: status, Priority: p}
}

func TestFilterByStatusNoAliasing(t *testing.T) {
	orders := []*model.WorkOrder{
		mk("a", model.StatusPending, model.PriorityLow),
		mk("b", model.StatusCompleted, model.PriorityLow),
		mk("c", model.StatusPending, model.PriorityLow),
	}
	got := FilterByStatus(orders, model.StatusPending)
	if len(got) != 2 {
		t.Fatalf("len=%d want 2", len(got))
	}
	// 关键断言：过滤后入参切片内容必须保持不变（不能原地复用底层数组）。
	if orders[0].ID != "a" || orders[1].ID != "b" || orders[2].ID != "c" {
		t.Fatalf("FilterByStatus corrupted input order: %v %v %v", orders[0].ID, orders[1].ID, orders[2].ID)
	}
}

func TestFilterActive(t *testing.T) {
	orders := []*model.WorkOrder{
		mk("a", model.StatusPending, model.PriorityLow),
		mk("b", model.StatusCompleted, model.PriorityLow),
		mk("c", model.StatusRetrying, model.PriorityLow),
	}
	got := FilterActive(orders)
	if len(got) != 2 {
		t.Fatalf("len=%d want 2", len(got))
	}
	seen := map[string]bool{}
	for _, o := range got {
		seen[o.ID] = true
	}
	if !seen["a"] || !seen["c"] {
		t.Fatalf("got=%v", got)
	}
}

func TestSortByPriority(t *testing.T) {
	orders := []*model.WorkOrder{
		mk("a", model.StatusPending, model.PriorityLow),
		mk("b", model.StatusPending, model.PriorityUrgent),
		mk("c", model.StatusPending, model.PriorityHigh),
	}
	got := SortByPriority(orders)
	want := []string{"b", "c", "a"}
	for i := range want {
		if got[i].ID != want[i] {
			t.Fatalf("got[%d]=%s want %s", i, got[i].ID, want[i])
		}
	}
	if orders[0].ID != "a" {
		t.Fatal("SortByPriority mutated input order")
	}
}

func TestTopByPriority(t *testing.T) {
	orders := []*model.WorkOrder{
		mk("a", model.StatusPending, model.PriorityLow),
		mk("b", model.StatusPending, model.PriorityUrgent),
		mk("c", model.StatusPending, model.PriorityHigh),
	}
	got := TopByPriority(orders, 2)
	if len(got) != 2 || got[0].ID != "b" || got[1].ID != "c" {
		t.Fatalf("got=%v", got)
	}
}

func TestEquipCategory(t *testing.T) {
	cases := map[string]string{
		"CNC-01": "cnc",
		"BLR-02": "boiler",
		"CVY-03": "conveyor",
		"HVC-04": "hvac",
		"OTHER":  "default",
	}
	for in, want := range cases {
		if got := EquipCategory(in); got != want {
			t.Fatalf("EquipCategory(%q)=%q want %q", in, got, want)
		}
	}
}
