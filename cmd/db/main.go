package main

import (
	"context"
	"flag"
	"fmt"
	"os"
	"os/signal"
	"strconv"
	"syscall"
	"time"

	configpkg "github.com/luckinbyte/wg_ai/internal/common/config"
	"github.com/luckinbyte/wg_ai/internal/common/logger"
	"github.com/luckinbyte/wg_ai/internal/data"
	"github.com/luckinbyte/wg_ai/internal/db"
	"github.com/luckinbyte/wg_ai/internal/rpc"
	ss "github.com/luckinbyte/wg_ai/proto/ss"
)

func main() {
	configPath := flag.String("config", "config/db.yaml", "config file")
	flag.Parse()

	cfg, err := configpkg.LoadGameConfig(*configPath)
	if err != nil {
		fmt.Fprintf(os.Stderr, "load config failed: %v\n", err)
		os.Exit(1)
	}

	logger.New(os.Stderr, getEnv("LOG_LEVEL", cfg.Log.Level))

	mysqlCfg := &db.MySQLConfig{
		Host:     getEnv("MYSQL_HOST", cfg.Database.MySQL.Host),
		Port:     getEnvInt("MYSQL_PORT", cfg.Database.MySQL.Port),
		Database: getEnv("MYSQL_DATABASE", cfg.Database.MySQL.Database),
		Username: getEnv("MYSQL_USER", cfg.Database.MySQL.Username),
		Password: getEnv("MYSQL_PASSWORD", cfg.Database.MySQL.Password),
		MaxOpen:  cfg.Database.MySQL.MaxOpen,
		MaxIdle:  cfg.Database.MySQL.MaxIdle,
	}
	if err := db.EnsureDatabase(mysqlCfg); err != nil {
		logger.Log.Fatalf("MySQL database bootstrap failed: %v", err)
	}

	mysql, err := db.NewMySQL(mysqlCfg)
	if err != nil {
		logger.Log.Fatalf("MySQL init failed: %v", err)
	}
	defer mysql.Close()

	if err := db.EnsureTables(mysql); err != nil {
		logger.Log.Fatalf("MySQL table init failed: %v", err)
	}
	if err := data.NewPersist(mysql.DB()).CreateTable(); err != nil {
		logger.Log.Fatalf("player_data table init failed: %v", err)
	}

	redisCfg := &db.RedisConfig{
		Host:     getEnv("REDIS_HOST", cfg.Database.Redis.Host),
		Port:     getEnvInt("REDIS_PORT", cfg.Database.Redis.Port),
		DB:       cfg.Database.Redis.DB,
		PoolSize: cfg.Database.Redis.PoolSize,
	}
	redis, err := db.NewRedis(redisCfg)
	if err != nil {
		logger.Log.Fatalf("Redis init failed: %v", err)
	}
	defer redis.Close()

	grpcPort := getEnvInt("GRPC_PORT", cfg.Server.GRPCPort)
	addr := fmt.Sprintf(":%d", grpcPort)
	srv := rpc.NewServer(addr)
	dbService := db.NewDBService(mysql, redis)
	ss.RegisterDBServiceServer(srv.GRPCServer(), dbService)

	go func() {
		logger.Log.Infof("DB server starting on %s", addr)
		if err := srv.Start(); err != nil {
			logger.Log.Fatalf("Server failed: %v", err)
		}
	}()

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	logger.Log.Info("Shutting down...")

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	done := make(chan struct{})
	go func() {
		srv.Stop()
		close(done)
	}()

	select {
	case <-done:
		logger.Log.Info("Server stopped gracefully")
	case <-ctx.Done():
		logger.Log.Warn("Shutdown timeout, forcing exit")
	}
}

func getEnv(key, fallback string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return fallback
}

func getEnvInt(key string, fallback int) int {
	if v := os.Getenv(key); v != "" {
		if n, err := strconv.Atoi(v); err == nil {
			return n
		}
	}
	return fallback
}
