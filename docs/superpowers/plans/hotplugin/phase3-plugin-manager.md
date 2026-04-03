# Phase 3: 插件管理器

> **Goal:** 实现插件加载、热更、调用机制，以及 DataAccessor 适配器

---

## 3.1 实现 PluginManager 核心

**Files:**
- Create: `internal/plugin/manager.go`
- Create: `internal/plugin/manager_test.go`

- [ ] **Step 1: Write the failing test**

```go
// internal/plugin/manager_test.go
package plugin

import (
    "os"
    "path/filepath"
    "testing"
    
    p "github.com/yourorg/wg_ai/plugin"
)

// MockLogicModule 模拟逻辑模块
type MockLogicModule struct {
    name string
}

func (m *MockLogicModule) Name() string {
    return m.name
}

func (m *MockLogicModule) Handle(ctx *p.LogicContext, method string, params map[string]any) (*p.LogicResult, error) {
    return p.Success(map[string]any{"method": method}), nil
}

func TestNewManager(t *testing.T) {
    mgr := NewManager()
    if mgr == nil {
        t.Fatal("expected manager")
    }
    if mgr.modules == nil {
        t.Error("expected modules map")
    }
}

func TestManagerRegisterModule(t *testing.T) {
    mgr := NewManager()
    
    // 直接注册模块 (不通过 .so 文件)
    mgr.RegisterModule("mock", &MockLogicModule{name: "mock"})
    
    module := mgr.GetModule("mock")
    if module == nil {
        t.Fatal("module not found")
    }
    if module.Name() != "mock" {
        t.Error("name mismatch")
    }
}

func TestManagerGetModuleNotFound(t *testing.T) {
    mgr := NewManager()
    
    module := mgr.GetModule("notexist")
    if module != nil {
        t.Error("expected nil for non-existent module")
    }
}

func TestManagerListModules(t *testing.T) {
    mgr := NewManager()
    mgr.RegisterModule("role", &MockLogicModule{name: "role"})
    mgr.RegisterModule("item", &MockLogicModule{name: "item"})
    
    list := mgr.ListModules()
    if len(list) != 2 {
        t.Errorf("expected 2 modules, got %d", len(list))
    }
}
```

- [ ] **Step 2: Run test to verify it fails**

Run: `go test ./internal/plugin/... -v`
Expected: FAIL - package plugin not found

- [ ] **Step 3: Write minimal implementation**

```go
// internal/plugin/manager.go
package plugin

import (
    "plugin"
    "sync"
    
    p "github.com/yourorg/wg_ai/plugin"
)

// Manager 插件管理器
type Manager struct {
    plugins map[string]*plugin.Plugin // module -> go plugin (.so)
    modules map[string]p.LogicModule  // module -> LogicModule
    router  *p.Router                 // 消息路由
    mutex   sync.RWMutex
}

// NewManager 创建插件管理器
func NewManager() *Manager {
    return &Manager{
        plugins: make(map[string]*plugin.Plugin),
        modules: make(map[string]p.LogicModule),
        router:  p.NewRouter(),
    }
}

// RegisterModule 直接注册模块 (用于测试或不使用 .so 的场景)
func (m *Manager) RegisterModule(name string, module p.LogicModule) {
    m.mutex.Lock()
    defer m.mutex.Unlock()
    m.modules[name] = module
}

// LoadPlugin 加载 .so 插件
func (m *Manager) LoadPlugin(moduleName, path string) error {
    // 1. 打开 .so 文件
    p, err := plugin.Open(path)
    if err != nil {
        return err
    }
    
    // 2. 查找导出符号 (如 RoleModule, ItemModule)
    symName := moduleName + "Module"
    sym, err := p.Lookup(symName)
    if err != nil {
        return err
    }
    
    // 3. 类型断言为 LogicModule
    logicModule, ok := sym.(p.LogicModule)
    if !ok {
        return p.ErrInvalidModule
    }
    
    // 4. 注册模块
    m.mutex.Lock()
    defer m.mutex.Unlock()
    m.plugins[moduleName] = p
    m.modules[moduleName] = logicModule
    
    return nil
}

// HotReload 热更新插件 (Go plugin 无法卸载，直接加载新版本覆盖)
func (m *Manager) HotReload(moduleName, newPath string) error {
    return m.LoadPlugin(moduleName, newPath)
}

// GetModule 获取模块
func (m *Manager) GetModule(name string) p.LogicModule {
    m.mutex.RLock()
    defer m.mutex.RUnlock()
    return m.modules[name]
}

// ListModules 列出所有已加载模块
func (m *Manager) ListModules() []string {
    m.mutex.RLock()
    defer m.mutex.RUnlock()
    
    list := make([]string, 0, len(m.modules))
    for name := range m.modules {
        list = append(list, name)
    }
    return list
}

// Router 获取路由器
func (m *Manager) Router() *p.Router {
    return m.router
}

// LoadRoutes 加载路由配置
func (m *Manager) LoadRoutes(path string) error {
    return m.router.LoadFromConfig(path)
}
```

