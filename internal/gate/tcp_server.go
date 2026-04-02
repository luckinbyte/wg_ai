package gate

import (
	"net"
	"sync"
	"time"

	"github.com/yourorg/wg_ai/internal/agent"
	"github.com/yourorg/wg_ai/internal/session"
)

type TCPServer struct {
	addr       string
	listener   net.Listener
	stopCh     chan struct{}
	wg         sync.WaitGroup
	sessionMgr *session.Manager
	agentMgr   *agent.Manager
}

func NewTCPServer(addr string, sessionMgr *session.Manager, agentMgr *agent.Manager) *TCPServer {
	return &TCPServer{
		addr:       addr,
		stopCh:     make(chan struct{}),
		sessionMgr: sessionMgr,
		agentMgr:   agentMgr,
	}
}

func (s *TCPServer) Start() error {
	ln, err := net.Listen("tcp", s.addr)
	if err != nil {
		return err
	}
	s.listener = ln

	for {
		select {
		case <-s.stopCh:
			return nil
		default:
			conn, err := ln.Accept()
			if err != nil {
				continue
			}
			s.wg.Add(1)
			go s.handleConnection(conn)
		}
	}
}

func (s *TCPServer) Stop() {
	close(s.stopCh)
	if s.listener != nil {
		s.listener.Close()
	}
	s.wg.Wait()
}

func (s *TCPServer) Addr() string {
	if s.listener != nil {
		return s.listener.Addr().String()
	}
	return ""
}

func (s *TCPServer) handleConnection(conn net.Conn) {
	defer s.wg.Done()
	defer conn.Close()

	c := NewConnection(conn)
	conn.SetDeadline(time.Now().Add(30 * time.Second))

	// Read handshake
	msgType, _, err := c.ReadMessage()
	if err != nil {
		return
	}
	if msgType != MsgTypeHandshake {
		return
	}

	// TODO: Validate token via login service
	// For now, just send success response

	// Send handshake response
	resp := []byte{0, 0, 0, 0} // code = 0 (success)
	c.WriteMessage(MsgTypeResponse, resp)

	// Message loop
	for {
		conn.SetReadDeadline(time.Now().Add(30 * time.Second))
		msgType, data, err := c.ReadMessage()
		if err != nil {
			break
		}

		if msgType == MsgTypeRequest && len(data) >= 2 {
			ag := s.agentMgr.Assign()
			ag.Push(&agent.Message{
				MsgID:   uint16(data[0])<<8 | uint16(data[1]),
				Payload: data[2:],
			})
		}
	}
}
