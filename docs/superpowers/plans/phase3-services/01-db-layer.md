# Task 15-16: DB Layer - 数据访问层

## 背景与目标

实现 MySQL 和 Redis 数据访问层，为 DB Server 提供数据存储能力。

**输出：**
- `internal/db/mysql.go` - MySQL 操作
- `internal/db/redis.go` - Redis 操作

## 依赖

- Task 2: 配置系统（MySQLConfig, RedisConfig）

## 步骤

### Step 1: Write MySQL test

Create `internal/db/mysql_test.go`:

```go
package db

import (
	"testing"
)

func TestMySQLConfigDSN(t *testing.T) {
	cfg := &MySQLConfig{
		Host:     "localhost",
		Port:     3306,
		Database: "game",
		Username: "root",
		Password: "secret",
	}

	dsn := cfg.DSN()
	expected := "root:secret@tcp(localhost:3306)/game?charset=utf8mb4&parseTime=True"
	if dsn != expected {
		t.Errorf("DSN mismatch:\ngot:      %s\nexpected: %s", dsn, expected)
	}
}
```

### Step 2: Write MySQL implementation

Create `internal/db/mysql.go`:

```go
package db

import (
	"database/sql"
	"fmt"

	_ "github.com/go-sql-driver/mysql"
)

type MySQL struct {
	db *sql.DB
}

func NewMySQL(cfg *MySQLConfig) (*MySQL, error) {
	db, err := sql.Open("mysql", cfg.DSN())
	if err != nil {
		return nil, err
	}

	if cfg.MaxOpen > 0 {
		db.SetMaxOpenConns(cfg.MaxOpen)
	}
	if cfg.MaxIdle > 0 {
		db.SetMaxIdleConns(cfg.MaxIdle)
	}

	if err := db.Ping(); err != nil {
		return nil, err
	}

	return &MySQL{db: db}, nil
}

func (m *MySQL) Close() error {
	return m.db.Close()
}

func (m *MySQL) DB() *sql.DB {
	return m.db
}
```

### Step 3: Write Redis test

Create `internal/db/redis_test.go`:

```go
package db

import "testing"

func TestRedisConfigAddr(t *testing.T) {
	cfg := &RedisConfig{
		Host: "localhost",
		Port: 6379,
	}

	addr := cfg.Addr()
	if addr != "localhost:6379" {
		t.Errorf("expected localhost:6379, got %s", addr)
	}
}
```

### Step 4: Write Redis implementation

Create `internal/db/redis.go`:

```go
package db

import (
	"context"
	"fmt"
	"time"

	"github.com/redis/go-redis/v9"
)

type Redis struct {
	client *redis.Client
}

func NewRedis(cfg *RedisConfig) (*Redis, error) {
	client := redis.NewClient(&redis.Options{
		Addr:     cfg.Addr(),
		DB:       cfg.DB,
		PoolSize: cfg.PoolSize,
	})

	ctx := context.Background()
	if err := client.Ping(ctx).Err(); err != nil {
		return nil, err
	}

	return &Redis{client: client}, nil
}

func (r *Redis) Close() error {
	return r.client.Close()
}

func (r *Redis) Client() *redis.Client {
	return r.client
}

func (r *Redis) Get(ctx context.Context, key string) ([]byte, error) {
	return r.client.Get(ctx, key).Bytes()
}

func (r *Redis) Set(ctx context.Context, key string, value []byte, ttl time.Duration) error {
	return r.client.Set(ctx, key, value, ttl).Err()
}

func (r *Redis) Del(ctx context.Context, keys ...string) error {
	return r.client.Del(ctx, keys...).Err()
}
```

### Step 5: Install dependencies and test

```bash
cd /root/ai_project/wg_ai
go get github.com/go-sql-driver/mysql
go get github.com/redis/go-redis/v9
go test ./internal/db/...
```

Expected: PASS

### Step 6: Commit

```bash
git add .
git commit -m "feat: add MySQL and Redis data layer"
```

## 完成标志

- [ ] MySQL 测试通过
- [ ] Redis 测试通过
- [ ] 依赖安装完成
- [ ] Commit 完成
