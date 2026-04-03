package health

import "testing"

func TestHealthCheck(t *testing.T) {
	h := NewChecker()
	h.SetReady(true)
	h.SetHealthy(true)

	if !h.IsReady() {
		t.Error("should be ready")
	}
	if !h.IsHealthy() {
		t.Error("should be healthy")
	}
}

func TestHealthCheckUnset(t *testing.T) {
	h := NewChecker()

	if h.IsReady() {
		t.Error("should not be ready initially")
	}
	if h.IsHealthy() {
		t.Error("should not be healthy initially")
	}
}

func TestHealthCheckToggle(t *testing.T) {
	h := NewChecker()

	h.SetReady(true)
	if !h.IsReady() {
		t.Error("should be ready")
	}

	h.SetReady(false)
	if h.IsReady() {
		t.Error("should not be ready after toggle")
	}
}
