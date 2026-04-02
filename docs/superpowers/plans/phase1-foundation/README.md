# Phase 1: Foundation - 基础设施搭建

## 背景

本项目是将原 C/Lua (Skynet) 游戏服务器核心部分用 Go 重写，实现 game_server、login_server、db_server 三个核心服务。

Phase 1 是整个项目的基础设施搭建，后续所有 Phase 都依赖于本阶段创建的项目结构、配置系统、日志系统、错误系统和协议定义。

## 目标

- 建立标准 Go 项目结构
- 实现配置加载系统（使用 Viper）
- 实现日志系统（使用 Zap）
- 定义错误码体系
- 定义 Protobuf 协议（CS 和 SS）
- 生成 Go 协议代码

## 任务列表

| 任务 | 文件 | 描述 |
|------|------|------|
| Task 1 | `01-project-init.md` | 项目初始化 |
| Task 2 | `02-config-system.md` | 配置系统 |
| Task 3 | `03-logger.md` | 日志系统 |
| Task 4 | `04-errors.md` | 错误系统 |
| Task 5-7 | `05-protobuf.md` | Protobuf 定义与生成 |

## 依赖关系

```
Task 1 (项目初始化)
   └── Task 2 (配置系统)
         └── Task 3 (日志系统)
         └── Task 4 (错误系统)
         └── Task 5-7 (Protobuf)
```

## 执行顺序

按任务顺序依次执行。Task 3、4、5-7 可以并行执行（都依赖 Task 2，但彼此独立）。

## 验证

本阶段完成后，运行：
```bash
go test ./...
make proto
```

所有测试应通过，Protobuf 代码应成功生成。
