# Go Game Server Design Document

> 基于 wg_server (Skynet) 架构，使用 Go 重写 game + login + db 三个核心服务

## 1. Overview

### 1.1 Goals

- 将原 C/Lua (Skynet) 游戏服务器核心部分用 Go 重写
- 实现服务: game_server, login_server, db_server
- 保持与前端客户端的协议兼容性（改用Protobuf）
- 支持高并发连接 (10000+)

### 1.2 Non-Goals

- 不实现 battle_server, center_server, chat_server 等其他服务
- 不保留 Lua 脚本层，纯 Go 实现
- 不实现原 Skynet 的所有特性

### 1.3 Key Decisions

| 决策项 | 选择 | 原因 |
|--------|------|------|
| 协议格式 | Protobuf | 替换原Sproto，生态更成熟 |
| 并发模型 | Goroutine-per-conn | Go原生模式，简单高效 |
| 服务间通信 | gRPC | 成熟稳定，类型安全 |
| 数据库访问 | 原生SQL | 性能优先，控制精细 |
| 配置格式 | YAML | Go项目常用，可读性好 |
| 认证流程 | 简化Token | 去除原有多步握手 |

## 2. Architecture

### 2.1 System Architecture

```
                    ┌──────────────────────────────────────┐
                    │              Clients                  │
                    └─────────────────┬────────────────────┘
                                      │ TCP/Protobuf
                    ┌─────────────────▼────────────────────┐
                    │           Game Server                 │
                    │  ┌─────────────────────────────┐     │
                    │  │         Gate (TCP)          │     │
                    │  └─────────────┬───────────────┘     │
                    │                │                      │
                    │  ┌─────────────▼───────────────┐     │
                    │  │    Agent Pool (x100)        │     │
                    │  │  [Agent1] [Agent2] ...      │     │
                    │  └─────────────┬───────────────┘     │
                    │                │                      │
                    │  ┌─────────────▼───────────────┐     │
                    │  │      Session Manager        │     │
                    │  └─────────────┬───────────────┘     │
                    └────────────────┼────────────────────┘
                                     │ gRPC
              ┌──────────────────────┼──────────────────────┐
              │                      │                      │
    ┌─────────▼────────┐   ┌────────▼────────┐   ┌────────▼────────┐
    │   Login Server   │   │    DB Server    │   │     Redis       │
    │  (JWT Token)     │   │   (gRPC)        │   │                 │
    └──────────────────┘   └────────┬────────┘   └─────────────────┘
                                     │
                            ┌────────▼────────┐
                            │      MySQL      │
                            └─────────────────┘
```

### 2.2 Project Structure

```
wg_ai/
├── cmd/                        # 各服务入口
│   ├── game/main.go           # game_server入口
│   ├── login/main.go          # login_server入口
│   └── db/main.go             # db_server入口
├── internal/
│   ├── common/                # 公共代码
│   │   ├── config/            # 配置解析
│   │   ├── logger/            # 日志
│   │   ├── errors/            # 错误定义
│   │   └── utils/             # 工具函数
│   ├── gate/                  # 网关层
│   │   ├── tcp_server.go      # TCP监听
│   │   ├── connection.go      # 连接管理
│   │   └── protocol.go        # 协议解析
│   ├── agent/                 # 玩家代理
│   │   ├── agent.go           # Agent结构
│   │   ├── manager.go         # Agent池管理
│   │   └── dispatcher.go      # 消息分发
│   ├── session/               # 会话管理
│   │   ├── session.go         # 会话结构
│   │   └── manager.go         # 会话管理器
│   ├── rpc/                   # 服务间通信
│   │   ├── client.go          # gRPC客户端
│   │   └── server.go          # gRPC服务端
│   └── db/                    # 数据层
│       ├── mysql.go           # MySQL操作
│       └── redis.go           # Redis操作
├── proto/                     # Protobuf定义
│   ├── cs/                    # Client-Server协议
│   │   └── protocol.proto
│   └── ss/                    # Server-Server协议
│       └── rpc.proto
├── config/                    # 配置文件
│   ├── game.yaml
│   ├── login.yaml
│   └── db.yaml
├── go.mod
└── Makefile
```

## 3. Protocol Design

### 3.1 Network Packet Format

```
+----------------+----------------+------------------+
|  Length (4B)   |  MsgType (1B)  |  Protobuf Data   |
+----------------+----------------+------------------+

MsgType:
  0x01 = Request
  0x02 = Response
  0x03 = Push
  0x04 = Handshake
```

### 3.2 Client-Server Protocol (Protobuf)

