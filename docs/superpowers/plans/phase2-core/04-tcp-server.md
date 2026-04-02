# Task 13-14: TCP Server - TCP 网关服务器

## 背景与目标

实现 TCP 网关服务器，接受客户端连接，处理握手，分发消息到 Agent 池。

**为什么需要这个任务：**
- TCP 是游戏客户端的主要通信方式
- Gate 负责连接管理和协议解析
- 需要集成 Session 和 Agent 组件

**输出：**
- `internal/gate/tcp_server.go` - TCP 服务器
- `internal/gate/connection.go` - 连接封装

## 依赖

- Task 8: Protocol 编解码
- Task 9: Session 管理
- Task 10-12: Agent 模型

## 步骤

### Step 1: Write the failing test

Create `internal/gate/tcp_server_test.go`:

```go
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
```

### Step 2: Run test to verify it fails

```bash
cd /root/ai_project/wg_ai
go test ./internal/gate/...
```

Expected: FAIL - undefined: NewTCPServer, NewConnection

### Step 3: Write implementation

Create `internal/gate/connection.go`:

```go
package gate

import (
	"net"
	"time"
)

type Connection struct {
	conn       net.Conn
	remoteAddr string
}

func NewConnection(conn net.Conn) *Connection {
	return &Connection{
		conn:       conn,
		remoteAddr: conn.RemoteAddr().String(),
	}
}

func (c *Connection) ReadMessage() (msgType byte, payload []byte, err error) {
	return DecodePacket(c.conn)
}

func (c *Connection) WriteMessage(msgType byte, payload []byte) error {
	packet := EncodePacket(msgType, payload)
	c.conn.SetWriteDeadline(time.Now().Add(5 * time.Second))
	_, err := c.conn.Write(packet)
	return err
}

func (c *Connection) Close() error {
	return c.conn.Close()
}

func (c *Connection) RemoteAddr() string {
	return c.remoteAddr
}

func (c *Connection) SetReadTimeout(d time.Duration) error {
	return c.conn.SetReadDeadline(time.Now().Add(d))
}
```

Create `internal/gate/tcp_server.go`:

```go
package gate

import (
	"net"
	"sync"
	"time"

	"github.com/yourorg/wg_ai/internal/agent"
	"github.com/yourorg/wg_ai/internal/session"
)

type TCPServer struct {
	addr       string
	listener   net.Listener
	stopCh     chan struct{}
	wg         sync.WaitGroup
	sessionMgr *session.Manager
	agentMgr   *agent.Manager
}

func NewTCPServer(addr string, sessionMgr *session.Manager, agentMgr *agent.Manager) *TCPServer {
	return &TCPServer{
		addr:       addr,
		stopCh:     make(chan struct{}),
		sessionMgr: sessionMgr,
		agentMgr:   agentMgr,
	}
}

func (s *TCPServer) Start() error {
	ln, err := net.Listen("tcp", s.addr)
	if err != nil {
		return err
	}
	s.listener = ln

	for {
		select {
		case <-s.stopCh:
			return nil
		default:
			conn, err := ln.Accept()
			if err != nil {
				continue
			}
			s.wg.Add(1)
			go s.handleConnection(conn)
		}
	}
}

func (s *TCPServer) Stop() {
	close(s.stopCh)
	if s.listener != nil {
		s.listener.Close()
	}
	s.wg.Wait()
}

func (s *TCPServer) Addr() string {
	if s.listener != nil {
		return s.listener.Addr().String()
	}
	return ""
}

func (s *TCPServer) handleConnection(conn net.Conn) {
	defer s.wg.Done()
	defer conn.Close()

	c := NewConnection(conn)
	conn.SetDeadline(time.Now().Add(30 * time.Second))

	// Read handshake
	msgType, data, err := c.ReadMessage()
	if err != nil {
		return
	}
	if msgType != MsgTypeHandshake {
		return
	}

	// TODO: Validate token via login service
	// For now, create session with UID from token

	// Send handshake response
	resp := []byte{0, 0, 0, 0} // code = 0 (success)
	c.WriteMessage(MsgTypeResponse, resp)

	// Message loop
	for {
		conn.SetReadDeadline(time.Now().Add(30 * time.Second))
		msgType, data, err := c.ReadMessage()
		if err != nil {
			break
		}

		if msgType == MsgTypeRequest {
			ag := s.agentMgr.Assign()
			ag.Push(&agent.Message{
				MsgID:   uint16(data[0])<<8 | uint16(data[1]),
				Payload: data[2:],
			})
		}
	}
}
```

### Step 4: Run test to verify it passes

```bash
cd /root/ai_project/wg_ai
go test ./internal/gate/...
```

Expected: PASS

### Step 5: Commit

```bash
git add .
git commit -m "feat: add TCP server with connection handling"
```

## 验证

```bash
cd /root/ai_project/wg_ai
go test ./internal/gate/... -v
```

Expected: PASS

## 完成标志

- [ ] 测试通过
- [ ] tcp_server.go 包含 TCPServer 结构
- [ ] connection.go 包含 Connection 封装
- [ ] 集成 Session 和 Agent
- [ ] Commit 完成
