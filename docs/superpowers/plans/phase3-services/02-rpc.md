# Task 17-18: gRPC Layer - gRPC 通信层

## 背景与目标

实现 gRPC 客户端和服务端，用于服务间通信。

**输出：**
- `internal/rpc/client.go` - gRPC 客户端
- `internal/rpc/server.go` - gRPC 服务端

## 依赖

- Task 5-7: Protobuf 代码已生成
- Task 15-16: DB Layer（可选，用于类型引用）

## 步骤

### Step 1: Write test

Create `internal/rpc/client_test.go`:

```go
package rpc

import (
	"testing"
)

func TestNewClient(t *testing.T) {
	cfg := &ClientConfig{
		DBAddr: "localhost:50052",
	}
	
	client := NewClient(cfg)
	if client == nil {
		t.Error("client should not be nil")
	}
}
```

Create `internal/rpc/server_test.go`:

```go
package rpc

import (
	"testing"
)

func TestNewServer(t *testing.T) {
	srv := NewServer(":50052")
	if srv == nil {
		t.Error("server should not be nil")
	}
}
```

### Step 2: Write client implementation

Create `internal/rpc/client.go`:

```go
package rpc

import (
	"context"

	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"

	ss "github.com/yourorg/wg_ai/proto/ss"
)

type ClientConfig struct {
	DBAddr    string
	LoginAddr string
}

type Client struct {
	dbConn   *grpc.ClientConn
	dbClient ss.DBServiceClient
}

func NewClient(cfg *ClientConfig) *Client {
	return &Client{}
}

func (c *Client) ConnectDB(addr string) error {
	conn, err := grpc.Dial(addr,
		grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		return err
	}
	c.dbConn = conn
	c.dbClient = ss.NewDBServiceClient(conn)
	return nil
}

func (c *Client) Close() {
	if c.dbConn != nil {
		c.dbConn.Close()
	}
}

func (c *Client) LoadRole(ctx context.Context, rid int64) ([]byte, bool, error) {
	resp, err := c.dbClient.LoadRole(ctx, &ss.LoadRoleRequest{Rid: rid})
	if err != nil {
		return nil, false, err
	}
	return resp.Data, resp.Found, nil
}

func (c *Client) SaveRole(ctx context.Context, rid int64, data []byte) error {
	_, err := c.dbClient.SaveRole(ctx, &ss.SaveRoleRequest{Rid: rid, Data: data})
	return err
}

func (c *Client) CreateUser(ctx context.Context, username, passwordHash string) (int64, error) {
	resp, err := c.dbClient.CreateUser(ctx, &ss.CreateUserRequest{
		Username:     username,
		PasswordHash: passwordHash,
	})
	if err != nil {
		return 0, err
	}
	return resp.Uid, nil
}

func (c *Client) GetUser(ctx context.Context, username string) (int64, string, bool, error) {
	resp, err := c.dbClient.GetUser(ctx, &ss.GetUserRequest{Username: username})
	if err != nil {
		return 0, "", false, err
	}
	return resp.Uid, resp.PasswordHash, resp.Found, nil
}
```

### Step 3: Write server implementation

Create `internal/rpc/server.go`:

```go
package rpc

import (
	"net"

	"google.golang.org/grpc"
)

type Server struct {
	addr   string
	server *grpc.Server
}

func NewServer(addr string) *Server {
	return &Server{
		addr:   addr,
		server: grpc.NewServer(),
	}
}

func (s *Server) Start() error {
	lis, err := net.Listen("tcp", s.addr)
	if err != nil {
		return err
	}
	return s.server.Serve(lis)
}

func (s *Server) Stop() {
	s.server.GracefulStop()
}

func (s *Server) GRPCServer() *grpc.Server {
	return s.server
}
```

### Step 4: Test

```bash
cd /root/ai_project/wg_ai
go test ./internal/rpc/...
```

Expected: PASS

### Step 5: Commit

```bash
git add .
git commit -m "feat: add gRPC client and server"
```

## 完成标志

- [ ] Client 测试通过
- [ ] Server 测试通过
- [ ] Commit 完成
