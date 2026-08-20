package config

import (
	"os"
	"path/filepath"
	"testing"
)

func TestNewConfig(t *testing.T) {
	c := NewConfig()
	if c.LogLevel != "info" {
		t.Errorf("LogLevel = %q, want %q", c.LogLevel, "info")
	}
	if c.Token != "" {
		t.Errorf("Token = %q, want empty", c.Token)
	}
	if c.DataPath != "./data/" {
		t.Errorf("DataPath = %q, want %q", c.DataPath, "./data/")
	}
}

func TestLoad(t *testing.T) {
	path := filepath.Join(t.TempDir(), "config.json")
	content := `{"log_level":"debug","token":"test-token","data_path":"/tmp/data/"}`
	if err := os.WriteFile(path, []byte(content), 0644); err != nil {
		t.Fatal(err)
	}

	c := NewConfig()
	if err := c.Load(path); err != nil {
		t.Fatalf("Load() error = %v", err)
	}
	if c.LogLevel != "debug" {
		t.Errorf("LogLevel = %q, want %q", c.LogLevel, "debug")
	}
	if c.Token != "test-token" {
		t.Errorf("Token = %q, want %q", c.Token, "test-token")
	}
	if c.DataPath != "/tmp/data/" {
		t.Errorf("DataPath = %q, want %q", c.DataPath, "/tmp/data/")
	}
}

func TestLoadKeepsDefaultsForMissingFields(t *testing.T) {
	path := filepath.Join(t.TempDir(), "config.json")
	if err := os.WriteFile(path, []byte(`{"token":"only-token"}`), 0644); err != nil {
		t.Fatal(err)
	}

	c := NewConfig()
	if err := c.Load(path); err != nil {
		t.Fatalf("Load() error = %v", err)
	}
	if c.Token != "only-token" {
		t.Errorf("Token = %q, want %q", c.Token, "only-token")
	}
	if c.LogLevel != "info" {
		t.Errorf("LogLevel = %q, want default %q", c.LogLevel, "info")
	}
	if c.DataPath != "./data/" {
		t.Errorf("DataPath = %q, want default %q", c.DataPath, "./data/")
	}
}

func TestLoadFileNotFound(t *testing.T) {
	c := NewConfig()
	err := c.Load(filepath.Join(t.TempDir(), "missing.json"))
	if err == nil {
		t.Fatal("Load() error = nil, want error for missing file")
	}
}

func TestLoadInvalidJSON(t *testing.T) {
	path := filepath.Join(t.TempDir(), "config.json")
	if err := os.WriteFile(path, []byte(`{invalid`), 0644); err != nil {
		t.Fatal(err)
	}

	c := NewConfig()
	if err := c.Load(path); err == nil {
		t.Fatal("Load() error = nil, want error for invalid JSON")
	}
}
