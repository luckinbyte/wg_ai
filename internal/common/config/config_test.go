package config

import (
	"os"
	"testing"
	"time"
)

func TestLoadGameConfig(t *testing.T) {
	content := `
server:
  id: 1
  name: "test-game"
  host: "0.0.0.0"
  port: 44445
  max_conn: 1000
gate:
  read_timeout: 30s
  write_timeout: 30s
`
	tmpFile, err := os.CreateTemp("", "config-*.yaml")
	if err != nil {
		t.Fatal(err)
	}
	defer os.Remove(tmpFile.Name())
	tmpFile.Write([]byte(content))
	tmpFile.Close()

	cfg, err := LoadGameConfig(tmpFile.Name())
	if err != nil {
		t.Fatalf("LoadGameConfig failed: %v", err)
	}

	if cfg.Server.ID != 1 {
		t.Errorf("expected Server.ID=1, got %d", cfg.Server.ID)
	}
	if cfg.Server.Port != 44445 {
		t.Errorf("expected Server.Port=44445, got %d", cfg.Server.Port)
	}
	if cfg.Gate.ReadTimeout != 30*time.Second {
		t.Errorf("expected ReadTimeout=30s, got %v", cfg.Gate.ReadTimeout)
	}
}
