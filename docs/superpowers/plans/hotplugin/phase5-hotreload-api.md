# Phase 5: 热更 API

> **Goal:** 提供 HTTP 热更入口和可选的文件监听自动热更

---

## 5.1 HTTP 热更接口

**Files:**
- Create: `internal/admin/handler.go`
- Create: `internal/admin/handler_test.go`

- [ ] **Step 1: Write the failing test**

```go
// internal/admin/handler_test.go
package admin

import (
    "bytes"
    "encoding/json"
    "net/http"
    "net/http/httptest"
    "testing"
    
    "github.com/yourorg/wg_ai/internal/plugin"
)

func TestHandleHotReload(t *testing.T) {
    // 创建插件管理器
    mgr := plugin.NewManager()
    handler := NewHandler(mgr)
    
    // 创建测试服务器
    mux := http.NewServeMux()
    handler.RegisterRoutes(mux)
    srv := httptest.NewServer(mux)
    defer srv.Close()
    
    // 测试热更请求 (不存在的插件，会失败)
    body := HotReloadRequest{
        Module: "test",
        Path:   "./plugins/test.so",
    }
    jsonBody, _ := json.Marshal(body)
    
    resp, err := http.Post(srv.URL+"/admin/hotreload", "application/json", bytes.NewReader(jsonBody))
    if err != nil {
        t.Fatal(err)
    }
    defer resp.Body.Close()
    
    // 预期失败 (插件不存在)
    var result HotReloadResponse
    json.NewDecoder(resp.Body).Decode(&result)
    
    if result.Module != "test" {
        t.Error("module mismatch")
    }
}

func TestHandleListPlugins(t *testing.T) {
    mgr := plugin.NewManager()
    handler := NewHandler(mgr)
    
    mux := http.NewServeMux()
    handler.RegisterRoutes(mux)
    srv := httptest.NewServer(mux)
    defer srv.Close()
    
    resp, err := http.Get(srv.URL + "/admin/plugins")
    if err != nil {
        t.Fatal(err)
    }
    defer resp.Body.Close()
    
    if resp.StatusCode != http.StatusOK {
        t.Errorf("expected 200, got %d", resp.StatusCode)
    }
}

func TestHandleHealth(t *testing.T) {
    mgr := plugin.NewManager()
    handler := NewHandler(mgr)
    
    mux := http.NewServeMux()
    handler.RegisterRoutes(mux)
    srv := httptest.NewServer(mux)
    defer srv.Close()
    
    resp, err := http.Get(srv.URL + "/admin/health")
    if err != nil {
        t.Fatal(err)
    }
    defer resp.Body.Close()
    
    var result map[string]any
    json.NewDecoder(resp.Body).Decode(&result)
    
    if result["status"] != "ok" {
        t.Error("health check failed")
    }
}
```

- [ ] **Step 2: Run test to verify it fails**

Run: `go test ./internal/admin/... -v`
Expected: FAIL

- [ ] **Step 3: Write minimal implementation**

