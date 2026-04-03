# Phase 2: 插件接口定义

> **Goal:** 定义主程序与插件之间的通信契约，包括 LogicModule、DataAccessor、LogicContext 等核心接口

---

## 2.1 定义核心接口

**Files:**
- Create: `plugin/interface.go`
- Create: `plugin/interface_test.go`

- [ ] **Step 1: Write the failing test**

```go
// plugin/interface_test.go
package plugin

import (
    "testing"
)

func TestLogicResult(t *testing.T) {
    // Test Success
    result := Success(map[string]any{"key": "value"})
    if result.Code != 0 {
        t.Errorf("expected code 0, got %d", result.Code)
    }
    if result.Data["key"] != "value" {
        t.Error("data mismatch")
    }
    
    // Test Error
    errResult := Error(100, "test error")
    if errResult.Code != 100 {
        t.Errorf("expected code 100, got %d", errResult.Code)
    }
    if errResult.Message != "test error" {
        t.Error("message mismatch")
    }
}

func TestLogicResultWithPush(t *testing.T) {
    result := Success(nil).
        WithPush(1001, []byte("data1")).
        WithPush(1002, []byte("data2"))
    
    if len(result.Push) != 2 {
        t.Errorf("expected 2 pushes, got %d", len(result.Push))
    }
    if result.Push[0].MsgID != 1001 {
        t.Error("push msg_id mismatch")
    }
}

func TestErrors(t *testing.T) {
    if ErrModuleNotFound.Error() != "module not found" {
        t.Error("error message mismatch")
    }
    if ErrMethodNotFound.Error() != "method not found" {
        t.Error("error message mismatch")
    }
}
```

- [ ] **Step 2: Run test to verify it fails**

Run: `go test ./plugin/... -v`
Expected: FAIL - package plugin not found

- [ ] **Step 3: Write minimal implementation**

```go
// plugin/interface.go
package plugin

import "errors"

// ============ 错误定义 ============

var (
    ErrModuleNotFound  = errors.New("module not found")
    ErrMethodNotFound  = errors.New("method not found")
    ErrInvalidModule   = errors.New("invalid module type")
    ErrPluginLoadFail  = errors.New("plugin load failed")
    ErrPlayerNotFound  = errors.New("player not found")
)

// ============ 数据访问接口 (供插件使用) ============

// DataAccessor 数据访问接口 - 简化版，供插件调用
type DataAccessor interface {
    // 获取基础字段 (返回值可直接类型断言修改)
    GetField(key string) (any, error)
    
    // 设置基础字段
    SetField(key string, value any) error
    
    // 获取数组字段 (返回指针，可直接修改)
    GetArray(key string) (any, error)
    
    // 标记脏数据
    MarkDirty()
}

// ============ 会话推送接口 ============

// SessionPush 会话推送接口
type SessionPush interface {
    // 推送消息给客户端
    Push(msgID uint16, data []byte) error
}

// ============ 逻辑上下文 ============

// LogicContext 逻辑处理上下文
type LogicContext struct {
    RID     int64        // 玩家ID
    UID     int64        // 用户ID
    Data    DataAccessor // 数据访问接口
    Session SessionPush  // 会话推送接口
}

// ============ 逻辑结果 ============

// LogicResult 逻辑处理结果
type LogicResult struct {
    Code    int            `json:"code"`    // 错误码，0表示成功
    Message string         `json:"message"` // 错误信息
    Data    map[string]any `json:"data"`    // 返回数据
    Push    []PushData     `json:"push"`    // 推送数据列表
}

// PushData 推送数据
type PushData struct {
    MsgID uint16 `json:"msg_id"` // 消息ID
    Data  []byte `json:"data"`   // 消息内容
}

// WithPush 添加推送数据 (链式调用)
func (r *LogicResult) WithPush(msgID uint16, data []byte) *LogicResult {
    r.Push = append(r.Push, PushData{MsgID: msgID, Data: data})
    return r
}

// ============ 逻辑模块接口 ============

// LogicModule 逻辑模块接口 - 插件必须实现
type LogicModule interface {
    // 模块名称
    Name() string
    
    // 处理请求
    Handle(ctx *LogicContext, method string, params map[string]any) (*LogicResult, error)
}

// ============ 辅助函数 ============

// Success 创建成功结果
func Success(data map[string]any) *LogicResult {
    return &LogicResult{
        Code:    0,
        Message: "success",
        Data:    data,
        Push:    nil,
    }
}

// Error 创建错误结果
func Error(code int, message string) *LogicResult {
    return &LogicResult{
        Code:    code,
        Message: message,
        Data:    nil,
        Push:    nil,
    }
}

// ErrorWithData 创建带数据的错误结果
func ErrorWithData(code int, message string, data map[string]any) *LogicResult {
    return &LogicResult{
        Code:    code,
        Message: message,
        Data:    data,
        Push:    nil,
    }
}
```

- [ ] **Step 4: Run test to verify it passes**

