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
	_ = flag.String("config", "config/login.yaml", "config file")
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
