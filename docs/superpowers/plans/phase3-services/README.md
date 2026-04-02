# Phase 3: Services - 服务层

## 背景

Phase 2 完成了核心组件，现在需要实现三个独立的微服务：DB Server、Login Server、Game Server。

## 目标

- 实现 DB Server（gRPC + MySQL/Redis）
- 实现 Login Server（JWT 认证）
- 实现 Game Server（整合所有组件）

## 任务列表

| 任务 | 文件 | 描述 |
|------|------|------|
| Task 15-16 | `01-db-layer.md` | MySQL/Redis 数据层 |
| Task 17-18 | `02-rpc.md` | gRPC 客户端/服务端 |
| Task 19 | `03-db-server.md` | DB 服务入口 |
| Task 20 | `04-login-server.md` | Login 服务入口 |
| Task 21 | `05-game-server.md` | Game 服务入口 |

## 依赖关系

```
Phase 1-2 (必须完成)
   │
   ├── Task 15-16 (DB Layer)
   │       │
   │       └── Task 17-18 (RPC) ──┬── Task 19 (DB Server)
   │                              │
   │                              └── Task 20 (Login Server)
   │                                      │
   └──────────────────────────────────────┴── Task 21 (Game Server)
```

## 执行顺序

1. Task 15-16 (DB Layer) - 数据库访问层
2. Task 17-18 (RPC) - gRPC 通信层
3. Task 19 (DB Server) - 可独立运行
4. Task 20 (Login Server) - 依赖 DB Server
5. Task 21 (Game Server) - 整合所有服务

## 验证

```bash
# 启动 DB Server
./bin/db -config config/db.yaml

# 启动 Login Server
./bin/login -config config/login.yaml

# 启动 Game Server
./bin/game -config config/game.yaml
```

三个服务应能正常启动。
