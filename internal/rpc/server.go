package rpc

import (
	"net"

	"google.golang.org/grpc"
)

type Server struct {
	addr   string
	server *grpc.Server
}

func NewServer(addr string) *Server {
	return &Server{
		addr:   addr,
		server: grpc.NewServer(),
	}
}

func (s *Server) Start() error {
	lis, err := net.Listen("tcp", s.addr)
	if err != nil {
		return err
	}
	return s.server.Serve(lis)
}

func (s *Server) Stop() {
	s.server.GracefulStop()
}

func (s *Server) GRPCServer() *grpc.Server {
	return s.server
}