```go
// internal/admin/handler.go
package admin

import (
    "encoding/json"
    "net/http"
    
    "github.com/yourorg/wg_ai/internal/plugin"
)

// Handler 管理接口处理器
type Handler struct {
    pluginMgr *plugin.Manager
}

// NewHandler 创建管理接口处理器
func NewHandler(pluginMgr *plugin.Manager) *Handler {
    return &Handler{pluginMgr: pluginMgr}
}

// RegisterRoutes 注册路由
func (h *Handler) RegisterRoutes(mux *http.ServeMux) {
    mux.HandleFunc("/admin/hotreload", h.handleHotReload)
    mux.HandleFunc("/admin/plugins", h.handleListPlugins)
    mux.HandleFunc("/admin/health", h.handleHealth)
}

// HotReloadRequest 热更请求
type HotReloadRequest struct {
    Module string `json:"module"` // 模块名: role, item
    Path   string `json:"path"`   // 插件路径: ./plugins/role.so
}

// HotReloadResponse 热更响应
type HotReloadResponse struct {
    Success bool   `json:"success"`
    Message string `json:"message"`
    Module  string `json:"module"`
}

// handleHotReload 处理热更请求
// POST /admin/hotreload
func (h *Handler) handleHotReload(w http.ResponseWriter, r *http.Request) {
    if r.Method != http.MethodPost {
        http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
        return
    }
    
    var req HotReloadRequest
    if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
        http.Error(w, err.Error(), http.StatusBadRequest)
        return
    }
    
    // 参数校验
    if req.Module == "" || req.Path == "" {
        json.NewEncoder(w).Encode(HotReloadResponse{
            Success: false,
            Message: "module and path are required",
            Module:  req.Module,
        })
        return
    }
    
    // 执行热更
    err := h.pluginMgr.HotReload(req.Module, req.Path)
    if err != nil {
        json.NewEncoder(w).Encode(HotReloadResponse{
            Success: false,
            Message: err.Error(),
            Module:  req.Module,
        })
        return
    }
    
    json.NewEncoder(w).Encode(HotReloadResponse{
        Success: true,
        Message: "hot reload success",
        Module:  req.Module,
    })
}

// handleListPlugins 列出已加载插件
// GET /admin/plugins
func (h *Handler) handleListPlugins(w http.ResponseWriter, r *http.Request) {
    modules := h.pluginMgr.ListModules()
    
    w.Header().Set("Content-Type", "application/json")
    json.NewEncoder(w).Encode(map[string]any{
        "plugins": modules,
        "count":   len(modules),
    })
}

// handleHealth 健康检查
// GET /admin/health
func (h *Handler) handleHealth(w http.ResponseWriter, r *http.Request) {
    w.Header().Set("Content-Type", "application/json")
    json.NewEncoder(w).Encode(map[string]any{
        "status": "ok",
    })
}
```

- [ ] **Step 4: Run test to verify it passes**

Run: `go test ./internal/admin/... -v`
Expected: PASS

- [ ] **Step 5: Commit**

```bash
git add internal/admin/handler.go internal/admin/handler_test.go
git commit -m "feat(admin): add HTTP hotreload API with /admin/hotreload"
```

---

## 5.2 文件监听自动热更 (可选)

**Files:**
- Create: `internal/plugin/watcher.go`
- Create: `internal/plugin/watcher_test.go`

- [ ] **Step 1: Write the failing test**

```go
// internal/plugin/watcher_test.go
package plugin

import (
    "os"
    "path/filepath"
    "testing"
    "time"
)

func TestWatcherStart(t *testing.T) {
    // 创建临时目录
    tmpDir, err := os.MkdirTemp("", "plugin_test")
    if err != nil {
        t.Fatal(err)
    }
    defer os.RemoveAll(tmpDir)
    
    // 创建空的 .so 文件
    soPath := filepath.Join(tmpDir, "test.so")
    if err := os.WriteFile(soPath, []byte("fake"), 0644); err != nil {
        t.Fatal(err)
    }
    
    // 创建 watcher
    mgr := NewManager()
    watcher, err := NewWatcher(mgr, tmpDir)
    if err != nil {
        t.Fatal(err)
    }
    defer watcher.Stop()
    
    if err := watcher.Start(); err != nil {
        t.Fatal(err)
    }
}

func TestWatcherExtractModule(t *testing.T) {
    tests := []struct {
        path     string
        expected string
    }{
        {"./plugins/role.so", "role"},
        {"/home/user/plugins/item.so", "item"},
        {"plugins/test_v2.so", "test_v2"},
    }
    
    for _, tt := range tests {
        result := extractModule(tt.path)
        if result != tt.expected {
            t.Errorf("extractModule(%s) = %s, want %s", tt.path, result, tt.expected)
        }
    }
}
```

- [ ] **Step 2: Run test to verify it fails**

Run: `go test ./internal/plugin/... -v -run TestWatcher`
Expected: FAIL

