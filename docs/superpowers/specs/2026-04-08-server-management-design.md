# 服务统一启停脚本与有序停机设计

## 目标

为当前项目增加一个 `scripts/server.sh` 管理脚本，实现：

- 一键启动完整服务栈
- 一键关闭完整服务栈
- 按依赖顺序启动与停止服务
- 在关闭 `db` 服务前，确保业务侧 `game` 服务已经完成落盘并退出

本设计仅覆盖当前单机开发/测试形态下的三个核心进程：

- `db`
- `login`
- `game`

## 背景与现状

当前仓库已有三个入口：

- `cmd/db/main.go`
- `cmd/login/main.go`
- `cmd/game/main.go`

现状问题：

1. 缺少统一管理脚本，启动和停止需要手工操作多个进程。
2. `game` 服务当前 `Stop()` 只关闭网络入口、后台组件和 RPC 连接，没有显式“停机前落盘”阶段。
3. 如果先停 `db`，再停 `game`，则 `game` 的运行期脏数据可能无法正常写入。
4. 当前没有统一的 PID、日志、状态检查与失败回滚机制。

因此，不能只做一个简单 shell 脚本，还需要补齐 `game` 的优雅停机语义。

## 方案概览

采用“两层实现”的方案：

1. **外层脚本编排**：新增 `scripts/server.sh`，负责统一 start/stop/restart/status。
2. **服务内优雅停机**：增强 `internal/game/server.go`，使 `game` 在退出前完成明确的停止入口、停止业务循环、落盘、再断开 DB RPC。

这样脚本层只要保证先停 `game` 并等待其退出，就可以把“业务已落盘”的保证交给 `game` 进程内部实现，最后再安全关闭 `db`。

## 设计一：管理脚本 `scripts/server.sh`

### 支持命令

脚本提供以下命令：

- `start`：启动全部服务
- `stop`：停止全部服务
- `restart`：先停再启
- `status`：查看服务状态
- 可选扩展：`logs`（本次不必实现）

本次至少实现：`start`、`stop`、`restart`、`status`。

### 运行目录约定

脚本统一管理运行时目录：

- `runtime/pid/db.pid`
- `runtime/pid/login.pid`
- `runtime/pid/game.pid`
- `runtime/logs/db.log`
- `runtime/logs/login.log`
- `runtime/logs/game.log`

脚本启动时若目录不存在则自动创建。

### 启动顺序

严格按依赖顺序启动：

1. `db`
2. `login`
3. `game`

原因：

- `game` 依赖 `db` RPC
- `game` 配置中也依赖 `login` RPC
- `db` 无上游依赖，必须先就绪

### 停止顺序

严格按反向依赖顺序停止：

1. `game`
2. `login`
3. `db`

其中关键点是：

- `game` 必须先收到 `SIGTERM`
- 脚本必须等待 `game` 进程真正退出
- 只有在 `game` 完成优雅停机后，才继续关闭 `login` 和 `db`

### PID 与脏状态处理

脚本读取 PID 文件判断服务是否在运行。

规则：

- 若 PID 文件存在且进程存在，则视为运行中
- 若 PID 文件存在但进程已不存在，则清理脏 PID 文件
- `start` 遇到已运行进程时跳过，不重复拉起
- `stop` 遇到未运行进程时跳过，但清理残留 PID 文件

### 就绪检查与失败回滚

`start` 不仅拉起进程，还要等待服务真正就绪。

建议采用“PID 存活 + 端口监听”双重检查。端口可直接取配置中的固定值：

- `db`: `50052`
- `login`: `50051`
- `game`: `44445`
- `game ws`: `44446`（如果启用）

启动策略：

1. 启动 `db`，等待 gRPC 端口监听
2. 启动 `login`，等待 gRPC 端口监听
3. 启动 `game`，等待 TCP 端口监听，若配置启用 WS，再等待 WS 端口监听

失败回滚：

- `db` 启动失败：直接报错退出
- `login` 启动失败：停止已启动的 `db`
- `game` 启动失败：停止已启动的 `login` 和 `db`

### 停止超时策略

`stop` 使用两阶段停止：

1. 先发 `SIGTERM`
2. 在超时窗口内轮询等待进程退出
3. 超时后发 `SIGKILL`

建议超时：

- `game`: 30 秒（因为需要落盘）
- `login`: 10 秒
- `db`: 10 秒

脚本输出必须明确指出：

- 正在停止哪个服务
- 是否正常退出
- 是否触发超时强杀

### 状态检查

`status` 至少输出：

- 服务名
- PID 文件中的 PID
- 进程是否存活
- 对应端口是否监听

若 PID 和端口状态不一致，也要打印异常提示，便于排查。

## 设计二：`game` 服务的优雅停机

为了满足“关闭时先落盘再停 db”，必须增强 `internal/game/server.go` 的关闭流程。

### 新的停机阶段

`game` 的 `Stop()` 不再只是简单关闭组件，而是分阶段执行：

1. **进入 stopping 状态**
   - 标记服务器进入停机中
   - 后续新连接和新请求应尽量拒绝或不再进入新业务处理

