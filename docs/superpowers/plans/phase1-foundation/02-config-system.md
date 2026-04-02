# Task 2: Configuration System - 配置系统

## 背景与目标

建立统一的配置加载系统，使用 Viper 库支持 YAML 配置文件。

**为什么需要这个任务：**
- 所有服务都需要读取配置（端口、数据库连接等）
- 统一的配置接口方便维护
- Viper 是 Go 生态中最流行的配置库

**输出：**
- `internal/common/config/config.go` - 配置加载器
- `config/game.yaml` - 游戏服务器配置示例

## 依赖

- Task 1: 项目必须已初始化（go.mod 存在）

## 步骤

### Step 1: Write the failing test

Create `internal/common/config/config_test.go`:

```go
package config

import (
	"os"
	"testing"
)

func TestLoadGameConfig(t *testing.T) {
	content := `
server:
  id: 1
  name: "test-game"
  host: "0.0.0.0"
  port: 44445
  max_conn: 1000
gate:
  read_timeout: 30s
  write_timeout: 30s
`
	tmpFile, err := os.CreateTemp("", "config-*.yaml")
	if err != nil {
		t.Fatal(err)
	}
	defer os.Remove(tmpFile.Name())
	tmpFile.Write([]byte(content))
	tmpFile.Close()

	cfg, err := LoadGameConfig(tmpFile.Name())
	if err != nil {
		t.Fatalf("LoadGameConfig failed: %v", err)
	}

	if cfg.Server.ID != 1 {
		t.Errorf("expected Server.ID=1, got %d", cfg.Server.ID)
	}
	if cfg.Server.Port != 44445 {
		t.Errorf("expected Server.Port=44445, got %d", cfg.Server.Port)
	}
	if cfg.Gate.ReadTimeout != 30*time.Second {
		t.Errorf("expected ReadTimeout=30s, got %v", cfg.Gate.ReadTimeout)
	}
}
```

### Step 2: Run test to verify it fails

```bash
cd /root/ai_project/wg_ai
go test ./internal/common/config/...
```

Expected: FAIL - undefined: LoadGameConfig, GameConfig

### Step 3: Write minimal implementation

Create `internal/common/config/config.go`:

```go
package config

import (
	"fmt"
	"time"

	"github.com/spf13/viper"
)

type GameConfig struct {
	Server  ServerConfig  `mapstructure:"server"`
	Gate    GateConfig    `mapstructure:"gate"`
	Agent   AgentConfig   `mapstructure:"agent"`
	Cluster ClusterConfig `mapstructure:"cluster"`
	Database DatabaseConfig `mapstructure:"database"`
	Log     LogConfig     `mapstructure:"log"`
}

type ServerConfig struct {
	ID      int    `mapstructure:"id"`
	Name    string `mapstructure:"name"`
	Host    string `mapstructure:"host"`
	Port    int    `mapstructure:"port"`
	MaxConn int    `mapstructure:"max_conn"`
}

func (c *ServerConfig) Addr() string {
	return fmt.Sprintf("%s:%d", c.Host, c.Port)
}

type GateConfig struct {
	ReadTimeout   time.Duration `mapstructure:"read_timeout"`
	WriteTimeout  time.Duration `mapstructure:"write_timeout"`
	MsgQueueSize  int           `mapstructure:"msg_queue_size"`
}

type AgentConfig struct {
	Count           int `mapstructure:"count"`
	PlayersPerAgent int `mapstructure:"players_per_agent"`
}

type ClusterConfig struct {
	LoginAddr string `mapstructure:"login_addr"`
	DBAddr    string `mapstructure:"db_addr"`
}

type DatabaseConfig struct {
	MySQL MySQLConfig `mapstructure:"mysql"`
	Redis RedisConfig `mapstructure:"redis"`
}

type MySQLConfig struct {
	Host     string `mapstructure:"host"`
	Port     int    `mapstructure:"port"`
	Database string `mapstructure:"database"`
	Username string `mapstructure:"username"`
	Password string `mapstructure:"password"`
	MaxOpen  int    `mapstructure:"max_open"`
	MaxIdle  int    `mapstructure:"max_idle"`
}

func (c *MySQLConfig) DSN() string {
	return fmt.Sprintf("%s:%s@tcp(%s:%d)/%s?charset=utf8mb4&parseTime=True",
		c.Username, c.Password, c.Host, c.Port, c.Database)
}

type RedisConfig struct {
	Host     string `mapstructure:"host"`
	Port     int    `mapstructure:"port"`
	DB       int    `mapstructure:"db"`
	PoolSize int    `mapstructure:"pool_size"`
}

func (c *RedisConfig) Addr() string {
	return fmt.Sprintf("%s:%d", c.Host, c.Port)
}

type LogConfig struct {
	Level  string `mapstructure:"level"`
	Output string `mapstructure:"output"`
}

func LoadGameConfig(path string) (*GameConfig, error) {
	v := viper.New()
	v.SetConfigFile(path)

	if err := v.ReadInConfig(); err != nil {
		return nil, err
	}

	var cfg GameConfig
	if err := v.Unmarshal(&cfg); err != nil {
		return nil, err
	}

	return &cfg, nil
}
```

### Step 4: Install dependency and run test

```bash
cd /root/ai_project/wg_ai
go get github.com/spf13/viper
go test ./internal/common/config/...
```

Expected: PASS

### Step 5: Create game.yaml config file

Create `config/game.yaml`:

```yaml
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

### Step 6: Commit

```bash
git add .
git commit -m "feat: add configuration system with viper"
```

## 验证

```bash
cd /root/ai_project/wg_ai
go test ./internal/common/config/... -v
```

Expected: PASS

## 完成标志

- [ ] 测试通过
- [ ] config.go 包含所有配置结构体
- [ ] game.yaml 配置文件存在
- [ ] Commit 完成
