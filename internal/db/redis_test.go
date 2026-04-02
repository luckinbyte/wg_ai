package db

import "testing"

func TestRedisConfigAddr(t *testing.T) {
	cfg := &RedisConfig{
		Host: "localhost",
		Port: 6379,
	}

	addr := cfg.Addr()
	if addr != "localhost:6379" {
		t.Errorf("expected localhost:6379, got %s", addr)
	}
}
