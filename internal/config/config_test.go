package config

import "testing"

func TestLoadDefaults(t *testing.T) {
	t.Setenv("WORKORDER_WORKERS", "")
	t.Setenv("WORKORDER_RETRY_LIMIT", "")
	t.Setenv("WORKORDER_EXEC_TIMEOUT_MS", "")
	t.Setenv("WORKORDER_POLL_INTERVAL_MS", "")
	c := Load()
	if c.Workers != 4 {
		t.Fatalf("Workers=%d want 4", c.Workers)
	}
	if c.RetryLimit != 3 {
		t.Fatalf("RetryLimit=%d want 3", c.RetryLimit)
	}
	if len(c.SkillRoutes) == 0 {
		t.Fatal("SkillRoutes should be populated by defaults")
	}
}

func TestLoadEnv(t *testing.T) {
	t.Setenv("WORKORDER_WORKERS", "8")
	t.Setenv("WORKORDER_RETRY_LIMIT", "6")
	c := Load()
	if c.Workers != 8 {
		t.Fatalf("Workers=%d want 8", c.Workers)
	}
	if c.RetryLimit != 6 {
		t.Fatalf("RetryLimit=%d want 6", c.RetryLimit)
	}
}
