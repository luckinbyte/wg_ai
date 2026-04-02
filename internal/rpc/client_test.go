package rpc

import (
	"testing"
)

func TestNewClient(t *testing.T) {
	cfg := &ClientConfig{
		DBAddr: "localhost:50052",
	}

	client := NewClient(cfg)
	if client == nil {
		t.Error("client should not be nil")
	}
}
