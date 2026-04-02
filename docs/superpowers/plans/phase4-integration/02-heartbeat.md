# Task 23: Heartbeat Handler - 心跳处理

## 背景

实现心跳消息处理，保持连接活跃。

## 步骤

### Step 1: Create handlers file

Create `internal/agent/handlers.go`:

```go
package agent

import (
	"google.golang.org/protobuf/proto"

	cs "github.com/yourorg/wg_ai/proto/cs"
)

func RegisterDefaultHandlers(a *Agent) {
	a.RegisterHandler(1002, handleHeartbeat) // MSG_ID_HEARTBEAT
}

func handleHeartbeat(a *Agent, msg *Message) ([]byte, error) {
	resp := &cs.HeartbeatResponse{}
	return proto.Marshal(resp)
}
```

### Step 2: Update Agent creation

在 `internal/agent/agent.go` 的 New 函数中注册默认处理器：

```go
func New(id, queueSize int) *Agent {
	a := &Agent{
		ID:         id,
		players:    make(map[int64]*session.Session),
		msgQueue:   make(chan *Message, queueSize),
		stopCh:     make(chan struct{}),
		dispatcher: NewDispatcher(),
	}
	RegisterDefaultHandlers(a)
	return a
}
```

### Step 3: Test

```bash
cd /root/ai_project/wg_ai
go test ./internal/agent/...
```

### Step 4: Commit

```bash
git add .
git commit -m "feat: add heartbeat handler"
```

## 完成标志

- [ ] handlers.go 存在
- [ ] 心跳处理器注册
- [ ] 测试通过
