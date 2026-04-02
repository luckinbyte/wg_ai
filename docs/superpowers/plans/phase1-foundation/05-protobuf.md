# Task 5-7: Protobuf Protocol Definition - 协议定义与生成

## 背景与目标

定义客户端-服务器(CS)和服务器-服务器(SS)的通信协议，使用 Protobuf 作为序列化格式。

**为什么需要这个任务：**
- Protobuf 比 JSON 更高效，适合游戏服务器
- 明确的协议定义方便前后端协作
- gRPC 依赖 Protobuf 生成代码

**输出：**
- `proto/cs/protocol.proto` - 客户端协议
- `proto/ss/rpc.proto` - 服务间 RPC 协议
- 生成的 Go 代码

## 依赖

- Task 1: 项目结构已创建（proto/ 目录存在）

## 步骤

### Step 1: Create CS protocol definition

Create `proto/cs/protocol.proto`:

```protobuf
syntax = "proto3";
package cs;
option go_package = "github.com/yourorg/wg_ai/proto/cs";

// 消息头
message Header {
    uint32 msg_id = 1;
    uint32 sequence = 2;
}

// 请求包装
message Request {
    Header header = 1;
    bytes payload = 2;
}

// 响应包装
message Response {
    Header header = 1;
    int32 code = 2;
    bytes payload = 3;
}

// 推送消息
message Push {
    uint32 msg_id = 1;
    bytes payload = 2;
}

// 握手
message Handshake {
    string token = 1;
}

message HandshakeResponse {
    int32 code = 1;
    string message = 2;
}

// 登录
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

// 消息ID定义
enum MsgID {
    MSG_ID_NONE = 0;
    MSG_ID_HANDSHAKE = 1;
    MSG_ID_LOGIN = 1001;
    MSG_ID_HEARTBEAT = 1002;
}
```

### Step 2: Create SS protocol definition

Create `proto/ss/rpc.proto`:

```protobuf
syntax = "proto3";
package ss;
option go_package = "github.com/yourorg/wg_ai/proto/ss";

// Login Service
service LoginService {
    rpc NotifyLogin(LoginNotifyRequest) returns (LoginNotifyResponse);
    rpc ValidateToken(ValidateTokenRequest) returns (ValidateTokenResponse);
}

message LoginNotifyRequest {
    int64 uid = 1;
    string token = 2;
}

message LoginNotifyResponse {
    bool success = 1;
}

message ValidateTokenRequest {
    string token = 1;
}

message ValidateTokenResponse {
    int64 uid = 1;
    bool valid = 2;
}

// DB Service
service DBService {
    rpc LoadRole(LoadRoleRequest) returns (LoadRoleResponse);
    rpc SaveRole(SaveRoleRequest) returns (SaveRoleResponse);
    rpc CreateUser(CreateUserRequest) returns (CreateUserResponse);
    rpc GetUser(GetUserRequest) returns (GetUserResponse);
}

message LoadRoleRequest {
    int64 rid = 1;
}

message LoadRoleResponse {
    bytes data = 1;
    bool found = 2;
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

message GetUserRequest {
    string username = 1;
}

message GetUserResponse {
    int64 uid = 1;
    string password_hash = 2;
    bool found = 3;
}
```

### Step 3: Install protoc and Go plugins

```bash
# Install protoc compiler
apt-get update && apt-get install -y protobuf-compiler

# Install Go protoc plugins
go install google.golang.org/protobuf/cmd/protoc-gen-go@latest
go install google.golang.org/grpc/cmd/protoc-gen-go-grpc@latest

# Ensure plugins are in PATH
export PATH=$PATH:$(go env GOPATH)/bin
```

### Step 4: Generate Go code

```bash
cd /root/ai_project/wg_ai

protoc --go_out=. --go_opt=paths=source_relative \
  --go-grpc_out=. --go-grpc_opt=paths=source_relative \
  proto/cs/*.proto proto/ss/*.proto
```

Expected output: Creates
- `proto/cs/protocol.pb.go`
- `proto/ss/rpc.pb.go`
- `proto/ss/rpc_grpc.pb.go`

### Step 5: Add dependencies

```bash
cd /root/ai_project/wg_ai
go get google.golang.org/grpc
go get google.golang.org/protobuf
```

### Step 6: Verify generated files

```bash
ls -la /root/ai_project/wg_ai/proto/cs/
ls -la /root/ai_project/wg_ai/proto/ss/
```

Expected: See `.pb.go` and `_grpc.pb.go` files

### Step 7: Commit

```bash
git add .
git commit -m "feat: add protobuf protocol definitions and generated code"
```

## 验证

```bash
cd /root/ai_project/wg_ai
go build ./proto/...
```

Expected: 编译成功，无错误

## 完成标志

- [ ] proto/cs/protocol.proto 存在
- [ ] proto/ss/rpc.proto 存在
- [ ] 生成的 .pb.go 文件存在
- [ ] go build ./proto/... 成功
- [ ] Commit 完成