- [ ] **Step 4: Run test to verify it passes**

Run: `go test ./internal/plugin/... -v`
Expected: PASS

- [ ] **Step 5: Commit**

```bash
git add internal/plugin/manager.go internal/plugin/manager_test.go
git commit -m "feat(plugin): add PluginManager with module registration"
```

---

## 3.2 实现 DataAccessor 适配器

**Files:**
- Create: `internal/plugin/adapter.go`
- Create: `internal/plugin/adapter_test.go`

- [ ] **Step 1: Write the failing test**

```go
// internal/plugin/adapter_test.go
package plugin

import (
    "testing"
    
    "github.com/yourorg/wg_ai/internal/data"
)

func TestDataAdapterGetField(t *testing.T) {
    playerData := data.NewPlayerData(1)
    playerData.SetField("name", "player1")
    playerData.SetField("level", int64(10))
    
    adapter := NewDataAdapter(1, playerData)
    
    val, err := adapter.GetField("name")
    if err != nil {
        t.Fatal(err)
    }
    if val != "player1" {
        t.Errorf("expected 'player1', got '%v'", val)
    }
    
    val, err = adapter.GetField("level")
    if err != nil {
        t.Fatal(err)
    }
    if val != int64(10) {
        t.Errorf("expected 10, got %v", val)
    }
}

func TestDataAdapterSetField(t *testing.T) {
    playerData := data.NewPlayerData(1)
    adapter := NewDataAdapter(1, playerData)
    
    err := adapter.SetField("gold", int64(1000))
    if err != nil {
        t.Fatal(err)
    }
    
    val, _ := adapter.GetField("gold")
    if val != int64(1000) {
        t.Errorf("expected 1000, got %v", val)
    }
    
    // 验证脏标记
    if !playerData.Dirty {
        t.Error("expected dirty flag")
    }
}

func TestDataAdapterGetArray(t *testing.T) {
    playerData := data.NewPlayerData(1)
    
    // 设置数组
    items := []map[string]any{{"id": 1}, {"id": 2}}
    playerData.Arrays["items"] = &items
    
    adapter := NewDataAdapter(1, playerData)
    
    arr, err := adapter.GetArray("items")
    if err != nil {
        t.Fatal(err)
    }
    
    itemsPtr, ok := arr.(*[]map[string]any)
    if !ok {
        t.Fatal("type assertion failed")
    }
    
    if len(*itemsPtr) != 2 {
        t.Errorf("expected 2 items, got %d", len(*itemsPtr))
    }
}

func TestDataAdapterMarkDirty(t *testing.T) {
    playerData := data.NewPlayerData(1)
    playerData.Dirty = false
    
    adapter := NewDataAdapter(1, playerData)
    adapter.MarkDirty()
    
    if !playerData.Dirty {
        t.Error("expected dirty flag after MarkDirty")
    }
}
```

- [ ] **Step 2: Run test to verify it fails**

Run: `go test ./internal/plugin/... -v -run TestAdapter`
Expected: FAIL

- [ ] **Step 3: Write minimal implementation**

