# 玩家数据热更方案设计

## 概述

基于 Go Plugin 实现玩家数据处理逻辑的热更新，核心原则：
- **数据与逻辑分离**：数据层不可热更，逻辑层可热更
- **指针传递**：使用 `map[string]any` + `slice` 存储数据，传递指针避免序列化开销
- **稳定的数据结构**：内存 struct + 数据库表 + protobuf 协议均不可热更

## 整体架构

```
┌────────────────────────────────────────────────────────────────┐
│                          Game Server                            │
│                                                                 │
│  ┌─────────────────────────────────────────────────────────┐   │
│  │                     Gateway Layer                        │   │
│  │  TCPServer ──► SessionMgr ──► AgentMgr                  │   │
│  └──────────────────────────────┬──────────────────────────┘   │
│                                 │                               │
│  ┌──────────────────────────────▼──────────────────────────┐   │
│  │                   Data Layer (不可热更)                   │   │
│  │  ┌─────────────┐  ┌─────────────┐  ┌─────────────────┐  │   │
│  │  │ PlayerStore │  │  ItemStore  │  │   CacheManager  │  │   │
│  │  └─────────────┘  └─────────────┘  └─────────────────┘  │   │
│  └──────────────────────────────┬──────────────────────────┘   │
│                                 │                               │
│  ┌──────────────────────────────▼──────────────────────────┐   │
│  │                   Logic Layer (可热更 .so)               │   │
│  │  ┌──────────────────────────────────────────────────┐   │   │
│  │  │              PluginManager                        │   │   │
│  │  └──────────────────────────────────────────────────┘   │   │
│  │  ┌───────────┐ ┌───────────┐ ┌───────────┐             │   │
│  │  │RoleLogic  │ │ ItemLogic │ │ HeroLogic │  ...        │   │
│  │  │  .so      │ │   .so     │ │   .so     │             │   │
│  │  └───────────┘ └───────────┘ └───────────┘             │   │
│  └───────────────────────────────────────────────────────┘   │
│                                                                 │
│  ┌─────────────────────────────────────────────────────────┐   │
│  │                   Persistence Layer                      │   │
│  │  MySQL (持久化) ◄─── DataManager 定时落库               │   │
│  └─────────────────────────────────────────────────────────┘   │
└────────────────────────────────────────────────────────────────┘
```

## 数据层设计 (Data Layer)

### 核心数据结构

```go
// internal/data/types.go

// PlayerData 玩家数据容器
type PlayerData struct {
    RID     int64               `json:"rid"`
    Base    map[string]any      `json:"base"`    // 基础字段: name, level, exp, gold...
    Arrays  map[string]any      `json:"arrays"`  // 数组字段: items, heroes, tasks...
    Dirty   bool                `json:"-"`       // 脏标记
    mutex   sync.RWMutex        `json:"-"`
}

// 示例数据结构:
// Base: {
//     "name": "player1",
//     "level": 10,
//     "exp": 5000,
//     "gold": 1000,
//     "vip": 0,
// }
// Arrays: {
//     "items": []ItemData{...},
//     "heroes": []HeroData{...},
// }
```

### DataStore 接口

```go
// internal/data/store.go

// DataStore 数据存储接口 - 供逻辑层调用
type DataStore interface {
    // 获取玩家数据指针
    GetPlayer(rid int64) (*PlayerData, error)

    // 获取基础字段 (返回指针，可直接修改)
    GetField(rid int64, key string) (any, error)
    SetField(rid int64, key string, value any) error

    // 获取数组字段 (返回指针，可直接修改)
    GetArray(rid int64, key string) (any, error)

    // 标记脏数据
    MarkDirty(rid int64)

    // 批量获取 (用于跨玩家操作)
    GetPlayers(rids []int64) ([]*PlayerData, error)
}

// PlayerStore 玩家数据管理器
type PlayerStore struct {
    players map[int64]*PlayerData
    mutex   sync.RWMutex
    db      *sql.DB
    redis   *redis.Client
}

func NewPlayerStore(db *sql.DB, redis *redis.Client) *PlayerStore
func (s *PlayerStore) Load(rid int64) (*PlayerData, error)
func (s *PlayerStore) Save(rid int64) error
func (s *PlayerStore) SaveAll() error  // 定时落库
```

### 数据访问示例