Run: `go test ./plugin/... -v`
Expected: PASS

- [ ] **Step 5: Commit**

```bash
git add plugin/interface.go plugin/interface_test.go
git commit -m "feat(plugin): add core interface definitions"
```

---

## 2.2 实现消息路由

**Files:**
- Create: `plugin/router.go`
- Create: `plugin/router_test.go`
- Create: `config/routes.json`

- [ ] **Step 1: Write the failing test**

```go
// plugin/router_test.go
package plugin

import (
    "testing"
)

func TestRouterRegister(t *testing.T) {
    r := NewRouter()
    r.Register(1001, "role", "login")
    
    cfg, ok := r.Get(1001)
    if !ok {
        t.Fatal("route not found")
    }
    if cfg.Module != "role" {
        t.Errorf("expected module 'role', got '%s'", cfg.Module)
    }
    if cfg.Method != "login" {
        t.Errorf("expected method 'login', got '%s'", cfg.Method)
    }
}

func TestRouterNotFound(t *testing.T) {
    r := NewRouter()
    _, ok := r.Get(9999)
    if ok {
        t.Error("expected not found")
    }
}

func TestRouterLoadFromMap(t *testing.T) {
    r := NewRouter()
    
    routes := []map[string]any{
        {"msg_id": float64(1001), "module": "role", "method": "login"},
        {"msg_id": float64(2001), "module": "item", "method": "use"},
    }
    
    if err := r.LoadFromMap(routes); err != nil {
        t.Fatal(err)
    }
    
    cfg, ok := r.Get(1001)
    if !ok || cfg.Module != "role" {
        t.Error("route 1001 mismatch")
    }
    
    cfg, ok = r.Get(2001)
    if !ok || cfg.Module != "item" {
        t.Error("route 2001 mismatch")
    }
}
```

- [ ] **Step 2: Run test to verify it fails**

Run: `go test ./plugin/... -v -run TestRouter`
Expected: FAIL

- [ ] **Step 3: Write minimal implementation**

```go
// plugin/router.go
package plugin

import (
    "encoding/json"
    "os"
)

// RouteConfig 路由配置
type RouteConfig struct {
    MsgID  uint16 `json:"msg_id"`
    Module string `json:"module"`
    Method string `json:"method"`
}

// Router 消息路由器
type Router struct {
    routes map[uint16]*RouteConfig
}

// NewRouter 创建路由器
func NewRouter() *Router {
    return &Router{
        routes: make(map[uint16]*RouteConfig),
    }
}

// Register 注册路由
func (r *Router) Register(msgID uint16, module, method string) {
    r.routes[msgID] = &RouteConfig{
        MsgID:  msgID,
        Module: module,
        Method: method,
    }
}

// Get 获取路由配置
func (r *Router) Get(msgID uint16) (*RouteConfig, bool) {
    cfg, ok := r.routes[msgID]
    return cfg, ok
}

// LoadFromMap 从 map 列表加载路由
func (r *Router) LoadFromMap(routes []map[string]any) error {
    for _, route := range routes {
        msgID, ok := route["msg_id"].(float64)
        if !ok {
            continue
        }
        
        module, _ := route["module"].(string)
        method, _ := route["method"].(string)
        
        r.Register(uint16(msgID), module, method)
    }
    return nil
}

// LoadFromConfig 从配置文件加载路由
func (r *Router) LoadFromConfig(path string) error {
    data, err := os.ReadFile(path)
    if err != nil {
        return err
    }
    
    var routes []map[string]any
    if err := json.Unmarshal(data, &routes); err != nil {
        return err
    }
    
    return r.LoadFromMap(routes)
}

// All 获取所有路由
func (r *Router) All() map[uint16]*RouteConfig {
    return r.routes
}
```

- [ ] **Step 4: Run test to verify it passes**

Run: `go test ./plugin/... -v -run TestRouter`
Expected: PASS

- [ ] **Step 5: Create routes config**

```json
// config/routes.json
[
    {"msg_id": 1001, "module": "role", "method": "login"},
    {"msg_id": 1002, "module": "role", "method": "heartbeat"},
    {"msg_id": 1003, "module": "role", "method": "get_info"},
    {"msg_id": 2001, "module": "item", "method": "list"},
    {"msg_id": 2002, "module": "item", "method": "use"},
    {"msg_id": 2003, "module": "item", "method": "add"}
]
```

- [ ] **Step 6: Commit**

```bash
git add plugin/router.go plugin/router_test.go config/routes.json
git commit -m "feat(plugin): add message router with config loading"
```

---

## Phase 2 Summary

| File | Description |
|------|-------------|
| `plugin/interface.go` | 核心接口：LogicModule, DataAccessor, LogicContext, LogicResult |
| `plugin/interface_test.go` | 单元测试 |
| `plugin/router.go` | 消息路由器 |
| `plugin/router_test.go` | 单元测试 |
| `config/routes.json` | 路由配置文件 |
