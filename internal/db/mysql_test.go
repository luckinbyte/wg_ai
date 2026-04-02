package db

import (
	"testing"
)

func TestMySQLConfigDSN(t *testing.T) {
	cfg := &MySQLConfig{
		Host:     "localhost",
		Port:     3306,
		Database: "game",
		Username: "root",
		Password: "secret",
	}

	dsn := cfg.DSN()
	expected := "root:secret@tcp(localhost:3306)/game?charset=utf8mb4&parseTime=True"
	if dsn != expected {
		t.Errorf("DSN mismatch:\ngot:      %s\nexpected: %s", dsn, expected)
	}
}
