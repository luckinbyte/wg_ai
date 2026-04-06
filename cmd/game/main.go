package main

import (
	"context"
	"flag"
	"fmt"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/luckinbyte/wg_ai/internal/common/config"
	"github.com/luckinbyte/wg_ai/internal/common/logger"
	"github.com/luckinbyte/wg_ai/internal/game"
)

func main() {
	configPath := flag.String("config", "config/game.yaml", "config file path")
	flag.Parse()

	cfg, err := config.LoadGameConfig(*configPath)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Failed to load config: %v\n", err)
		os.Exit(1)
	}

	logger.New(os.Stderr, cfg.Log.Level)
	logger.Log.Infof("Starting game server %s (id=%d)", cfg.Server.Name, cfg.Server.ID)

	srv := game.NewServer(cfg)
	if err := srv.Start(); err != nil {
		logger.Log.Fatalf("Failed to start server: %v", err)
	}

	logger.Log.Infof("Game server listening on %s", cfg.Server.Addr())

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
