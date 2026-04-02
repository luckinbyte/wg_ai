# Task 27: Metrics - 指标收集

## 背景

实现基础指标收集，用于监控。

## 步骤

### Step 1: Create metrics

Create `internal/common/metrics/metrics.go`:

```go
package metrics

import (
	"sync/atomic"
)

type Metrics struct {
	connections int64
	messages    int64
	errors      int64
}

func NewMetrics() *Metrics {
	return &Metrics{}
}

func (m *Metrics) IncConnections() {
	atomic.AddInt64(&m.connections, 1)
}

func (m *Metrics) DecConnections() {
	atomic.AddInt64(&m.connections, -1)
}

func (m *Metrics) GetConnections() int64 {
	return atomic.LoadInt64(&m.connections)
}

func (m *Metrics) IncMessages() {
	atomic.AddInt64(&m.messages, 1)
}

func (m *Metrics) GetMessages() int64 {
	return atomic.LoadInt64(&m.messages)
}

func (m *Metrics) IncErrors() {
	atomic.AddInt64(&m.errors, 1)
}

func (m *Metrics) GetErrors() int64 {
	return atomic.LoadInt64(&m.errors)
}
```

### Step 2: Write test

Create `internal/common/metrics/metrics_test.go`:

```go
package metrics

import "testing"

func TestMetrics(t *testing.T) {
	m := NewMetrics()
	m.IncConnections()
	m.IncMessages()
	m.IncErrors()

	if m.GetConnections() != 1 {
		t.Error("connections should be 1")
	}
	if m.GetMessages() != 1 {
		t.Error("messages should be 1")
	}
	if m.GetErrors() != 1 {
		t.Error("errors should be 1")
	}
}
```

### Step 3: Test and commit

```bash
cd /root/ai_project/wg_ai
go test ./internal/common/metrics/...
git add .
git commit -m "feat: add basic metrics collection"
```

## 完成标志

- [ ] metrics.go 存在
- [ ] 测试通过