```go
// 逻辑层通过 DataStore 接口操作数据
func (l *ItemLogic) UseItem(ctx *LogicContext, itemID int64, count int) error {
    // 获取背包数据指针
    itemsPtr, err := ctx.Data.GetArray(ctx.RID, "items")
    if err != nil {
        return err
    }

    items := itemsPtr.(*[]ItemData)

    // 直接修改 (指针操作，无需序列化)
    for i, item := range *items {
        if item.ID == itemID {
            (*items)[i].Count -= count
            ctx.Data.MarkDirty(ctx.RID)  // 标记脏数据
            break
        }
    }

    return nil
}
```

## 逻辑层设计 (Logic Layer)

### 插件接口定义

```go
// plugin/interface.go (主程序和插件共用)

// LogicModule 逻辑模块接口
type LogicModule interface {
    // 模块名称
    Name() string

    // 处理请求
    Handle(ctx *LogicContext, method string, params map[string]any) (*LogicResult, error)
}

// LogicContext 逻辑上下文
type LogicContext struct {
    RID      int64           // 玩家ID
    Data     DataAccessor    // 数据访问接口
    Session  any             // Session 引用 (用于推送)
}

// DataAccessor 数据访问接口 (简化版，供插件使用)
type DataAccessor interface {
    GetField(key string) (any, error)
    SetField(key string, value any) error
    GetArray(key string) (any, error)
    MarkDirty()
}

// LogicResult 处理结果
type LogicResult struct {
    Code     int             `json:"code"`
    Data     map[string]any  `json:"data"`
    Push     []PushData      `json:"push"`    // 需要推送的数据
}

type PushData struct {
    MsgID    uint16          `json:"msg_id"`
    Data     []byte          `json:"data"`
}
```

### 插件实现示例

```go
// plugin/role/role.go

package main

import "plugin/interface"

type RoleLogic struct{}

func init() {
    // 注册到全局
    LogicModules["role"] = &RoleLogic{}
}

func (l *RoleLogic) Name() string {
    return "role"
}

func (l *RoleLogic) Handle(ctx *LogicContext, method string, params map[string]any) (*LogicResult, error) {
    switch method {
    case "login":
        return l.handleLogin(ctx, params)
    case "level_up":
        return l.handleLevelUp(ctx, params)
    }
    return nil, ErrMethodNotFound
}

func (l *RoleLogic) handleLogin(ctx *LogicContext, params map[string]any) (*LogicResult, error) {
    // 获取玩家等级
    level, _ := ctx.Data.GetField("level")
    name, _ := ctx.Data.GetField("name")

    return &LogicResult{
        Code: 0,
        Data: map[string]any{
            "level": level,
            "name":  name,
        },
    }, nil
}

// 导出符号
var RoleModule LogicModule = &RoleLogic{}
```

### 插件管理器

```go
// internal/plugin/manager.go

type PluginManager struct {
    plugins map[string]*plugin.Plugin  // module -> plugin
    modules map[string]LogicModule     // module -> LogicModule
    mutex   sync.RWMutex
}

func NewPluginManager() *PluginManager

// LoadPlugin 加载插件
func (m *PluginManager) LoadPlugin(module, path string) error {
    p, err := plugin.Open(path)
    if err != nil {
        return err
    }

    sym, err := p.Lookup(module + "Module")
    if err != nil {
        return err
    }

    logicModule, ok := sym.(LogicModule)
    if !ok {
        return ErrInvalidModule
    }

    m.mutex.Lock()
    m.plugins[module] = p
    m.modules[module] = logicModule
    m.mutex.Unlock()

    return nil
}

// HotReload 热更新插件
func (m *PluginManager) HotReload(module, newPath string) error {
    // Go plugin 无法卸载，直接加载新版本覆盖
    return m.LoadPlugin(module, newPath)
}

// Call 调用逻辑
func (m *PluginManager) Call(module, method string, ctx *LogicContext, params map[string]any) (*LogicResult, error) {
    m.mutex.RLock()
    logic := m.modules[module]
    m.mutex.RUnlock()

    if logic == nil {
        return nil, ErrModuleNotFound
    }

    return logic.Handle(ctx, method, params)
}
```

## 热更流程

### 1. 热更触发方式

```go
// 支持两种方式:
// 1. API 触发 (HTTP 接口)
// 2. 文件监听 (fsnotify)

// API 方式
POST /admin/hotreload
{
    "module": "role",
    "path": "./plugins/role.so"
}

// 文件监听方式
// 监听 ./plugins/ 目录，检测 .so 文件变化自动热更
```

### 2. 热更流程图

