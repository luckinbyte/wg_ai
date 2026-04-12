package gate

import (
	"log"
	"net/http"
	"sync"
	"time"

	"github.com/gorilla/websocket"
	"github.com/luckinbyte/wg_ai/internal/agent"
	"github.com/luckinbyte/wg_ai/internal/session"
)

var upgrader = websocket.Upgrader{
	ReadBufferSize:  4096,
	WriteBufferSize: 4096,
	CheckOrigin: func(r *http.Request) bool {
		return true
	},
}

type WSServer struct {
	addr       string
	server     *http.Server
	stopCh     chan struct{}
	wg         sync.WaitGroup
	sessionMgr *session.Manager
	agentMgr   *agent.Manager
}

func NewWSServer(addr string, sessionMgr *session.Manager, agentMgr *agent.Manager) *WSServer {
	return &WSServer{
		addr:       addr,
		stopCh:     make(chan struct{}),
		sessionMgr: sessionMgr,
		agentMgr:   agentMgr,
	}
}

func (s *WSServer) Start() error {
	mux := http.NewServeMux()
	mux.HandleFunc("/ws", s.handleWebSocket)

	s.server = &http.Server{
		Addr:    s.addr,
		Handler: mux,
	}

	go func() {
		log.Printf("WebSocket server starting on %s", s.addr)
		if err := s.server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Printf("WebSocket server error: %v", err)
		}
	}()

	return nil
}

func (s *WSServer) Stop() {
	close(s.stopCh)
	if s.server != nil {
		s.server.Close()
	}
	s.wg.Wait()
}

func (s *WSServer) Addr() string {
	if s.server != nil {
		return s.server.Addr
	}
	return ""
}

func (s *WSServer) handleWebSocket(w http.ResponseWriter, r *http.Request) {
	conn, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		log.Printf("WebSocket upgrade failed: %v", err)
		return
	}

	log.Printf("WebSocket connection established from %s", conn.RemoteAddr())
	s.wg.Add(1)
	go s.handleConnection(conn)
}

func (s *WSServer) handleConnection(conn *websocket.Conn) {
	defer s.wg.Done()
	defer conn.Close()

	wsConn := NewWSConnection(conn)

	conn.SetReadDeadline(time.Now().Add(30 * time.Second))

	msgType, data, err := wsConn.ReadMessage()
	if err != nil {
		log.Printf("WebSocket read handshake failed: %v", err)
		return
	}
	if msgType != MsgTypeHandshake {
		log.Printf("WebSocket invalid handshake message type: %d", msgType)
		return
	}

	_ = data

	sess := session.New(1, 0, nil)
	sess.SetWSConn(wsConn)

	sess = s.sessionMgr.Create(1, nil)
	sess.SetWSConn(wsConn)

	ag := s.agentMgr.Assign()
	ag.BindSession(sess)

	resp := []byte{0, 0, 0, 0}
	wsConn.WriteMessage(MsgTypeResponse, resp)

	log.Printf("WebSocket handshake completed for UID=1")

	for {
		conn.SetReadDeadline(time.Now().Add(30 * time.Second))
		msgType, data, err := wsConn.ReadMessage()
		if err != nil {
			log.Printf("WebSocket read error: %v", err)
			break
		}

		if msgType == MsgTypeRequest && len(data) >= 2 {
			ag.Push(&agent.Message{
				MsgID:   uint16(data[0])<<8 | uint16(data[1]),
				Payload: data[2:],
				Sess:    sess,
			})
		}
	}

	ag.UnbindSession(sess.RID)
	s.sessionMgr.Remove(sess.RID)
}

type WSConnection struct {
	conn  *websocket.Conn
	mutex sync.Mutex
}

func NewWSConnection(conn *websocket.Conn) *WSConnection {
	return &WSConnection{conn: conn}
}

func (c *WSConnection) ReadMessage() (msgType byte, payload []byte, err error) {
	_, data, err := c.conn.ReadMessage()
	if err != nil {
		return 0, nil, err
	}

	if len(data) < 1 {
		return 0, nil, ErrInvalidMessage
	}

	msgType = data[0]
	payload = data[1:]
	return msgType, payload, nil
}

func (c *WSConnection) WriteMessage(msgType byte, payload []byte) error {
	c.mutex.Lock()
	defer c.mutex.Unlock()

	data := make([]byte, 1+len(payload))
	data[0] = msgType
	copy(data[1:], payload)

	return c.conn.WriteMessage(websocket.BinaryMessage, data)
}

func (c *WSConnection) Send(data []byte) error {
	c.mutex.Lock()
	defer c.mutex.Unlock()

	return c.conn.WriteMessage(websocket.BinaryMessage, data)
}

func (c *WSConnection) Close() error {
	return c.conn.Close()
}

func (c *WSConnection) RemoteAddr() string {
	return c.conn.RemoteAddr().String()
}
