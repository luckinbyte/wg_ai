package gate

import (
	"net"
	"testing"
	"time"

	"github.com/yourorg/wg_ai/internal/agent"
	"github.com/yourorg/wg_ai/internal/session"
)

func TestTCPServerStart(t *testing.T) {
	sessionMgr := session.NewManager()
	agentMgr := agent.NewManager(2, 10)
	defer agentMgr.Stop()

	srv := NewTCPServer(":0", sessionMgr, agentMgr)

	errCh := make(chan error, 1)
	go func() {
		errCh <- srv.Start()
	}()

	time.Sleep(100 * time.Millisecond)

	// Verify server is listening
	addr := srv.Addr()
	if addr == "" {
		t.Error("server address is empty")
	}

	// Try to connect
	conn, err := net.Dial("tcp", addr)
	if err != nil {
		t.Fatalf("failed to connect: %v", err)
	}
	conn.Close()

	srv.Stop()
}

func TestConnectionReadMessage(t *testing.T) {
	client, server := net.Pipe()
	defer client.Close()
	defer server.Close()

	// Send a test packet
	payload := []byte("test")
	packet := EncodePacket(MsgTypeRequest, payload)
	go client.Write(packet)

	conn := NewConnection(server)
	msgType, data, err := conn.ReadMessage()
	if err != nil {
		t.Fatalf("read failed: %v", err)
	}
	if msgType != MsgTypeRequest {
		t.Errorf("expected msgType %d, got %d", MsgTypeRequest, msgType)
	}
	if string(data) != "test" {
		t.Errorf("expected 'test', got %s", string(data))
	}
}
