package game

import (
	"github.com/yourorg/wg_ai/internal/agent"
	"github.com/yourorg/wg_ai/internal/common/config"
	"github.com/yourorg/wg_ai/internal/gate"
	"github.com/yourorg/wg_ai/internal/rpc"
	"github.com/yourorg/wg_ai/internal/session"
)

type Server struct {
	config     *config.GameConfig
	tcpServer  *gate.TCPServer
	agentMgr   *agent.Manager
	sessionMgr *session.Manager
	rpcClient  *rpc.Client
}

func NewServer(cfg *config.GameConfig) *Server {
	return &Server{
		config:     cfg,
		sessionMgr: session.NewManager(),
	}
}

func (s *Server) Start() error {
	// Create agent manager
	s.agentMgr = agent.NewManager(s.config.Agent.Count, s.config.Gate.MsgQueueSize)

	// Connect to services
	s.rpcClient = rpc.NewClient(&rpc.ClientConfig{
		DBAddr:    s.config.Cluster.DBAddr,
		LoginAddr: s.config.Cluster.LoginAddr,
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

func (s *Server) Stop() {
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
