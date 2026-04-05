# WG_AI 游戏服务器架构文档

> 本文档记录项目的整体架构、模块设计和实现细节。随着项目开发持续更新。

## 项目概述

WG_AI 是一个基于 Go 语言的游戏服务器框架，支持热更新插件系统。

### 技术栈

- **语言**: Go 1.21+
- **网络**: TCP 长连接
- **数据存储**: MySQL + Redis
- **序列化**: Protobuf
- **插件系统**: Go Plugin (.so 动态加载)

## 目录结构

```
wg_ai/
├── cmd/                    # 程序入口
│   ├── game/              # 游戏服务器
│   ├── login/             # 登录服务器
│   └── db/                # 数据库服务器
├── config/                # 配置文件
│   ├── game.yaml          # 游戏服务器配置
│   └── routes.json        # 消息路由配置
├── internal/              # 内部包 (不可被外部导入)
│   ├── admin/             # HTTP 管理接口
│   ├── agent/             # 消息处理代理
│   ├── auth/              # JWT 认证
│   ├── common/            # 公共组件
│   │   ├── config/        # 配置解析
│   │   ├── errors/        # 错误定义
│   │   ├── health/        # 健康检查
│   │   ├── logger/        # 日志系统
│   │   └── metrics/       # 指标统计
│   ├── data/              # 玩家数据层
│   ├── db/                # 数据库连接
│   ├── game/              # 游戏服务器主逻辑
│   ├── gate/              # TCP 网关
│   ├── plugin/            # 插件管理器
│   ├── rpc/               # RPC 通信
│   ├── scene/             # 场景/地图系统 (九宫格AOI)
│   └── session/           # 会话管理
├── plugin/                # 插件接口定义 & 插件实现
│   ├── interface.go       # 插件接口
│   ├── router.go          # 消息路由
│   ├── role/              # 角色模块插件
│   └── item/              # 物品模块插件
├── proto/                 # Protobuf 协议定义
│   ├── cs/                # 客户端-服务器协议
│   └── ss/                # 服务器-服务器协议
├── tests/                 # 集成测试
│   └── integration/
└── Makefile               # 构建脚本
```

## 核心架构

### 1. 服务器启动流程

```
┌─────────────────────────────────────────────────────────────┐
│                      game.Server.Start()                     │
├─────────────────────────────────────────────────────────────┤
│  1. 初始化数据层 (dataStore)                                 │
│  2. 初始化插件管理器 (pluginMgr)                              │
│  3. 加载路由配置 (routes.json)                               │
│  4. 初始化场景管理器 (sceneMgr) + 注册内置模块                │
│  5. 创建 Agent Manager (工作线程池)                           │
│  6. 连接 RPC (db, login)                                    │
│  7. 启动 TCP 服务 (gate)                                    │
│  8. 启动 Admin HTTP 服务                                     │
│  9. 启动插件文件监听 (可选)                                   │
└─────────────────────────────────────────────────────────────┘
```

### 2. 请求处理流程

```
客户端请求
    │
    ▼
┌─────────────┐
│  TCP Gate   │  接收数据包
└─────────────┘
    │
    ▼
┌─────────────┐
│  Session    │  绑定玩家连接
└─────────────┘
    │
    ▼
┌─────────────┐
│   Agent     │  消息分发 (工作线程)
└─────────────┘
    │
    ▼
┌─────────────┐
│ PluginMgr   │  路由查找: msgID -> (module, method)
└─────────────┘
    │
    ▼
┌─────────────┐
│   Module    │  业务处理 (role/item/scene...)
└─────────────┘
    │
    ▼
  返回响应 + 推送消息
```

### 3. 数据流向

```
┌─────────────────────────────────────────────────────────────┐
│                        PlayerData                            │
├─────────────────────────────────────────────────────────────┤
│  Base: map[string]any                                       │
│    - name, level, exp, gold, scene_id, pos_x, pos_y...      │
│                                                             │
│  Arrays: map[string]any                                      │
│    - items, heroes, tasks... (存储为切片指针，可直接修改)     │
│                                                             │
│  Dirty: bool  (脏标记，用于持久化)                            │
└─────────────────────────────────────────────────────────────┘
         │
         │ DataAdapter (实现 DataAccessor 接口)
         ▼
┌─────────────────────────────────────────────────────────────┐
│                    Plugin LogicContext                       │
│  ctx.Data.GetField("level")                                 │
│  ctx.Data.SetField("gold", 100)                             │
│  ctx.Data.MarkDirty()                                       │
└─────────────────────────────────────────────────────────────┘
```

## 模块详解

### 插件系统 (internal/plugin)

**核心接口** (`plugin/interface.go`):

```go
// 数据访问接口
type DataAccessor interface {
    GetField(key string) (any, error)
    SetField(key string, value any) error
    GetArray(key string) (any, error)
    MarkDirty()
}

// 逻辑模块接口 (插件必须实现)
type LogicModule interface {
    Name() string
    Handle(ctx *LogicContext, method string, params map[string]any) (*LogicResult, error)
}

// 逻辑上下文
type LogicContext struct {
    RID     int64        // 玩家ID
    UID     int64        // 用户ID
    Data    DataAccessor // 数据访问
    Session SessionPush  // 推送接口
}
```

**插件加载**:
- 热更插件: `.so` 文件通过 `plugin.Open()` 加载
- 内置模块: 通过 `RegisterModule()` 直接注册

**已实现插件**:

| 模块 | 文件 | 方法 | 描述 |
|------|------|------|------|
| role | `plugin/role/` | login, heartbeat, get_info, update_name | 角色管理 |
| item | `plugin/item/` | list, use, add | 物品管理 |
| scene | `internal/scene/module.go` | enter, move, leave, get_nearby, get_scene_info | 场景管理 |