2. **停止入口层**
   - 关闭 admin HTTP 服务
   - 停止 plugin watcher
   - 停止 TCP server
   - 停止 WebSocket server

3. **停止后台业务循环**
   - 停止 march manager
   - 停止 agent manager
   - 停止其他后续发现的后台协程入口

4. **执行落盘**
   - 遍历当前 `dataStore` 已加载玩家数据
   - 仅对脏数据执行持久化
   - 所有落盘完成后才允许继续后续停机

5. **关闭 RPC**
   - 最后关闭 `rpcClient`

这保证了：`game` 仍能在落盘阶段访问 `db` 服务，因为 RPC 连接和 `db` 服务都还活着。

### 落盘实现边界

当前 `game` 服务内存里使用的是 `data.PlayerStore`，而 `internal/data/persist.go` 的 SQL 落盘逻辑存在于 `db` 侧代码语义里；`game` 到 `db` 的持久化通路当前主要是通过 RPC 的 `SaveRole` / `LoadRole`。

因此本次设计建议：

- 在 `game` 内新增一个“flush loaded players”能力
- 它不直接写 MySQL，而是复用现有 DB RPC 通路
- 将每个脏玩家内存数据序列化后，通过 `rpcClient.SaveRole(...)` 逐个写入 `db`

这样不破坏现有架构边界：

- `game` 只负责组织内存数据
- `db` 仍是唯一数据写入服务

### PlayerStore 需要补的能力

当前 `PlayerStore` 只有 `GetPlayer`/`SetField`/`SetArray` 等访问方法，没有遍历或批量落盘辅助能力。

建议增加最小接口：

- 获取当前所有已加载玩家快照
- 或直接提供一个遍历接口

例如：

- `SnapshotPlayers() map[int64]*PlayerData`
- 或 `ForEachLoadedPlayer(func(rid int64, p *PlayerData) error) error`

推荐使用遍历接口，避免直接暴露内部 map。

### 序列化与脏标记处理

落盘流程建议：

1. 从 `PlayerStore` 遍历已加载玩家
2. 判断 `Dirty` 标记
3. 使用已有序列化逻辑把玩家内存数据转成字节
4. 通过 `rpcClient.SaveRole(ctx, rid, data)` 写入 DB
5. 成功后清除 `Dirty`
6. 任一写入失败则记录错误，并向上返回停机失败日志

停机阶段的原则是：

- **尽量全部写完再退出**
- 若个别玩家写入失败，应打印明确错误
- 仍然允许进程继续退出，但日志必须能指出数据未完全落盘

本次不引入重试队列、事务编排或增量 checkpoint。

## 设计三：入口层配合停机状态

为了避免停机过程中还接收新业务请求，建议 `game` 增加轻量 stopping 标记。

目标：

- `Stop()` 开始后不再接受新连接
- 已有连接尽快结束
- 避免一边 flush、一边继续写脏数据

本次只要求做到“尽量阻止新流量进入”，不要求构建复杂 draining 协议。

可接受的最小实现是：

- 先停 TCP / WS 监听器
- 再停 agent/业务循环
- 然后进行 flush

只要时序正确，就已经比当前状态明显安全。

## 测试与验证

### 脚本验证

至少验证以下场景：

1. `start`
   - 三个服务按顺序启动
   - PID 文件写入正确
   - 端口检查通过

2. `status`
   - 可显示三服务状态
   - 能识别脏 PID 文件

3. `stop`
   - 按 `game -> login -> db` 顺序停止
   - 进程退出后 PID 文件被移除

4. `restart`
   - 等价于 `stop` 后再 `start`

5. 启动失败回滚
   - 人为制造 `login` 或 `game` 启动失败，确认下游不会残留半启动进程

### 业务落盘验证

至少验证以下场景：

1. 启动服务后，令 `game` 产生脏玩家数据
2. 执行 `scripts/server.sh stop`
3. 观察 `game` 停机日志，确认进入 flush 阶段
4. 观察 `db` 在 `game` 退出后才关闭
5. 重启服务后，确认停机前数据已可重新读出

## 不在本次范围内

本次明确不做：

- systemd 单元文件
- Docker / docker-compose 编排
- 自动拉起守护进程
- 崩溃自动恢复
- 多实例/集群启停
- 复杂健康探针系统
- 通用进程管理平台

## 推荐实施顺序

1. 给 `PlayerStore` 增加已加载玩家遍历能力
2. 给 `game.Server` 增加 flush 逻辑和分阶段停机顺序
3. 确保 `rpcClient` 在 flush 完成前不关闭
4. 实现 `scripts/server.sh`
5. 加入启动就绪检查、停止超时、失败回滚
6. 做一轮手工集成验证

## 决策总结

本方案的核心决策是：

- 用 `scripts/server.sh` 做统一启停入口
- 用反向依赖顺序保证停止顺序
- 用 `game` 进程内部 flush 语义保证“先落盘再停 db”
- 不引入额外进程管理系统，保持最小可用实现

这样既能满足你的一键管理需求，也能避免“脚本看起来有顺序，但实际上数据还没落完”的伪优雅停机问题。