```go
// internal/plugin/adapter.go
package plugin

import (
    "github.com/yourorg/wg_ai/internal/data"
    p "github.com/yourorg/wg_ai/plugin"
)

// DataAdapter 数据访问适配器 - 连接 PlayerData 和 DataAccessor 接口
type DataAdapter struct {
    rid  int64
    data *data.PlayerData
}

// NewDataAdapter 创建数据适配器
func NewDataAdapter(rid int64, playerData *data.PlayerData) *DataAdapter {
    return &DataAdapter{
        rid:  rid,
        data: playerData,
    }
}

// GetField 实现 DataAccessor 接口
func (a *DataAdapter) GetField(key string) (any, error) {
    return a.data.GetField(key), nil
}

// SetField 实现 DataAccessor 接口
func (a *DataAdapter) SetField(key string, value any) error {
    a.data.SetField(key, value)
    a.data.Lock()
    a.data.Dirty = true
    a.data.Unlock()
    return nil
}

// GetArray 实现 DataAccessor 接口 - 返回指针，可直接修改
func (a *DataAdapter) GetArray(key string) (any, error) {
    return a.data.GetArray(key), nil
}

// MarkDirty 实现 DataAccessor 接口
func (a *DataAdapter) MarkDirty() {
    a.data.Lock()
    a.data.Dirty = true
    a.data.Unlock()
}

// RID 获取玩家ID
func (a *DataAdapter) RID() int64 {
    return a.rid
}
```

- [ ] **Step 4: Run test to verify it passes**

Run: `go test ./internal/plugin/... -v -run TestAdapter`
Expected: PASS

- [ ] **Step 5: Commit**

```bash
git add internal/plugin/adapter.go internal/plugin/adapter_test.go
git commit -m "feat(plugin): add DataAdapter for PlayerData access"
```

---

## 3.3 实现 Session 推送适配器

**Files:**
- Create: `internal/plugin/session.go`
- Create: `internal/plugin/session_test.go`

- [ ] **Step 1: Write the failing test**

```go
// internal/plugin/session_test.go
package plugin

import (
    "net"
    "testing"
    
    "github.com/yourorg/wg_ai/internal/session"
)

func TestSessionAdapterPush(t *testing.T) {
    // 创建 mock connection
    server, client := net.Pipe()
    defer server.Close()
    defer client.Close()
    
    sess := session.New(1, 1, server)
    adapter := NewSessionAdapter(sess)
    
    // 测试推送
    done := make(chan []byte, 1)
    go func() {
        buf := make([]byte, 1024)
        n, _ := client.Read(buf)
        done <- buf[:n]
    }()
    
    err := adapter.Push(1001, []byte("test"))
    if err != nil {
        t.Fatal(err)
    }
    
    data := <-done
    if len(data) == 0 {
        t.Error("expected data")
    }
}
```

- [ ] **Step 2: Run test to verify it fails**

Run: `go test ./internal/plugin/... -v -run TestSession`
Expected: FAIL

- [ ] **Step 3: Write minimal implementation**

```go
// internal/plugin/session.go
package plugin

import (
    "encoding/binary"
    
    "github.com/yourorg/wg_ai/internal/session"
    p "github.com/yourorg/wg_ai/plugin"
)

// SessionAdapter 会话推送适配器
type SessionAdapter struct {
    sess *session.Session
}

// NewSessionAdapter 创建会话适配器
func NewSessionAdapter(sess *session.Session) *SessionAdapter {
    return &SessionAdapter{sess: sess}
}

// Push 实现 SessionPush 接口 - 推送消息给客户端
func (a *SessionAdapter) Push(msgID uint16, data []byte) error {
    packet := makePacket(msgID, data)
    return a.sess.Send(packet)
}

// makePacket 构造消息包
// 格式: [总长度4字节][消息ID 2字节][消息体 N字节]
func makePacket(msgID uint16, body []byte) []byte {
    totalLen := 4 + 2 + len(body) // len + msgID + body
    packet := make([]byte, totalLen)
    
    // 写入总长度
    binary.BigEndian.PutUint32(packet[0:4], uint32(totalLen))
    // 写入消息ID
    binary.BigEndian.PutUint16(packet[4:6], msgID)
    // 写入消息体
    copy(packet[6:], body)
    
    return packet
}
```

