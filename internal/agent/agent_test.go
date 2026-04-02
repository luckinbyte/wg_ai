package agent

import (
	"testing"
)

func TestAgentNew(t *testing.T) {
	a := New(1, 100)
	if a.ID != 1 {
		t.Errorf("expected ID 1, got %d", a.ID)
	}
	if cap(a.msgQueue) != 100 {
		t.Errorf("expected queue size 100, got %d", cap(a.msgQueue))
	}
}

func TestAgentManager(t *testing.T) {
	mgr := NewManager(3, 10)

	// Test round-robin assignment
	a1 := mgr.Assign()
	a2 := mgr.Assign()
	a3 := mgr.Assign()

	if a1 == nil || a2 == nil || a3 == nil {
		t.Fatal("assigned agent should not be nil")
	}

	// Test Get
	got := mgr.Get(a1.ID)
	if got != a1 {
		t.Error("Get should return same agent")
	}

	mgr.Stop()
}

func TestDispatcher(t *testing.T) {
	d := NewDispatcher()

	called := false
	d.Register(1001, func(a *Agent, msg *Message) ([]byte, error) {
		called = true
		return []byte("ok"), nil
	})

	handler := d.Get(1001)
	if handler == nil {
		t.Fatal("handler not registered")
	}

	_, _ = handler(nil, &Message{MsgID: 1001})
	if !called {
		t.Error("handler was not called")
	}
}
