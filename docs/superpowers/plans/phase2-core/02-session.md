# Task 9: Session Management - 会话管理

## 背景与目标

实现玩家会话管理，每个连接对应一个 Session，存储玩家的连接状态和基本信息。

**为什么需要这个任务：**
- 需要追踪每个在线玩家的连接状态
- Session 是 Agent 和 Gate 之间的桥梁
- 支持会话超时和清理

**输出：**
- `internal/session/session.go` - Session 结构
- `internal/session/manager.go` - Session 管理器

## 依赖

- Task 1: 项目结构已创建

## 步骤

### Step 1: Write the failing test

Create `internal/session/session_test.go`:

```go
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
```

### Step 2: Run test to verify it fails

```bash
cd /root/ai_project/wg_ai
go test ./internal/session/...
```

Expected: FAIL - undefined: New, NewManager

### Step 3: Write minimal implementation

Create `internal/session/session.go`:

```go
package session

import (
	"net"
	"sync"
	"time"
)

type Session struct {
	RID        int64
	UID        int64
	Conn       net.Conn
	CreatedAt  time.Time
	LastActive time.Time
	mutex      sync.RWMutex
}

func New(uid, rid int64, conn net.Conn) *Session {
	return &Session{
		UID:        uid,
		RID:        rid,
		Conn:       conn,
		CreatedAt:  time.Now(),
		LastActive: time.Now(),
	}
}

func (s *Session) Send(data []byte) error {
	s.mutex.Lock()
	defer s.mutex.Unlock()

	s.Conn.SetWriteDeadline(time.Now().Add(5 * time.Second))
	_, err := s.Conn.Write(data)
	return err
}

func (s *Session) Close() error {
	s.mutex.Lock()
	defer s.mutex.Unlock()
	if s.Conn != nil {
		return s.Conn.Close()
	}
	return nil
}

func (s *Session) UpdateActive() {
	s.mutex.Lock()
	defer s.mutex.Unlock()
	s.LastActive = time.Now()
}
```

Create `internal/session/manager.go`:

```go
package session

import (
	"net"
	"sync"
	"sync/atomic"
	"time"
)

type Manager struct {
	sessions map[int64]*Session
	mutex    sync.RWMutex
	ridSeq   int64
}

func NewManager() *Manager {
	return &Manager{
		sessions: make(map[int64]*Session),
		ridSeq:   time.Now().Unix(),
	}
}

func (m *Manager) Create(uid int64, conn net.Conn) *Session {
	m.mutex.Lock()
	defer m.mutex.Unlock()

	rid := atomic.AddInt64(&m.ridSeq, 1)
	sess := New(uid, rid, conn)
	m.sessions[rid] = sess
	return sess
}

func (m *Manager) Get(rid int64) *Session {
	m.mutex.RLock()
	defer m.mutex.RUnlock()
	return m.sessions[rid]
}

func (m *Manager) Remove(rid int64) {
	m.mutex.Lock()
	defer m.mutex.Unlock()
	delete(m.sessions, rid)
}

func (m *Manager) Count() int {
	m.mutex.RLock()
	defer m.mutex.RUnlock()
	return len(m.sessions)
}

func (m *Manager) GetAll() []*Session {
	m.mutex.RLock()
	defer m.mutex.RUnlock()
	result := make([]*Session, 0, len(m.sessions))
	for _, s := range m.sessions {
		result = append(result, s)
	}
	return result
}
```

### Step 4: Run test to verify it passes

```bash
cd /root/ai_project/wg_ai
go test ./internal/session/...
```

Expected: PASS

### Step 5: Commit

```bash
git add .
git commit -m "feat: add session management"
```

## 验证

```bash
cd /root/ai_project/wg_ai
go test ./internal/session/... -v
```

Expected: PASS

## 完成标志

- [ ] 测试通过
- [ ] session.go 包含 Session 结构和方法
- [ ] manager.go 包含 Manager 结构和 CRUD 方法
- [ ] Commit 完成