- [ ] **Step 4: Run test to verify it passes**

Run: `go test ./internal/plugin/... -v -run TestSession`
Expected: PASS

- [ ] **Step 5: Commit**

```bash
git add internal/plugin/session.go internal/plugin/session_test.go
git commit -m "feat(plugin): add SessionAdapter for client push"
```

---

## 3.4 实现 Call 调用逻辑

**Files:**
- Modify: `internal/plugin/manager.go`
- Modify: `internal/plugin/manager_test.go`

- [ ] **Step 1: Write the failing test**

```go
// 添加到 internal/plugin/manager_test.go

func TestManagerCall(t *testing.T) {
    mgr := NewManager()
    mgr.RegisterModule("role", &MockLogicModule{name: "role"})
    
    // 注册路由
    mgr.Router().Register(1001, "role", "login")
    
    // 创建上下文
    playerData := data.NewPlayerData(1)
    ctx := &p.LogicContext{
        RID:  1,
        UID:  1,
        Data: NewDataAdapter(1, playerData),
    }
    
    // 调用
    result, err := mgr.Call(ctx, 1001, map[string]any{})
    if err != nil {
        t.Fatal(err)
    }
    if result.Code != 0 {
        t.Errorf("expected code 0, got %d", result.Code)
    }
}

func TestManagerCallModuleNotFound(t *testing.T) {
    mgr := NewManager()
    mgr.Router().Register(9999, "notexist", "test")
    
    playerData := data.NewPlayerData(1)
    ctx := &p.LogicContext{
        RID:  1,
        Data: NewDataAdapter(1, playerData),
    }
    
    _, err := mgr.Call(ctx, 9999, nil)
    if err != p.ErrModuleNotFound {
        t.Errorf("expected ErrModuleNotFound, got %v", err)
    }
}

func TestManagerCallRouteNotFound(t *testing.T) {
    mgr := NewManager()
    
    playerData := data.NewPlayerData(1)
    ctx := &p.LogicContext{
        RID:  1,
        Data: NewDataAdapter(1, playerData),
    }
    
    _, err := mgr.Call(ctx, 8888, nil)
    if err != p.ErrModuleNotFound {
        t.Errorf("expected ErrModuleNotFound, got %v", err)
    }
}
```

- [ ] **Step 2: Run test to verify it fails**

Run: `go test ./internal/plugin/... -v -run TestManagerCall`
Expected: FAIL

- [ ] **Step 3: Add Call method to manager.go**

```go
// 添加到 internal/plugin/manager.go

// Call 调用逻辑处理
func (m *Manager) Call(ctx *p.LogicContext, msgID uint16, params map[string]any) (*p.LogicResult, error) {
    // 1. 根据消息ID获取路由
    route, ok := m.router.Get(msgID)
    if !ok {
        return nil, p.ErrModuleNotFound
    }
    
    // 2. 获取模块
    m.mutex.RLock()
    module := m.modules[route.Module]
    m.mutex.RUnlock()
    
    if module == nil {
        return nil, p.ErrModuleNotFound
    }
    
    // 3. 调用处理方法
    return module.Handle(ctx, route.Method, params)
}
```

- [ ] **Step 4: Run test to verify it passes**

Run: `go test ./internal/plugin/... -v -run TestManagerCall`
Expected: PASS

- [ ] **Step 5: Commit**

```bash
git add internal/plugin/manager.go internal/plugin/manager_test.go
git commit -m "feat(plugin): add Call method for message dispatch"
```

---

## Phase 3 Summary

| File | Description |
|------|-------------|
| `internal/plugin/manager.go` | 插件管理器核心 |
| `internal/plugin/manager_test.go` | 单元测试 |
| `internal/plugin/adapter.go` | DataAccessor 适配器 |
| `internal/plugin/adapter_test.go` | 单元测试 |
| `internal/plugin/session.go` | Session 推送适配器 |
| `internal/plugin/session_test.go` | 单元测试 |