### 场景系统 (internal/scene)

**九宫格 AOI 算法**:

```
配置: 地图 1000x1000, 格子 50x50, 共 20x20=400 格子

玩家位置 -> 格子坐标: (250, 250) -> (5, 5)

视野检测 (九宫格):
┌───┬───┬───┐
│4,4│5,4│6,4│
├───┼───┼───┤
│4,5│5,5│6,5│  ← 玩家在(5,5)
├───┼───┼───┤
│4,6│5,6│6,6│
└───┴───┴───┘

事件类型:
- EventEnter: 实体进入视野
- EventLeave: 实体离开视野
```

**核心结构**:

```go
type Manager struct {
    scenes map[int64]*Scene  // 多场景支持
}

type Scene struct {
    ID       int64
    Width    int
    Height   int
    GridSize int
    aoi      *AOI
    entities map[int64]*Entity
}

type AOI struct {
    grids   [][]*Grid  // 2D格子数组
}

type Entity struct {
    ID       int64
    Type     EntityType  // Player, NPC, Monster...
    Position Vector2
    SceneID  int64
    Data     map[string]any
}
```

### 数据层 (internal/data)

```go
type PlayerData struct {
    RID    int64
    Base   map[string]any  // 基础字段
    Arrays map[string]any  // 数组字段 (存储切片指针)
    Dirty  bool
    mutex  sync.RWMutex
}

// 数组字段示例 (item 插件使用)
type ItemData struct {
    ID    int64 `json:"id"`
    CfgID int64 `json:"cfg_id"`
    Count int64 `json:"count"`
}

// 存储方式
Arrays["items"] = &[]ItemData{...}

// 修改方式 (直接操作指针)
items := ctx.Data.GetArray("items").(*[]ItemData)
(*items)[0].Count -= 1
ctx.Data.MarkDirty()
```

### 网关层 (internal/gate)

```go
// 数据包格式
┌──────────┬──────────┬──────────┐
│  MsgID   │  Length  │   Data   │
│ 2 bytes  │  4 bytes │ N bytes  │
└──────────┴──────────┴──────────┘

// TCP Server
type TCPServer struct {
    addr        string
    sessionMgr  *session.Manager
    agentMgr    *agent.Manager
}
```

## 消息路由

**配置文件** (`config/routes.json`):

```json
[
    {"msg_id": 1001, "module": "role", "method": "login"},
    {"msg_id": 1002, "module": "role", "method": "heartbeat"},
    {"msg_id": 1003, "module": "role", "method": "get_info"},
    {"msg_id": 2001, "module": "item", "method": "list"},
    {"msg_id": 2002, "module": "item", "method": "use"},
    {"msg_id": 2003, "module": "item", "method": "add"},
    {"msg_id": 3001, "module": "scene", "method": "enter"},
    {"msg_id": 3002, "module": "scene", "method": "move"},
    {"msg_id": 3003, "module": "scene", "method": "leave"},
    {"msg_id": 3004, "module": "scene", "method": "get_nearby"},
    {"msg_id": 3005, "module": "scene", "method": "get_scene_info"}
]
```

## 热更新

### 编译插件

```bash
make plugin-role    # 编译角色模块
make plugin-item    # 编译物品模块
make plugins        # 编译所有插件
```

### 热更方式

**方式一: HTTP API**

```bash
curl -X POST http://localhost:8081/admin/hotreload \
  -H "Content-Type: application/json" \
  -d '{"module": "role", "path": "./plugins/role.so"}'
```

**方式二: 文件监听**

配置 `plugin.watch: true` 后，直接替换 `.so` 文件自动触发热更。

### Admin API

| 路径 | 方法 | 描述 |
|------|------|------|
| `/admin/plugins` | GET | 列出已加载插件 |
| `/admin/health` | GET | 健康检查 |
| `/admin/routes` | GET | 列出路由配置 |
| `/admin/hotreload` | POST | 热更新插件 |

## 测试

```bash
# 运行所有测试
export PATH=$PATH:/usr/local/go/bin && go test -v ./...

# 运行特定包测试
go test -v ./internal/scene/...

# 运行集成测试
go test -v ./tests/integration/...
```

## 构建命令

```bash
# 编译主程序
make build

# 编译所有插件
make plugins

# 热更 + 调用API
make hotreload-role
```

## 开发规范

### 添加新模块

1. **创建插件目录** (`plugin/mymodule/`)

```go
package main

import baseplugin "github.com/yourorg/wg_ai/plugin"

type MyLogic struct{}

func (l *MyLogic) Name() string { return "mymodule" }

func (l *MyLogic) Handle(ctx *baseplugin.LogicContext, method string, params map[string]any) (*baseplugin.LogicResult, error) {
    switch method {
    case "do_something":
        // 实现逻辑
        return baseplugin.Success(map[string]any{"result": "ok"}), nil
    }
    return nil, baseplugin.ErrMethodNotFound
}

var MyModule baseplugin.LogicModule = &MyLogic{}
```

2. **添加路由** (`config/routes.json`)

```json
{"msg_id": 4001, "module": "mymodule", "method": "do_something"}
```

3. **添加编译目标** (`Makefile`)

```makefile
plugin-mymodule:
	go build -buildmode=plugin -o plugins/mymodule.so ./plugin/mymodule
```

### 添加内置模块

如果模块性能敏感，不适合热更，可作为内置模块:

1. 在 `internal/` 下创建包
2. 实现 `LogicModule` 接口
3. 在 `game/server.go` 中注册:

```go
// 4.1 注册内置模块
myModule := mypackage.NewModule()
s.pluginMgr.RegisterModule("mymodule", myModule)
```

---

*最后更新: 2026-04-03*