- [ ] **Step 3: Write minimal implementation**

```go
// internal/plugin/watcher.go
package plugin

import (
    "path/filepath"
    "strings"
    
    "github.com/fsnotify/fsnotify"
)

// Watcher 文件监听器
type Watcher struct {
    pluginMgr *Manager
    watcher   *fsnotify.Watcher
    pluginDir string
    stopCh    chan struct{}
}

// NewWatcher 创建文件监听器
func NewWatcher(pluginMgr *Manager, pluginDir string) (*Watcher, error) {
    w, err := fsnotify.NewWatcher()
    if err != nil {
        return nil, err
    }
    
    return &Watcher{
        pluginMgr: pluginMgr,
        watcher:   w,
        pluginDir: pluginDir,
        stopCh:    make(chan struct{}),
    }, nil
}

// Start 开始监听
func (w *Watcher) Start() error {
    // 添加监听目录
    if err := w.watcher.Add(w.pluginDir); err != nil {
        return err
    }
    
    go w.run()
    return nil
}

// Stop 停止监听
func (w *Watcher) Stop() {
    close(w.stopCh)
    w.watcher.Close()
}

func (w *Watcher) run() {
    for {
        select {
        case event, ok := <-w.watcher.Events:
            if !ok {
                return
            }
            
            // 检测 .so 文件写入/创建事件
            if (event.Has(fsnotify.Write) || event.Has(fsnotify.Create)) && 
               strings.HasSuffix(event.Name, ".so") {
                w.handleSoChange(event.Name)
            }
            
        case err, ok := <-w.watcher.Errors:
            if !ok {
                return
            }
            // 记录错误日志
            _ = err
            
        case <-w.stopCh:
            return
        }
    }
}

func (w *Watcher) handleSoChange(path string) {
    // 从文件名提取模块名
    module := extractModule(path)
    
    // 执行热更
    if err := w.pluginMgr.HotReload(module, path); err != nil {
        // 记录错误日志
        return
    }
    
    // 记录成功日志
}

// extractModule 从路径提取模块名
// ./plugins/role.so -> role
func extractModule(path string) string {
    base := filepath.Base(path)
    return strings.TrimSuffix(base, ".so")
}
```

- [ ] **Step 4: Run test to verify it passes**

Run: `go test ./internal/plugin/... -v -run TestWatcher`
Expected: PASS

- [ ] **Step 5: Commit**

```bash
git add internal/plugin/watcher.go internal/plugin/watcher_test.go
git commit -m "feat(plugin): add file watcher for auto hotreload"
```

---

## 5.3 更新配置文件

**Files:**
- Modify: `internal/common/config/config.go`
- Create: `config/game.yaml`

- [ ] **Step 1: Add plugin config to config.go**

```go
// 添加到 internal/common/config/config.go

// PluginConfig 插件配置
type PluginConfig struct {
    Dir       string `mapstructure:"dir"`        // 插件目录: ./plugins
    RouteFile string `mapstructure:"route_file"` // 路由配置: ./config/routes.json
    Watch     bool   `mapstructure:"watch"`      // 是否启用文件监听
}

// AdminConfig 管理接口配置
type AdminConfig struct {
    Addr string `mapstructure:"addr"` // 管理接口地址: :8081
}

// GameConfig 添加字段
type GameConfig struct {
    // ... 现有字段
    
    Plugin PluginConfig `mapstructure:"plugin"`
    Admin  AdminConfig  `mapstructure:"admin"`
}
```

- [ ] **Step 2: Create game.yaml**

```yaml
# config/game.yaml
server:
  host: "0.0.0.0"
  port: 8080

gate:
  msg_queue_size: 10000

agent:
  count: 4

plugin:
  dir: "./plugins"
  route_file: "./config/routes.json"
  watch: true

admin:
  addr: ":8081"

database:
  host: "127.0.0.1"
  port: 3306
  database: "game"
  username: "root"
  password: "password"
  max_open: 100
  max_idle: 10

redis:
  host: "127.0.0.1"
  port: 6379
  db: 0
```