```protobuf
// proto/cs/protocol.proto
syntax = "proto3";
package cs;

// 消息头
message Header {
    uint32 msg_id = 1;      // 消息ID
    uint32 sequence = 2;    // 序列号(用于响应匹配)
}

// 请求包装
message Request {
    Header header = 1;
    bytes payload = 2;      // 实际消息体
}

// 响应包装
message Response {
    Header header = 1;
    int32 code = 2;         // 错误码
    bytes payload = 3;
}

// 推送消息
message Push {
    uint32 msg_id = 1;
    bytes payload = 2;
}

// 登录请求
message LoginRequest {
    string token = 1;
}

message LoginResponse {
    int64 rid = 1;
    string name = 2;
}

// 心跳
message HeartbeatRequest {}
message HeartbeatResponse {}
```

### 3.3 Server-Server Protocol (gRPC)

```protobuf
// proto/ss/rpc.proto
syntax = "proto3";
package ss;

// Login -> Game: 通知玩家登录
service LoginService {
    rpc NotifyLogin(LoginNotifyRequest) returns (LoginNotifyResponse);
}

message LoginNotifyRequest {
    int64 uid = 1;
    string token = 2;
}

message LoginNotifyResponse {
    bool success = 1;
}

// Game -> DB: 数据存取
service DBService {
    rpc LoadRole(LoadRoleRequest) returns (LoadRoleResponse);
    rpc SaveRole(SaveRoleRequest) returns (SaveRoleResponse);
    rpc CreateUser(CreateUserRequest) returns (CreateUserResponse);
}

message LoadRoleRequest {
    int64 rid = 1;
}

message LoadRoleResponse {
    bytes data = 1;  // 序列化的角色数据
}

message SaveRoleRequest {
    int64 rid = 1;
    bytes data = 2;
}

message SaveRoleResponse {
    bool success = 1;
}

message CreateUserRequest {
    string username = 1;
    string password_hash = 2;
}

message CreateUserResponse {
    int64 uid = 1;
}
```

## 4. Data Model

### 4.1 Key Identifiers

| 标识符 | 说明 | 生成时机 |
|--------|------|---------|
| UID | 用户账号ID | 注册时由DB生成 |
| RID | 角色ID | 首次进入游戏时生成 |
| SessionID | 会话ID | 每次连接时生成 |

**关系**: 一个UID可以对应多个RID(不同服务器)，一个RID只能属于一个UID。

### 4.2 Player Structure

```go
// internal/agent/player.go
type Player struct {
    RID       int64
    UID       int64
    Name      string
    Level     int32
    Data      []byte  // 序列化的角色详细数据
    Session   *session.Session
    LastSave  time.Time
}
```

## 5. Component Design

### 4.1 Gate (TCP Server)

```go
// internal/gate/tcp_server.go
type TCPServer struct {
    listener   net.Listener
    agentMgr   *agent.Manager
    sessionMgr *session.Manager
    config     *config.GateConfig
}

func (s *TCPServer) Start(addr string) error {
    ln, err := net.Listen("tcp", addr)
    if err != nil {
        return err
    }
    s.listener = ln

    for {
        conn, err := ln.Accept()
        if err != nil {
            continue
        }
        go s.handleConnection(conn)
    }
}

func (s *TCPServer) handleConnection(conn net.Conn) {
    defer conn.Close()

    // 1. 设置读写超时
    conn.SetDeadline(time.Now().Add(s.config.ReadTimeout))

    // 2. 读取握手包
    handshake, err := s.readHandshake(conn)
    if err != nil {
        return
    }

    // 3. 验证Token
    claims, err := s.validateToken(handshake.Token)
    if err != nil {
        s.sendHandshakeResponse(conn, 401, "Unauthorized")
        return
    }

    // 4. 创建Session
    sess := s.sessionMgr.Create(claims.UID, conn)

    // 5. 分配Agent
    agent := s.agentMgr.Assign(sess)

    // 6. 发送握手成功
    s.sendHandshakeResponse(conn, 200, "OK")

    // 7. 进入消息循环
    s.messageLoop(conn, agent, sess)
}
```

### 4.2 Agent Model

