package model

import "testing"

func TestCanTransition(t *testing.T) {
	cases := []struct {
		from, to Status
		want     bool
	}{
		{StatusPending, StatusAssigned, true},
		{StatusPending, StatusInProgress, false},
		{StatusAssigned, StatusInProgress, true},
		{StatusInProgress, StatusCompleted, true},
		{StatusInProgress, StatusFailed, true},
		{StatusFailed, StatusRetrying, true},
		{StatusRetrying, StatusInProgress, true},
		{StatusCompleted, StatusRetrying, false},
	}
	for _, c := range cases {
		if got := CanTransition(c.from, c.to); got != c.want {
			t.Fatalf("CanTransition(%s,%s)=%v want %v", c.from, c.to, got, c.want)
		}
	}
}

func TestActiveStatusesIncludesRetrying(t *testing.T) {
	if !ActiveStatuses[StatusRetrying] {
		t.Fatal("ActiveStatuses must include retrying")
	}
	if ActiveStatuses[StatusCompleted] {
		t.Fatal("ActiveStatuses must not include completed")
	}
}

func TestHasSkill(t *testing.T) {
	tech := &Technician{Skills: []string{"cnc-repair", "hvac-cert"}}
	if !tech.HasSkill("cnc-repair") {
		t.Fatal("expected HasSkill(cnc-repair) true")
	}
	if tech.HasSkill("boiler-cert") {
		t.Fatal("expected HasSkill(boiler-cert) false")
	}
}
