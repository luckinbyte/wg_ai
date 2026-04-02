# Go Game Server Implementation Plan Index

## 概览

本计划将原 C/Lua (Skynet) 游戏服务器用 Go 重写，实现 game_server、login_server、db_server 三个核心服务。

**架构：** Gate-Agent-Session 分层，gRPC 服务间通信，Protobuf 协议

**技术栈：** Go 1.21+, gRPC, Protobuf, MySQL, Redis, JWT

---

## Phase 概览

| Phase | 任务数 | 描述 | 产出 |
|-------|--------|------|------|
| [Phase 1](phase1-foundation/) | 5 | 基础设施 | 项目结构、配置、日志、错误、协议 |
| [Phase 2](phase2-core/) | 4 | 核心组件 | Protocol、Session、Agent、TCP Server |
| [Phase 3](phase3-services/) | 5 | 服务层 | DB Layer、gRPC、DB/Login/Game Server |
| [Phase 4](phase4-integration/) | 3 | 集成 | 组件集成、心跳、测试 |
| [Phase 5](phase5-polish/) | 4 | 完善 | 优雅关闭、健康检查、指标 |

**总计：21 个任务文档，28 个实现步骤**

---

## 执行顺序

```
Phase 1 (Foundation)
    │
    ├── Task 1: 项目初始化
    │       │
    │       ├── Task 2: 配置系统
    │       │       │
    │       │       ├── Task 3: 日志系统 ────────┐
    │       │       ├── Task 4: 错误系统 ────────┤ 可并行
    │       │       └── Task 5-7: Protobuf ──────┘
    │       │
    ↓
Phase 2 (Core Components)
    │
    ├── Task 8: Protocol Codec ──────────────────┐
    │                                             │
    ├── Task 9: Session Management ──────────────┤
    │       │                                     │
    │       └── Task 10-12: Agent Model ─────────┤
    │                    │                        │
    │                    └── Task 13-14: TCP ─────┘
    │
    ↓
Phase 3 (Services)
    │
    ├── Task 15-16: DB Layer
    │       │
    │       └── Task 17-18: gRPC
    │               │
    │               ├── Task 19: DB Server
    │               │
    │               └── Task 20: Login Server
    │                       │
    │                       └── Task 21: Game Server
    │
    ↓
Phase 4 (Integration)
    │
    ├── Task 22: 组件集成
    │       │
    │       ├── Task 23: 心跳处理
    │       │
    │       └── Task 24: 集成测试
    │
    ↓
Phase 5 (Polish)
    │
    ├── Task 25: 优雅关闭
    │
    ├── Task 26: 健康检查
    │
    ├── Task 27: 指标收集
    │
    └── Task 28: 最终构建
```

---

## 目录结构

```
plans/
├── INDEX.md                    # 本文件
├── phase1-foundation/
│   ├── README.md
│   ├── 01-project-init.md
│   ├── 02-config-system.md
│   ├── 03-logger.md
│   ├── 04-errors.md
│   └── 05-protobuf.md
├── phase2-core/
│   ├── README.md
│   ├── 01-protocol.md
│   ├── 02-session.md
│   ├── 03-agent.md
│   └── 04-tcp-server.md
├── phase3-services/
│   ├── README.md
│   ├── 01-db-layer.md
│   ├── 02-rpc.md
│   ├── 03-db-server.md
│   ├── 04-login-server.md
│   └── 05-game-server.md
├── phase4-integration/
│   ├── README.md
│   ├── 01-integrate.md
│   ├── 02-heartbeat.md
│   └── 03-integration-test.md
└── phase5-polish/
    ├── README.md
    ├── 01-graceful.md
    ├── 02-health.md
    ├── 03-metrics.md
    └── 04-final.md
```

---

## Agent 执行指南

每个任务文档都包含：

1. **背景与目标** - 为什么做这个任务
2. **依赖** - 前置任务
3. **步骤** - 具体实现（含完整代码）
4. **验证** - 如何验证完成
5. **完成标志** - Checklist

执行时：
- 按目录顺序执行
- 每个 Phase 的 README.md 包含该 Phase 的概览
- 完成一个任务后再执行下一个

---

## 验证命令

```bash
# 运行所有单元测试
go test ./...

# 运行集成测试
go test -tags=integration ./tests/...

# 构建所有服务
make build

# 验证构建
ls -la bin/
```
