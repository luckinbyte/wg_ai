# Task 20: Login Server - 登录服务

## 背景与目标

实现 Login Server，提供 JWT Token 生成和验证。

**输出：**
- `internal/auth/jwt.go` - JWT 认证
- `cmd/login/main.go` - 服务入口
- `config/login.yaml` - 配置文件

## 依赖

- Task 17-18: gRPC Client（验证 Token）
- Task 19: DB Server（用户数据）

## 步骤

### Step 1: Write JWT test

Create `internal/auth/jwt_test.go`:

```go
package auth

import (
	"testing"
)

func TestJWTGenerateAndValidate(t *testing.T) {
	secret := "test-secret-key"
	
	token, err := GenerateToken(123, secret)
	if err != nil {
		t.Fatalf("generate failed: %v", err)
	}
	
	uid, err := ValidateToken(token, secret)
	if err != nil {
		t.Fatalf("validate failed: %v", err)
	}
	if uid != 123 {
		t.Errorf("expected uid 123, got %d", uid)
	}
}

func TestJWTInvalidToken(t *testing.T) {
	_, err := ValidateToken("invalid-token", "secret")
	if err == nil {
		t.Error("should fail for invalid token")
	}
}

func TestJWTWrongSecret(t *testing.T) {
	token, _ := GenerateToken(123, "secret1")
	_, err := ValidateToken(token, "secret2")
	if err == nil {
		t.Error("should fail for wrong secret")
	}
}
```

### Step 2: Write JWT implementation

Create `internal/auth/jwt.go`:

```go
package auth

import (
	"time"

	"github.com/golang-jwt/jwt/v5"
)

type Claims struct {
	UID int64 `json:"uid"`
	jwt.RegisteredClaims
}

func GenerateToken(uid int64, secret string) (string, error) {
	claims := Claims{
		UID: uid,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(24 * time.Hour)),
			IssuedAt:  jwt.NewNumericDate(time.Now()),
		},
	}
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString([]byte(secret))
}

func ValidateToken(tokenString, secret string) (int64, error) {
	token, err := jwt.ParseWithClaims(tokenString, &Claims{},
		func(token *jwt.Token) (interface{}, error) {
			return []byte(secret), nil
		})
	if err != nil {
		return 0, err
	}

	if claims, ok := token.Claims.(*Claims); ok && token.Valid {
		return claims.UID, nil
	}

	return 0, jwt.ErrSignatureInvalid
}
```

### Step 3: Install dependency and test

```bash
cd /root/ai_project/wg_ai
go get github.com/golang-jwt/jwt/v5
go test ./internal/auth/...
```

Expected: PASS

### Step 4: Create config

Create `config/login.yaml`:

```yaml
server:
  grpc_port: 50051

token:
  secret: "your-secret-key-change-in-production"
  expire: "24h"

database:
  db_addr: "127.0.0.1:50052"

log:
  level: "info"
```

### Step 5: Create main

Create `cmd/login/main.go`:

```go
package main

import (
	"context"
	"flag"
	"os"
	"os/signal"
	"syscall"

	"github.com/yourorg/wg_ai/internal/auth"
	"github.com/yourorg/wg_ai/internal/common/logger"
	"github.com/yourorg/wg_ai/internal/rpc"
	ss "github.com/yourorg/wg_ai/proto/ss"
)

type LoginServer struct {
	ss.UnimplementedLoginServiceServer
	tokenSecret string
}

func (s *LoginServer) ValidateToken(ctx context.Context, req *ss.ValidateTokenRequest) (*ss.ValidateTokenResponse, error) {
	uid, err := auth.ValidateToken(req.Token, s.tokenSecret)
	if err != nil {
		return &ss.ValidateTokenResponse{Valid: false}, nil
	}
	return &ss.ValidateTokenResponse{Uid: uid, Valid: true}, nil
}

func (s *LoginServer) NotifyLogin(ctx context.Context, req *ss.LoginNotifyRequest) (*ss.LoginNotifyResponse, error) {
	return &ss.LoginNotifyResponse{Success: true}, nil
}

func main() {
	configPath := flag.String("config", "config/login.yaml", "config file")
	flag.Parse()

	secret := getEnv("TOKEN_SECRET", "default-secret-key")

	logger.New(os.Stderr, "info")

	srv := rpc.NewServer(":50051")
	loginSrv := &LoginServer{tokenSecret: secret}
	ss.RegisterLoginServiceServer(srv.GRPCServer(), loginSrv)

	go func() {
		logger.Log.Info("Login server starting on :50051")
		if err := srv.Start(); err != nil {
			logger.Log.Fatalf("Server failed: %v", err)
		}
	}()

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit
	srv.Stop()
}

func getEnv(key, fallback string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return fallback
}
```

### Step 6: Build and commit

```bash
cd /root/ai_project/wg_ai
go build -o bin/login ./cmd/login
git add .
git commit -m "feat: add login server with JWT authentication"
```

## 完成标志

- [ ] JWT 测试通过
- [ ] main.go 可编译
- [ ] config/login.yaml 存在
- [ ] Commit 完成
