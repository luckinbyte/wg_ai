package main

import (
	"context"
	"flag"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/luckinbyte/wg_ai/internal/common/logger"
	"github.com/luckinbyte/wg_ai/internal/db"
	"github.com/luckinbyte/wg_ai/internal/rpc"
	ss "github.com/luckinbyte/wg_ai/proto/ss"
)

func main() {
	_ = flag.String("config", "config/db.yaml", "config file")
	flag.Parse()

	// TODO: Load config from file
	logger.New(os.Stderr, "info")

	// Initialize MySQL (use env vars for now)
	mysqlCfg := &db.MySQLConfig{
		Host:     getEnv("MYSQL_HOST", "127.0.0.1"),
		Port:     3306,
		Database: getEnv("MYSQL_DATABASE", "game"),
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
