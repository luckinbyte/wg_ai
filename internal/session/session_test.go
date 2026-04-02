package session

import (
	"net"
	"testing"
)

func TestSessionCreate(t *testing.T) {
	_, server := net.Pipe()
	defer server.Close()

	sess := New(123, 456, server)
	if sess.UID != 123 {
		t.Errorf("expected UID 123, got %d", sess.UID)
	}
	if sess.RID != 456 {
		t.Errorf("expected RID 456, got %d", sess.RID)
	}
}

func TestSessionManager(t *testing.T) {
	mgr := NewManager()

	_, server := net.Pipe()
	defer server.Close()

	sess := mgr.Create(1, server)
	if sess.RID == 0 {
		t.Error("RID should not be 0")
	}
	if sess.UID != 1 {
		t.Errorf("expected UID 1, got %d", sess.UID)
	}

	// Test Get
	got := mgr.Get(sess.RID)
	if got == nil {
		t.Error("session not found")
	}

	// Test Remove
	mgr.Remove(sess.RID)
	got = mgr.Get(sess.RID)
	if got != nil {
		t.Error("session should be removed")
	}
}

func TestSessionManagerCount(t *testing.T) {
	mgr := NewManager()

	if mgr.Count() != 0 {
		t.Error("initial count should be 0")
	}

	_, s1 := net.Pipe()
	_, s2 := net.Pipe()
	defer s1.Close()
	defer s2.Close()

	mgr.Create(1, s1)
	mgr.Create(2, s2)

	if mgr.Count() != 2 {
		t.Errorf("expected count 2, got %d", mgr.Count())
	}
}