```go
// internal/agent/agent.go
type Agent struct {
    ID       int
    players  map[int64]*Player
    msgQueue chan *Message
    dbClient *rpc.DBClient
}

func NewAgent(id int, queueSize int) *Agent {
    return &Agent{
        ID:       id,
        players:  make(map[int64]*Player),
        msgQueue: make(chan *Message, queueSize),
    }
}

func (a *Agent) Run() {
    for msg := range a.msgQueue {
        a.handleMessage(msg)
    }
}

func (a *Agent) handleMessage(msg *Message) {
    // 根据MsgID路由到对应处理函数
    handler, ok := handlers[msg.MsgID]
    if !ok {
        log.Warnf("unknown msg_id: %d", msg.MsgID)
        return
    }

    resp, err := handler(a, msg)
    if err != nil {
        a.sendError(msg, err)
        return
    }

    a.sendResponse(msg, resp)
}

// Message Router - 全局消息路由表
var handlers = map[uint16]HandlerFunc{
    1001: handleHeartbeat,
    1002: handleLogin,
    // 业务消息按模块划分: 2000-2999 角色, 3000-3999 背包, etc.
}

type HandlerFunc func(a *Agent, msg *Message) (proto.Message, error)

// Manager - Agent池管理
type Manager struct {
    agents     []*Agent
    roundRobin uint32
}

func NewManager(agentCount, queueSize int) *Manager {
    m := &Manager{
        agents: make([]*Agent, agentCount),
    }
    for i := 0; i < agentCount; i++ {
        agent := NewAgent(i, queueSize)
        m.agents[i] = agent
        go agent.Run()
    }
    return m
}

func (m *Manager) Assign(sess *session.Session) *Agent {
    idx := atomic.AddUint32(&m.roundRobin, 1) % uint32(len(m.agents))
    return m.agents[idx]
}
```

### 4.3 Session Management

```go
// internal/session/session.go
type Session struct {
    RID       int64
    UID       int64
    Conn      net.Conn
    Agent     *agent.Agent
    CreatedAt time.Time
    LastActive time.Time
    mutex     sync.RWMutex
}

func (s *Session) Send(data []byte) error {
    s.mutex.Lock()
    defer s.mutex.Unlock()

    s.Conn.SetWriteDeadline(time.Now().Add(5 * time.Second))
    _, err := s.Conn.Write(data)
    return err
}

// manager.go
type Manager struct {
    sessions map[int64]*Session  // rid -> Session
    mutex    sync.RWMutex
}

func (m *Manager) Create(uid int64, conn net.Conn) *Session {
    m.mutex.Lock()
    defer m.mutex.Unlock()

    rid := generateRID()
    sess := &Session{
        RID:        rid,
        UID:        uid,
        Conn:       conn,
        CreatedAt:  time.Now(),
        LastActive: time.Now(),
    }
    m.sessions[rid] = sess
    return sess
}

func (m *Manager) Get(rid int64) *Session {
    m.mutex.RLock()
    defer m.mutex.RUnlock()
    return m.sessions[rid]
}

func (m *Manager) Remove(rid int64) {
    m.mutex.Lock()
    defer m.mutex.Unlock()
    delete(m.sessions, rid)
}
```

## 5. Service Design

### 5.1 Login Server

```go
// cmd/login/main.go
type LoginServer struct {
    dbClient    *rpc.DBClient
    tokenSecret []byte
    gameAddrs   []string
}

func (s *LoginServer) HandleLogin(req *LoginRequest) (*LoginResponse, error) {
    // 1. 验证账号密码
    user, err := s.dbClient.GetUserByUsername(context.Background(), req.Username)
    if err != nil {
        return nil, errors.New("user not found")
    }

    if !verifyPassword(req.Password, user.PasswordHash) {
        return nil, errors.New("invalid password")
    }

    // 2. 生成JWT Token
    token := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{
        "uid": user.ID,
        "exp": time.Now().Add(24 * time.Hour).Unix(),
    })
    tokenString, err := token.SignedString(s.tokenSecret)
    if err != nil {
        return nil, err
    }

    // 3. 选择Game服务器
    gameAddr := s.selectGameServer()

    return &LoginResponse{
        Token:    tokenString,
        GameAddr: gameAddr,
    }, nil
}

func (s *LoginServer) selectGameServer() string {
    // 简单轮询，可替换为负载均衡
    idx := atomic.AddUint32(&s.roundRobin, 1)
    return s.gameAddrs[idx%uint32(len(s.gameAddrs))]
}
```

### 5.2 DB Server

```go
// cmd/db/main.go
type DBServer struct {
    mysql *sql.DB
    redis *redis.Client
}

func (s *DBServer) LoadRole(ctx context.Context, req *ss.LoadRoleRequest) (*ss.LoadRoleResponse, error) {
    cacheKey := fmt.Sprintf("role:%d", req.Rid)

    // 1. 查Redis缓存
    cached, err := s.redis.Get(ctx, cacheKey).Bytes()
    if err == nil {
        return &ss.LoadRoleResponse{Data: cached}, nil
    }

    // 2. 查MySQL
    var data []byte
    err = s.mysql.QueryRowContext(ctx,
        "SELECT data FROM role WHERE rid = ?", req.Rid).Scan(&data)
    if err != nil {
        return nil, err
    }

    // 3. 写入缓存
    s.redis.Set(ctx, cacheKey, data, 5*time.Minute)

    return &ss.LoadRoleResponse{Data: data}, nil
}

func (s *DBServer) SaveRole(ctx context.Context, req *ss.SaveRoleRequest) (*ss.SaveRoleResponse, error) {
    // 1. 写MySQL
    _, err := s.mysql.ExecContext(ctx,
        "UPDATE role SET data = ?, updated_at = NOW() WHERE rid = ?",
        req.Data, req.Rid)
    if err != nil {
        return nil, err
    }

    // 2. 更新缓存
    cacheKey := fmt.Sprintf("role:%d", req.Rid)
    s.redis.Set(ctx, cacheKey, req.Data, 5*time.Minute)

    return &ss.SaveRoleResponse{Success: true}, nil
}
```

