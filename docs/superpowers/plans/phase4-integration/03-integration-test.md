# Task 24: Integration Test - 集成测试

## 背景

编写端到端集成测试，验证完整的登录和消息流程。

## 步骤

### Step 1: Create integration test

Create `tests/integration/login_test.go`:

```go
//go:build integration
// +build integration

package integration

import (
	"net"
	"testing"
	"time"

	"github.com/yourorg/wg_ai/internal/agent"
	"github.com/yourorg/wg_ai/internal/gate"
	"github.com/yourorg/wg_ai/internal/session"
)

func TestConnectionAndHeartbeat(t *testing.T) {
	sessionMgr := session.NewManager()
	agentMgr := agent.NewManager(2, 10)
	defer agentMgr.Stop()

	srv := gate.NewTCPServer(":0", sessionMgr, agentMgr)
	
	go srv.Start()
	time.Sleep(100 * time.Millisecond)

	addr := srv.Addr()
	conn, err := net.Dial("tcp", addr)
	if err != nil {
		t.Fatalf("connect failed: %v", err)
	}
	defer conn.Close()

	// Send handshake
	handshake := gate.EncodePacket(gate.MsgTypeHandshake, []byte("test-token"))
	conn.Write(handshake)

	// Read response
	buf := make([]byte, 1024)
	conn.SetReadDeadline(time.Now().Add(2 * time.Second))
	n, err := conn.Read(buf)
	if err != nil {
		t.Fatalf("read response failed: %v", err)
	}
	if n == 0 {
		t.Error("no response received")
	}

	srv.Stop()
}
```

### Step 2: Run integration test

```bash
cd /root/ai_project/wg_ai
mkdir -p tests/integration
go test -tags=integration ./tests/integration/... -v
```

### Step 3: Commit

```bash
git add .
git commit -m "test: add integration test"
```

## 完成标志

- [ ] 集成测试文件存在
- [ ] 测试通过