```
热更请求
    │
    ▼
┌─────────────┐
│ 校验插件文件 │
└──────┬──────┘
       │
       ▼
┌─────────────┐
│ 编译新插件   │ (可选，支持自动编译)
└──────┬──────┘
       │
       ▼
┌─────────────┐
│ 加载 .so    │
└──────┬──────┘
       │
       ▼
┌─────────────┐
│ 查找导出符号 │
└──────┬──────┘
       │
       ▼
┌─────────────┐
│ 类型断言    │
└──────┬──────┘
       │
       ▼
┌─────────────┐
│ 替换旧模块   │ (旧 .so 无法卸载，保留内存)
└──────┬──────┘
       │
       ▼
   热更完成
```

### 3. 热更命令

```bash
# 编译插件
go build -buildmode=plugin -o plugins/role.so plugin/role/role.go

# 热更 (通过 API)
curl -X POST http://localhost:8080/admin/hotreload \
  -d '{"module": "role", "path": "./plugins/role.so"}'
```

## 文件结构

```
wg_ai/
├── cmd/
│   └── game/
│       └── main.go              # 入口
├── internal/
│   ├── data/                    # 数据层 (不可热更)
│   │   ├── store.go             # DataStore 实现
│   │   ├── types.go             # PlayerData 等类型
│   │   └── persist.go           # 持久化逻辑
│   ├── plugin/                  # 插件管理 (不可热更)
│   │   ├── manager.go           # PluginManager
│   │   └── interface.go         # DataAccessor 实现
│   ├── agent/                   # Agent (不可热更)
│   │   ├── agent.go
│   │   └── dispatcher.go        # 改为调用 PluginManager
│   ├── gate/                    # 网关层 (不可热更)
│   │   └── ...
│   └── db/                      # 数据库 (不可热更)
│       └── ...
├── plugin/                      # 插件源码 (可热更)
│   ├── interface/               # 接口定义 (共用)
│   │   └── interface.go
│   ├── role/                    # 角色逻辑
│   │   └── role.go
│   ├── item/                    # 物品逻辑
│   │   └── item.go
│   └── hero/                    # 英雄逻辑
│       └── hero.go
├── plugins/                     # 编译后的插件 (可热更)
│   ├── role.so
│   ├── item.so
│   └── hero.so
└── Makefile                     # 编译脚本
```

## Makefile

```makefile
# 编译主程序
build:
	go build -o bin/game ./cmd/game

# 编译所有插件
plugins:
	go build -buildmode=plugin -o plugins/role.so ./plugin/role
	go build -buildmode=plugin -o plugins/item.so ./plugin/item
	go build -buildmode=plugin -o plugins/hero.so ./plugin/hero

# 编译单个插件
plugin-%:
	go build -buildmode=plugin -o plugins/$*.so ./plugin/$*

# 全量编译
all: build plugins
```

## 实现步骤

### Phase 1: 数据层重构
1. 创建 `internal/data/` 目录
2. 定义 `PlayerData` 结构
3. 实现 `PlayerStore` 和 `DataStore` 接口
4. 实现数据加载/保存逻辑

### Phase 2: 插件基础设施
1. 创建 `plugin/interface/` 接口定义
2. 创建 `internal/plugin/` 插件管理器
3. 实现 `PluginManager`
4. 实现 `DataAccessor` 适配器

### Phase 3: 逻辑层迁移
1. 创建示例插件 `plugin/role/`
2. 将现有 Handler 逻辑迁移到插件
3. 修改 Agent Dispatcher 调用 PluginManager

### Phase 4: 热更机制
1. 实现热更 API 接口
2. (可选) 实现文件监听自动热更
3. 编写 Makefile 编译脚本

### Phase 5: 测试与文档
1. 编写单元测试
2. 编写集成测试
3. 编写使用文档

## 约束与限制

1. **平台限制**: Go Plugin 仅支持 Linux/macOS，不支持 Windows
2. **编译一致性**: 插件和主程序必须使用完全相同的 Go 版本编译
3. **内存泄漏**: 旧版本插件无法卸载，多次热更会累积内存
4. **类型安全**: `map[string]any` 方案牺牲了编译时类型检查

## 风险缓解

| 风险 | 缓解措施 |
|------|----------|
| 插件加载失败 | 保留旧版本继续运行，记录错误日志 |
| 热更导致崩溃 | 实现 recover 机制，崩溃时回滚到旧版本 |
| 内存泄漏 | 定期重启服务（低峰期），限制热更频率 |
| 类型不兼容 | 热更前校验插件符号，确保接口匹配 |
