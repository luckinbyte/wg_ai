# Task 10-12: Agent Model - 玩家代理模型

## 背景与目标

实现 Agent 模型，这是 Skynet 架构的核心概念。每个 Agent 是一个独立的 goroutine，负责处理一组玩家的消息。通过 Agent 池实现负载均衡。

**为什么需要这个任务：**
- Agent 模型避免每个玩家一个 goroutine 的开销
- 消息队列保证同一玩家的消息顺序处理
- Agent 池提供负载均衡和资源控制

**输出：**
- `internal/agent/agent.go` - Agent 结构
- `internal/agent/manager.go` - Agent 管理器
- `internal/agent/dispatcher.go` - 消息分发器

## 依赖

- Task 9: Session 管理已实现

## 步骤

### Step 1: Write the failing test

Create `internal/agent/agent_test.go`:

```go
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
```

### Step 2: Run test to verify it fails

```bash
cd /root/ai_project/wg_ai
go test ./internal/agent/...
```

Expected: FAIL - undefined: New, NewManager, NewDispatcher

### Step 3: Write minimal implementation

Create `internal/agent/agent.go`:

```go
package agent

import (
	"sync"

	"github.com/yourorg/wg_ai/internal/session"
)

type Message struct {
	MsgID    uint16
	Sequence uint32
	Payload  []byte
	Sess     *session.Session
}

type Agent struct {
	ID         int
	players    map[int64]*session.Session
	msgQueue   chan *Message
	mutex      sync.RWMutex
	stopCh     chan struct{}
	dispatcher *Dispatcher
}

func New(id, queueSize int) *Agent {
	return &Agent{
		ID:         id,
		players:    make(map[int64]*session.Session),
		msgQueue:   make(chan *Message, queueSize),
		stopCh:     make(chan struct{}),
		dispatcher: NewDispatcher(),
	}
}

func (a *Agent) Push(msg *Message) {
	select {
	case a.msgQueue <- msg:
	default:
		// queue full, drop message
	}
}

func (a *Agent) Run() {
	for {
		select {
		case msg := <-a.msgQueue:
			a.handleMessage(msg)
		case <-a.stopCh:
			return
		}
	}
}

func (a *Agent) Stop() {
	close(a.stopCh)
}

func (a *Agent) handleMessage(msg *Message) {
	resp, err := a.dispatcher.Dispatch(a, msg)
	if err != nil {
		// TODO: send error response
		return
	}
	if resp != nil && msg.Sess != nil {
		msg.Sess.Send(resp)
	}
}

func (a *Agent) BindSession(sess *session.Session) {
	a.mutex.Lock()
	defer a.mutex.Unlock()
	a.players[sess.RID] = sess
}

func (a *Agent) UnbindSession(rid int64) {
	a.mutex.Lock()
	defer a.mutex.Unlock()
	delete(a.players, rid)
}

func (a *Agent) GetSession(rid int64) *session.Session {
	a.mutex.RLock()
	defer a.mutex.RUnlock()
	return a.players[rid]
}

func (a *Agent) RegisterHandler(msgID uint16, handler HandlerFunc) {
	a.dispatcher.Register(msgID, handler)
}
```

Create `internal/agent/manager.go`:

```go
package agent

import (
	"sync/atomic"
)

type Manager struct {
	agents     []*Agent
	roundRobin uint32
}

func NewManager(agentCount, queueSize int) *Manager {
	m := &Manager{
		agents: make([]*Agent, agentCount),
	}
	for i := 0; i < agentCount; i++ {
		agent := New(i, queueSize)
		m.agents[i] = agent
		go agent.Run()
	}
	return m
}

func (m *Manager) Assign() *Agent {
	idx := atomic.AddUint32(&m.roundRobin, 1) - 1
	return m.agents[idx%uint32(len(m.agents))]
}

func (m *Manager) Get(id int) *Agent {
	if id >= 0 && id < len(m.agents) {
		return m.agents[id]
	}
	return nil
}

func (m *Manager) Stop() {
	for _, a := range m.agents {
		a.Stop()
	}
}
```

Create `internal/agent/dispatcher.go`:

```go
package agent

type HandlerFunc func(a *Agent, msg *Message) ([]byte, error)

type Dispatcher struct {
	handlers map[uint16]HandlerFunc
}

func NewDispatcher() *Dispatcher {
	return &Dispatcher{
		handlers: make(map[uint16]HandlerFunc),
	}
}

func (d *Dispatcher) Register(msgID uint16, handler HandlerFunc) {
	d.handlers[msgID] = handler
}

func (d *Dispatcher) Get(msgID uint16) HandlerFunc {
	return d.handlers[msgID]
}

func (d *Dispatcher) Dispatch(a *Agent, msg *Message) ([]byte, error) {
	handler := d.handlers[msg.MsgID]
	if handler == nil {
		return nil, nil
	}
	return handler(a, msg)
}
```

### Step 4: Run test to verify it passes

```bash
cd /root/ai_project/wg_ai
go test ./internal/agent/...
```

Expected: PASS

### Step 5: Commit

```bash
git add .
git commit -m "feat: add agent model with manager and dispatcher"
```

## 验证

```bash
cd /root/ai_project/wg_ai
go test ./internal/agent/... -v
```

Expected: PASS

## 完成标志

- [ ] 测试通过
- [ ] agent.go 包含 Agent 结构和消息处理
- [ ] manager.go 包含 Agent 池管理
- [ ] dispatcher.go 包含消息路由
- [ ] Commit 完成
