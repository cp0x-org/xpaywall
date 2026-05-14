package config

import (
	"strings"
	"testing"
)

func TestNewControlAPIConfigRequiresEnv(t *testing.T) {
	t.Setenv("CONTROL_DB_DSN", "")

	_, err := NewControlAPIConfig()
	if err == nil {
		t.Fatal("expected error")
	}

	if !strings.Contains(err.Error(), "CONTROL_DB_DSN") {
		t.Fatalf("error = %v, want mention of CONTROL_DB_DSN", err)
	}
}

func TestNewControlAPIConfigUsesEnv(t *testing.T) {
	t.Setenv("CONTROL_DB_DSN", "postgres://user:pass@localhost:5432/db")

	cfg, err := NewControlAPIConfig()
	if err != nil {
		t.Fatalf("NewControlAPIConfig() error = %v", err)
	}

	if cfg.DB_DSN != "postgres://user:pass@localhost:5432/db" {
		t.Fatalf("DBDSN = %q", cfg.DB_DSN)
	}
}
