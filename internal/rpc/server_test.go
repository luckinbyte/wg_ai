package rpc

import (
	"testing"
)

func TestNewServer(t *testing.T) {
	srv := NewServer(":50052")
	if srv == nil {
		t.Error("server should not be nil")
	}
}
