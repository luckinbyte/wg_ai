# Task 21: Game Server - 游戏服务

## 背景与目标

实现 Game Server，整合 Gate、Agent、Session 等组件，提供完整的游戏服务。

**输出：**
- `internal/game/server.go` - Game 服务器
- `cmd/game/main.go` - 服务入口

## 依赖

- Phase 2: 所有核心组件
- Task 17-18: gRPC Client
- Task 19: DB Server
- Task 20: Login Server

## 步骤

### Step 1: Create game server

Create `internal/game/server.go`:

```go
package game

import (
	"github.com/yourorg/wg_ai/internal/agent"
	"github.com/yourorg/wg_ai/internal/common/config"
	"github.com/yourorg/wg_ai/internal/gate"
	"github.com/yourorg/wg_ai/internal/rpc"
	"github.com/yourorg/wg_ai/internal/session"
)

type GameServer struct {
	config     *config.GameConfig
	tcpServer  *gate.TCPServer
	agentMgr   *agent.Manager
	sessionMgr *session.Manager
	rpcClient  *rpc.Client
}

func NewGameServer(cfg *config.GameConfig) *GameServer {
	return &GameServer{
		config:     cfg,
		sessionMgr: session.NewManager(),
	}
}

func (s *GameServer) Start() error {
	// Create agent manager
	s.agentMgr = agent.NewManager(s.config.Agent.Count, s.config.Gate.MsgQueueSize)

	// Connect to DB
	s.rpcClient = rpc.NewClient(&rpc.ClientConfig{
		DBAddr: s.config.Cluster.DBAddr,
	})
	if err := s.rpcClient.ConnectDB(s.config.Cluster.DBAddr); err != nil {
		return err
	}

	// Start TCP server
	addr := s.config.Server.Addr()
	s.tcpServer = gate.NewTCPServer(addr, s.sessionMgr, s.agentMgr)

	go func() {
		if err := s.tcpServer.Start(); err != nil {
			panic(err)
		}
	}()

	return nil
}

func (s *GameServer) Stop() {
	if s.tcpServer != nil {
		s.tcpServer.Stop()
	}
	if s.agentMgr != nil {
		s.agentMgr.Stop()
	}
	if s.rpcClient != nil {
		s.rpcClient.Close()
	}
}
```

### Step 2: Create main

Create `cmd/game/main.go`:

```go
package main

import (
	"flag"
	"fmt"
	"os"
	"os/signal"
	"syscall"

	"github.com/yourorg/wg_ai/internal/common/config"
	"github.com/yourorg/wg_ai/internal/common/logger"
	"github.com/yourorg/wg_ai/internal/game"
)

func main() {
	configPath := flag.String("config", "config/game.yaml", "config file path")
	flag.Parse()

	cfg, err := config.LoadGameConfig(*configPath)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Load config failed: %v\n", err)
		os.Exit(1)
	}

	logger.New(os.Stderr, cfg.Log.Level)
	logger.Log.Infof("Starting game server %s (id=%d)", cfg.Server.Name, cfg.Server.ID)

	srv := game.NewGameServer(cfg)
	if err := srv.Start(); err != nil {
		logger.Log.Fatalf("Start server failed: %v", err)
	}
	defer srv.Stop()

	logger.Log.Infof("Game server listening on %s", cfg.Server.Addr())

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	logger.Log.Info("Shutting down...")
}
```

### Step 3: Build

```bash
cd /root/ai_project/wg_ai
go build -o bin/game ./cmd/game
```

### Step 4: Commit

```bash
git add .
git commit -m "feat: add game server entry point"
```

## 完成标志

- [ ] server.go 整合所有组件
- [ ] main.go 可编译
- [ ] Commit 完成
