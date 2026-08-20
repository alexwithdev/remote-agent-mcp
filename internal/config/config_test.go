package config

import (
	"os"
	"testing"
)

func TestLoadDefaults(t *testing.T) {
	os.Clearenv()
	cfg, err := Load(nil)
	if err != nil {
		t.Fatalf("Load: %v", err)
	}
	if cfg.Addr != "0.0.0.0:9090" {
		t.Errorf("Addr = %q, want 0.0.0.0:9090", cfg.Addr)
	}
	if cfg.Root == "" {
		t.Error("Root should default to the working directory, not empty")
	}
	if cfg.Token != "" {
		t.Errorf("Token = %q, want empty", cfg.Token)
	}
	if cfg.AllowAll || cfg.AllowUnconfirmed {
		t.Errorf("allow flags should default to false")
	}
}

func TestLoadEnvOverrides(t *testing.T) {
	os.Clearenv()
	os.Setenv("MCP_ADDR", "127.0.0.1:8080")
	os.Setenv("MCP_TOKEN", "secret")
	os.Setenv("MCP_ALLOW_ALL", "true")
	os.Setenv("MCP_ALLOW_UNCONFIRMED", "1")
	cfg, err := Load(nil)
	if err != nil {
		t.Fatalf("Load: %v", err)
	}
	if cfg.Addr != "127.0.0.1:8080" {
		t.Errorf("Addr = %q", cfg.Addr)
	}
	if cfg.Token != "secret" {
		t.Errorf("Token = %q", cfg.Token)
	}
	if !cfg.AllowAll {
		t.Error("AllowAll should be true")
	}
	if !cfg.AllowUnconfirmed {
		t.Error("AllowUnconfirmed should be true")
	}
}

func TestLoadFlagOverridesEnv(t *testing.T) {
	os.Clearenv()
	os.Setenv("MCP_ADDR", "127.0.0.1:8080")
	cfg, err := Load([]string{"-addr", "0.0.0.0:9999"})
	if err != nil {
		t.Fatalf("Load: %v", err)
	}
	if cfg.Addr != "0.0.0.0:9999" {
		t.Errorf("Addr = %q, want flag value to win over env", cfg.Addr)
	}
}
