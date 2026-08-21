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
	if cfg.LogFile != "" {
		t.Errorf("LogFile = %q, want empty", cfg.LogFile)
	}
	if cfg.LogLevel != "info" {
		t.Errorf("LogLevel = %q, want info", cfg.LogLevel)
	}
	if cfg.LogMaxSize != 10 {
		t.Errorf("LogMaxSize = %d, want 10", cfg.LogMaxSize)
	}
	if cfg.LogMaxBackups != 5 {
		t.Errorf("LogMaxBackups = %d, want 5", cfg.LogMaxBackups)
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

func TestLoadLogEnvOverrides(t *testing.T) {
	os.Clearenv()
	os.Setenv("MCP_LOG_FILE", "/tmp/app.log")
	os.Setenv("MCP_LOG_LEVEL", "debug")
	os.Setenv("MCP_LOG_MAX_SIZE", "25")
	os.Setenv("MCP_LOG_MAX_BACKUPS", "7")
	cfg, err := Load(nil)
	if err != nil {
		t.Fatalf("Load: %v", err)
	}
	if cfg.LogFile != "/tmp/app.log" {
		t.Errorf("LogFile = %q, want /tmp/app.log", cfg.LogFile)
	}
	if cfg.LogLevel != "debug" {
		t.Errorf("LogLevel = %q, want debug", cfg.LogLevel)
	}
	if cfg.LogMaxSize != 25 {
		t.Errorf("LogMaxSize = %d, want 25", cfg.LogMaxSize)
	}
	if cfg.LogMaxBackups != 7 {
		t.Errorf("LogMaxBackups = %d, want 7", cfg.LogMaxBackups)
	}
}

func TestLoadLogFlagOverridesEnv(t *testing.T) {
	os.Clearenv()
	os.Setenv("MCP_LOG_LEVEL", "debug")
	os.Setenv("MCP_LOG_MAX_SIZE", "25")
	cfg, err := Load([]string{"-log-level", "warn", "-log-max-size", "50"})
	if err != nil {
		t.Fatalf("Load: %v", err)
	}
	if cfg.LogLevel != "warn" {
		t.Errorf("LogLevel = %q, want flag value to win over env", cfg.LogLevel)
	}
	if cfg.LogMaxSize != 50 {
		t.Errorf("LogMaxSize = %d, want flag value to win over env", cfg.LogMaxSize)
	}
}