## 6. Error Handling

### 6.1 Error Codes

```go
// internal/common/errors/codes.go
const (
    // 通用错误 1-99
    CodeSuccess       = 0
    CodeUnknown       = 1
    CodeInvalidParam  = 2
    CodeTimeout       = 3

    // 登录相关 100-199
    CodeInvalidToken     = 100
    CodeTokenExpired     = 101
    CodeAccountNotFound  = 102
    CodePasswordWrong    = 103

    // 游戏逻辑 200-299
    CodeNotEnoughResource = 200
    CodeInvalidOperation  = 201
    CodePlayerNotFound    = 202
)

type GameError struct {
    Code    int
    Message string
}

func (e *GameError) Error() string {
    return fmt.Sprintf("[%d] %s", e.Code, e.Message)
}

func NewGameError(code int, msg string) *GameError {
    return &GameError{Code: code, Message: msg}
}
```

## 7. Configuration

### 7.1 Game Server Config

```yaml
# config/game.yaml
server:
  id: 1
  name: "game1"
  host: "0.0.0.0"
  port: 44445
  max_conn: 10000

gate:
  read_timeout: 30s
  write_timeout: 30s
  msg_queue_size: 1000

agent:
  count: 100
  players_per_agent: 100

cluster:
  login_addr: "127.0.0.1:50051"
  db_addr: "127.0.0.1:50052"

database:
  mysql:
    host: "127.0.0.1"
    port: 3306
    database: "game"
    username: "root"
    password: "xxx"
    max_open: 100
    max_idle: 20
  redis:
    host: "127.0.0.1"
    port: 6379
    db: 0
    pool_size: 100

log:
  level: "info"
  output: "./logs/game.log"
```

### 7.2 Login Server Config

```yaml
# config/login.yaml
server:
  id: 1
  grpc_port: 50051

token:
  secret: "your-secret-key"
  expire: "24h"

game_servers:
  - addr: "127.0.0.1:44445"
    weight: 1

database:
  mysql:
    host: "127.0.0.1"
    port: 3306
    database: "game"
    username: "root"
    password: "xxx"
```

## 8. Testing Strategy

### 8.1 Test Layers

```
├── Unit Tests
│   ├── Protocol encode/decode
│   ├── Message dispatch
│   └── Utility functions
├── Integration Tests
│   ├── gRPC services (mock DB)
│   ├── Agent message handling
│   └── Session management
└── Stress Tests
    ├── Connection count (10000+)
    └── Message throughput
```

### 8.2 Example Test

```go
// internal/gate/tcp_server_test.go
func TestHandshake(t *testing.T) {
    // 1. 创建mock token
    token := createTestToken(12345)

    // 2. 模拟连接
    client, server := net.Pipe()
    defer client.Close()
    defer server.Close()

    // 3. 发送握手包
    handshake := &cs.Handshake{Token: token}
    data, _ := proto.Marshal(handshake)
    packet := makePacket(MsgTypeHandshake, data)
    client.Write(packet)

    // 4. 验证响应
    resp := readResponse(client)
    assert.Equal(t, 200, resp.Code)
}
```

## 9. Dependencies

```go
// go.mod
module github.com/yourorg/wg_ai

go 1.21

require (
    github.com/spf13/viper v1.18.0
    go.uber.org/zap v1.26.0
    google.golang.org/grpc v1.60.0
    google.golang.org/protobuf v1.32.0
    github.com/go-sql-driver/mysql v1.7.1
    github.com/redis/go-redis/v9 v9.3.0
    github.com/golang-jwt/jwt/v5 v5.2.0
    github.com/panjf2000/gnet v2.3.3+incompatible  // 可选：高性能网络库
)
```

## 10. Implementation Phases

### Phase 1: Foundation
- 项目骨架搭建
- 配置系统
- 日志系统
- Protobuf定义

### Phase 2: Core Components
- TCP网关
- 协议编解码
- Agent模型
- Session管理

### Phase 3: Services
- DB Server (gRPC + MySQL/Redis)
- Login Server (JWT认证)
- Game Server (整合网关+Agent)

### Phase 4: Integration
- 服务间通信
- 完整登录流程
- 消息收发测试

### Phase 5: Polish
- 压力测试
- 性能优化
- 文档完善
