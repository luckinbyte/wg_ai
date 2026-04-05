package config

import (
	"fmt"
	"time"

	"github.com/spf13/viper"
)

type GameConfig struct {
	Server   ServerConfig   `mapstructure:"server"`
	Gate     GateConfig     `mapstructure:"gate"`
	Agent    AgentConfig    `mapstructure:"agent"`
	Cluster  ClusterConfig  `mapstructure:"cluster"`
	Database DatabaseConfig `mapstructure:"database"`
	Log      LogConfig      `mapstructure:"log"`
	Plugin   PluginConfig   `mapstructure:"plugin"`
	Admin    AdminConfig    `mapstructure:"admin"`
}

// PluginConfig 插件配置
type PluginConfig struct {
	Dir       string `mapstructure:"dir"`        // 插件目录: ./plugins
	RouteFile string `mapstructure:"route_file"` // 路由配置: ./config/routes.json
	Watch     bool   `mapstructure:"watch"`      // 是否启用文件监听
}

// AdminConfig 管理接口配置
type AdminConfig struct {
	Addr string `mapstructure:"addr"` // 管理接口地址: :8081
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
	ReadTimeout  time.Duration `mapstructure:"read_timeout"`
	WriteTimeout time.Duration `mapstructure:"write_timeout"`
	MsgQueueSize int           `mapstructure:"msg_queue_size"`
	WSPort       int           `mapstructure:"ws_port"` // WebSocket端口
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
