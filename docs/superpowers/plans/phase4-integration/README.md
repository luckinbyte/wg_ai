# Phase 4: Integration - 集成测试

## 背景

Phase 3 完成了三个服务的实现，现在需要将它们集成在一起，确保组件间协作正常。

## 目标

- 集成 TCP Server 与 Session/Agent
- 实现心跳处理
- 编写集成测试

## 任务列表

| 任务 | 文件 | 描述 |
|------|------|------|
| Task 22 | `01-integrate.md` | 组件集成 |
| Task 23 | `02-heartbeat.md` | 心跳处理 |
| Task 24 | `03-integration-test.md` | 集成测试 |

## 依赖

- Phase 1-3 全部完成

## 验证

```bash
# 运行集成测试
go test -tags=integration ./tests/...
```