- [ ] **Step 3: Commit**

```bash
git add internal/common/config/config.go config/game.yaml
git commit -m "feat(config): add plugin and admin config"
```

---

## 5.4 集成到 Game Server

**Files:**
- Modify: `internal/game/server.go`

- [ ] **Step 1: Update server.go**

```go
// internal/game/server.go
package game

import (
    "net/http"
    
    "github.com/yourorg/wg_ai/internal/admin"
    "github.com/yourorg/wg_ai/internal/agent"
    "github.com/yourorg/wg_ai/internal/common/config"
    "github.com/yourorg/wg_ai/internal/data"
    "github.com/yourorg/wg_ai/internal/gate"
    "github.com/yourorg/wg_ai/internal/plugin"
    "github.com/yourorg/wg_ai/internal/rpc"
    "github.com/yourorg/wg_ai/internal/session"
)

type Server struct {
    config     *config.GameConfig
    tcpServer  *gate.TCPServer
    agentMgr   *agent.Manager
    sessionMgr *session.Manager
    rpcClient  *rpc.Client
    
    // 新增
    dataStore  *data.PlayerStore
    pluginMgr  *plugin.Manager
    watcher    *plugin.Watcher
    adminSrv   *http.Server
}

func NewServer(cfg *config.GameConfig) *Server {
    return &Server{
        config:     cfg,
        sessionMgr: session.NewManager(),
    }
}

func (s *Server) Start() error {
    // 1. 初始化数据层
    s.dataStore = data.NewPlayerStore(nil) // TODO: 传入 db
    
    // 2. 初始化插件管理器
    s.pluginMgr = plugin.NewManager()
    
    // 3. 加载路由配置
    if s.config.Plugin.RouteFile != "" {
        if err := s.pluginMgr.LoadRoutes(s.config.Plugin.RouteFile); err != nil {
            return err
        }
    }
    
    // 4. 加载所有插件
    if s.config.Plugin.Dir != "" {
        if err := s.loadAllPlugins(); err != nil {
            return err
        }
    }
    
    // 5. 启动文件监听 (可选)
    if s.config.Plugin.Watch {
        s.startWatcher()
    }
    
    // 6. 创建 Agent Manager
    s.agentMgr = agent.NewManager(
        s.config.Agent.Count,
        s.config.Gate.MsgQueueSize,
        s.pluginMgr,
        s.dataStore,
    )
    
    // 7. 连接 RPC
    s.rpcClient = rpc.NewClient(&rpc.ClientConfig{
        DBAddr:    s.config.Cluster.DBAddr,
        LoginAddr: s.config.Cluster.LoginAddr,
    })
    
    // 8. 启动 TCP 服务
    addr := s.config.Server.Addr()
    s.tcpServer = gate.NewTCPServer(addr, s.sessionMgr, s.agentMgr)
    
    go func() {
        if err := s.tcpServer.Start(); err != nil {
            panic(err)
        }
    }()
    
    // 9. 启动管理接口
    s.startAdminServer()
    
    return nil
}

func (s *Server) loadAllPlugins() error {
    // 遍历插件目录，加载所有 .so 文件
    // TODO: 实现
    return nil
}

func (s *Server) startWatcher() {
    watcher, err := plugin.NewWatcher(s.pluginMgr, s.config.Plugin.Dir)
    if err != nil {
        return
    }
    s.watcher = watcher
    watcher.Start()
}

func (s *Server) startAdminServer() {
    mux := http.NewServeMux()
    handler := admin.NewHandler(s.pluginMgr)
    handler.RegisterRoutes(mux)
    
    s.adminSrv = &http.Server{
        Addr:    s.config.Admin.Addr,
        Handler: mux,
    }
    
    go s.adminSrv.ListenAndServe()
}

func (s *Server) Stop() {
    // 关闭管理接口
    if s.adminSrv != nil {
        s.adminSrv.Close()
    }
    
    // 停止文件监听
    if s.watcher != nil {
        s.watcher.Stop()
    }
    
    // 关闭 TCP 服务
    if s.tcpServer != nil {
        s.tcpServer.Stop()
    }
    
    // 关闭 Agent
    if s.agentMgr != nil {
        s.agentMgr.Stop()
    }
    
    // 关闭 RPC
    if s.rpcClient != nil {
        s.rpcClient.Close()
    }
}
```

