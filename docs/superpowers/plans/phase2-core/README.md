# Phase 2: Core Components - 核心组件

## 背景

Phase 1 完成了基础设施搭建，现在需要实现游戏服务器的核心组件。这些组件是服务器的骨架，处理网络连接、会话管理、消息分发等核心功能。

## 目标

- 实现协议编解码器（二进制格式）
- 实现 Session 会话管理
- 实现 Agent 玩家代理模型
- 实现消息分发器
- 实现 TCP 网关服务器

## 任务列表

| 任务 | 文件 | 描述 |
|------|------|------|
| Task 8 | `01-protocol.md` | 协议编解码 |
| Task 9 | `02-session.md` | Session 管理 |
| Task 10-12 | `03-agent.md` | Agent 模型、Manager、Dispatcher |
| Task 13-14 | `04-tcp-server.md` | TCP 服务器和连接处理 |

## 依赖关系

```
Phase 1 (必须完成)
   │
   ├── Task 8 (Protocol) ──────────────────┐
   │                                        │
   ├── Task 9 (Session)                     │
   │       │                                │
   │       └── Task 10-12 (Agent)           │
   │                    │                   │
   │                    └── Task 13-14 ─────┘
   │                         (TCP Server)
   │
   └── Task 8-14 都完成后 → Phase 3
```

## 架构说明

```
┌─────────────────────────────────────────┐
│              TCP Server (Gate)          │
│  ┌─────────────────────────────────┐   │
│  │         Connection Pool          │   │
│  └────────────────┬────────────────┘   │
│                   │                      │
│  ┌────────────────▼────────────────┐   │
│  │         Session Manager          │   │
│  └────────────────┬────────────────┘   │
│                   │                      │
│  ┌────────────────▼────────────────┐   │
│  │           Agent Pool             │   │
│  │   [Agent1] [Agent2] ... [AgentN] │   │
│  └─────────────────────────────────┘   │
└─────────────────────────────────────────┘
```

## 执行顺序

1. Task 8 (Protocol) - 独立，可先执行
2. Task 9 (Session) - 依赖 Phase 1
3. Task 10-12 (Agent) - 依赖 Task 9
4. Task 13-14 (TCP Server) - 依赖 Task 8, 9, 10-12

## 验证

本阶段完成后，运行：
```bash
go test ./internal/gate/... ./internal/agent/... ./internal/session/... -v
```

所有测试应通过。
