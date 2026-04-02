# Task 26: Health Check - 健康检查

## 背景

实现健康检查组件，用于 Kubernetes 等编排系统。

## 步骤

### Step 1: Create health check

Create `internal/common/health/health.go`:

```go
package health

import (
	"sync"
)

type Checker struct {
	ready   bool
	healthy bool
	mutex   sync.RWMutex
}

func NewChecker() *Checker {
	return &Checker{}
}

func (c *Checker) IsReady() bool {
	c.mutex.RLock()
	defer c.mutex.RUnlock()
	return c.ready
}

func (c *Checker) SetReady(ready bool) {
	c.mutex.Lock()
	defer c.mutex.Unlock()
	c.ready = ready
}

func (c *Checker) IsHealthy() bool {
	c.mutex.RLock()
	defer c.mutex.RUnlock()
	return c.healthy
}

func (c *Checker) SetHealthy(healthy bool) {
	c.mutex.Lock()
	defer c.mutex.Unlock()
	c.healthy = healthy
}
```

### Step 2: Write test

Create `internal/common/health/health_test.go`:

```go
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
```

### Step 3: Test and commit

```bash
cd /root/ai_project/wg_ai
go test ./internal/common/health/...
git add .
git commit -m "feat: add health check component"
```

## 完成标志

- [ ] health.go 存在
- [ ] 测试通过
