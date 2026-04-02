# Task 19: DB Server - 数据服务

## 背景与目标

实现 DB Server 入口，提供 gRPC 接口访问 MySQL/Redis。

**输出：**
- `internal/db/service.go` - DB 服务实现
- `cmd/db/main.go` - 服务入口
- `config/db.yaml` - 配置文件

## 依赖

- Task 15-16: DB Layer
- Task 17-18: gRPC Server
- Task 5-7: Protobuf (ss.DBService)

## 步骤

### Step 1: Create DB service

Create `internal/db/service.go`:

```go
package db

import (
	"context"
	"fmt"

	ss "github.com/yourorg/wg_ai/proto/ss"
)

type DBService struct {
	ss.UnimplementedDBServiceServer
	mysql *MySQL
	redis *Redis
}

func NewDBService(mysql *MySQL, redis *Redis) *DBService {
	return &DBService{
		mysql: mysql,
		redis: redis,
	}
}

func (s *DBService) LoadRole(ctx context.Context, req *ss.LoadRoleRequest) (*ss.LoadRoleResponse, error) {
	cacheKey := fmt.Sprintf("role:%d", req.Rid)

	// Try Redis first
	cached, err := s.redis.Get(ctx, cacheKey)
	if err == nil {
		return &ss.LoadRoleResponse{Data: cached, Found: true}, nil
	}

	// Query MySQL
	var data []byte
	err = s.mysql.DB().QueryRowContext(ctx,
		"SELECT data FROM role WHERE rid = ?", req.Rid).Scan(&data)
	if err != nil {
		return &ss.LoadRoleResponse{Found: false}, nil
	}

	// Cache it
	_ = s.redis.Set(ctx, cacheKey, data, 5*time.Minute)

	return &ss.LoadRoleResponse{Data: data, Found: true}, nil
}

func (s *DBService) SaveRole(ctx context.Context, req *ss.SaveRoleRequest) (*ss.SaveRoleResponse, error) {
	_, err := s.mysql.DB().ExecContext(ctx,
		"INSERT INTO role (rid, data, updated_at) VALUES (?, ?, NOW()) "+
			"ON DUPLICATE KEY UPDATE data = ?, updated_at = NOW()",
		req.Rid, req.Data, req.Data)
	if err != nil {
		return &ss.SaveRoleResponse{Success: false}, nil
	}

	cacheKey := fmt.Sprintf("role:%d", req.Rid)
	_ = s.redis.Set(ctx, cacheKey, req.Data, 5*time.Minute)

	return &ss.SaveRoleResponse{Success: true}, nil
}

func (s *DBService) CreateUser(ctx context.Context, req *ss.CreateUserRequest) (*ss.CreateUserResponse, error) {
	result, err := s.mysql.DB().ExecContext(ctx,
		"INSERT INTO user (username, password_hash, created_at) VALUES (?, ?, NOW())",
		req.Username, req.PasswordHash)
	if err != nil {
		return &ss.CreateUserResponse{Uid: 0}, nil
	}

	uid, _ := result.LastInsertId()
	return &ss.CreateUserResponse{Uid: uid}, nil
}

func (s *DBService) GetUser(ctx context.Context, req *ss.GetUserRequest) (*ss.GetUserResponse, error) {
	var uid int64
	var passwordHash string
	err := s.mysql.DB().QueryRowContext(ctx,
		"SELECT uid, password_hash FROM user WHERE username = ?", req.Username).
		Scan(&uid, &passwordHash)
	if err != nil {
		return &ss.GetUserResponse{Found: false}, nil
	}
	return &ss.GetUserResponse{Uid: uid, PasswordHash: passwordHash, Found: true}, nil
}
```

### Step 2: Create config

Create `config/db.yaml`:

```yaml
server:
  grpc_port: 50052

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
```

### Step 3: Create main

Create `cmd/db/main.go`:

```go
package main

import (
	"flag"
	"fmt"
	"os"
	"os/signal"
	"syscall"

	"github.com/yourorg/wg_ai/internal/common/logger"
	"github.com/yourorg/wg_ai/internal/db"
	"github.com/yourorg/wg_ai/internal/rpc"
	ss "github.com/yourorg/wg_ai/proto/ss"
)

func main() {
	configPath := flag.String("config", "config/db.yaml", "config file")
	flag.Parse()

	// TODO: Load config from file
	logger.New(os.Stderr, "info")

	// Initialize MySQL (use env vars for now)
	mysqlCfg := &db.MySQLConfig{
		Host:     getEnv("MYSQL_HOST", "127.0.0.1"),
		Port:     3306,
		Database: "game",
		Username: getEnv("MYSQL_USER", "root"),
		Password: getEnv("MYSQL_PASSWORD", ""),
		MaxOpen:  100,
		MaxIdle:  20,
	}
	mysql, err := db.NewMySQL(mysqlCfg)
	if err != nil {
		logger.Log.Fatalf("MySQL init failed: %v", err)
	}
	defer mysql.Close()

	// Initialize Redis
	redisCfg := &db.RedisConfig{
		Host:     getEnv("REDIS_HOST", "127.0.0.1"),
		Port:     6379,
		PoolSize: 100,
	}
	redis, err := db.NewRedis(redisCfg)
	if err != nil {
		logger.Log.Fatalf("Redis init failed: %v", err)
	}
	defer redis.Close()

	// Create gRPC server
	srv := rpc.NewServer(":50052")
	dbService := db.NewDBService(mysql, redis)
	ss.RegisterDBServiceServer(srv.GRPCServer(), dbService)

	// Start
	go func() {
		logger.Log.Info("DB server starting on :50052")
		if err := srv.Start(); err != nil {
			logger.Log.Fatalf("Server failed: %v", err)
		}
	}()

	// Wait for shutdown
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit
	logger.Log.Info("Shutting down...")
	srv.Stop()
}

func getEnv(key, fallback string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return fallback
}
```

### Step 4: Build and test

```bash
cd /root/ai_project/wg_ai
go build -o bin/db ./cmd/db
```

### Step 5: Commit

```bash
git add .
git commit -m "feat: add DB server entry point"
```

## 完成标志

- [ ] service.go 实现所有 RPC 方法
- [ ] main.go 可编译
- [ ] config/db.yaml 存在
- [ ] Commit 完成