- [ ] **Step 2: Commit**

```bash
git add internal/game/server.go
git commit -m "feat(game): integrate plugin manager and admin server"
```

---

## 5.5 使用文档

**Files:**
- Create: `docs/hotreload-guide.md`

- [ ] **Step 1: Write documentation**

```markdown
# 热更新使用指南

## 快速开始

### 1. 编译插件

```bash
make plugin-role    # 编译角色模块
make plugin-item    # 编译物品模块
make plugins        # 编译所有插件
```

### 2. 启动服务

```bash
make build
./bin/game -config ./config/game.yaml
```

### 3. 热更插件

**方式一: HTTP API**

```bash
# 编译新版本
make plugin-role

# 调用热更接口
curl -X POST http://localhost:8081/admin/hotreload \
  -H "Content-Type: application/json" \
  -d '{"module": "role", "path": "./plugins/role.so"}'
```

**方式二: 文件监听 (自动)**

配置 `plugin.watch: true` 后，直接替换 `plugins/` 目录下的 `.so` 文件即可自动触发热更。

### 4. 查看已加载插件

```bash
curl http://localhost:8081/admin/plugins
```

### 5. 健康检查

```bash
curl http://localhost:8081/admin/health
```

## 注意事项

1. **编译一致性**: 主程序和插件必须使用相同的 Go 版本编译
2. **内存泄漏**: 旧版本插件无法卸载，多次热更会累积内存
3. **建议在低峰期执行热更操作**
4. **热更前先在测试环境验证**

## 开发新插件

1. 在 `plugin/` 目录创建新模块:

```go
// plugin/mymodule/mymodule.go
package main

import p "github.com/yourorg/wg_ai/plugin"

type MyLogic struct{}

func (l *MyLogic) Name() string { return "mymodule" }

func (l *MyLogic) Handle(ctx *p.LogicContext, method string, params map[string]any) (*p.LogicResult, error) {
    // 实现逻辑
    return p.Success(nil), nil
}

var MyModule p.LogicModule = &MyLogic{}
```

2. 添加路由配置 `config/routes.json`:

```json
{"msg_id": 3001, "module": "mymodule", "method": "do_something"}
```

3. 添加编译目标到 Makefile:

```makefile
plugin-mymodule:
	go build -buildmode=plugin -o plugins/mymodule.so ./plugin/mymodule
```
```

- [ ] **Step 2: Commit**

```bash
git add docs/hotreload-guide.md
git commit -m "docs: add hotreload usage guide"
```

---

## Phase 5 Summary

| File | Description |
|------|-------------|
| `internal/admin/handler.go` | HTTP 热更接口 |
| `internal/admin/handler_test.go` | 单元测试 |
| `internal/plugin/watcher.go` | 文件监听 (可选) |
| `internal/plugin/watcher_test.go` | 单元测试 |
| `internal/common/config/config.go` | 添加插件配置 |
| `config/game.yaml` | 配置示例 |
| `internal/game/server.go` | 集成插件管理器 |
| `docs/hotreload-guide.md` | 使用文档 |

---

## 全部 Phase 完成清单

| Phase | 文件数 | 状态 |
|-------|--------|------|
| Phase 1: 数据层 | 6 | 待实现 |
| Phase 2: 插件接口 | 5 | 待实现 |
| Phase 3: 插件管理器 | 6 | 待实现 |
| Phase 4: 示例插件 | 5 | 待实现 |
| Phase 5: 热更 API | 8 | 待实现 |
| **总计** | **30** | |
