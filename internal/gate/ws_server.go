package gate

import (
	"net/http"
	"sync"
	"time"

	"github.com/gorilla/websocket"
	"github.com/yourorg/wg_ai/internal/agent"
	"github.com/yourorg/wg_ai/internal/session"
)

var upgrader = websocket.Upgrader{
	ReadBufferSize:  4096,
	WriteBufferSize: 4096,
	CheckOrigin: func(r *http.Request) bool {
		return true // 允许所有来源 (生产环境应限制)
	},
}

// WSServer WebSocket服务器
type WSServer struct {
	addr       string
	server     *http.Server
	stopCh     chan struct{}
	wg         sync.WaitGroup
	sessionMgr *session.Manager
	agentMgr   *agent.Manager
}

// NewWSServer 创建WebSocket服务器
func NewWSServer(addr string, sessionMgr *session.Manager, agentMgr *agent.Manager) *WSServer {
	return &WSServer{
		addr:       addr,
		stopCh:     make(chan struct{}),
		sessionMgr: sessionMgr,
		agentMgr:   agentMgr,
	}
}

// Start 启动服务器
func (s *WSServer) Start() error {
	mux := http.NewServeMux()
	mux.HandleFunc("/ws", s.handleWebSocket)

	s.server = &http.Server{
		Addr:    s.addr,
		Handler: mux,
	}

	go func() {
		if err := s.server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			// 服务器错误
		}
	}()

	return nil
}

// Stop 停止服务器
func (s *WSServer) Stop() {
	close(s.stopCh)
	if s.server != nil {
		s.server.Close()
	}
	s.wg.Wait()
}

// Addr 返回监听地址
func (s *WSServer) Addr() string {
	if s.server != nil {
		return s.server.Addr
	}
	return ""
}

// handleWebSocket 处理WebSocket连接
func (s *WSServer) handleWebSocket(w http.ResponseWriter, r *http.Request) {
	conn, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		return
	}

	s.wg.Add(1)
	go s.handleConnection(conn)
}

// handleConnection 处理连接
func (s *WSServer) handleConnection(conn *websocket.Conn) {
	defer s.wg.Done()
	defer conn.Close()

	wsConn := NewWSConnection(conn)

	// 设置读超时
	conn.SetReadDeadline(time.Now().Add(30 * time.Second))

	// 读取握手消息
	msgType, data, err := wsConn.ReadMessage()
	if err != nil {
		return
	}
	if msgType != MsgTypeHandshake {
		return
	}

	// TODO: 解析并验证token
	// 暂时使用 UID=1 测试
	_ = data

	// 创建会话
	sess := session.New(1, 0, nil) // UID=1
	sess.SetWSConn(wsConn)         // 设置WS连接

	// 注册到会话管理器
	sess = s.sessionMgr.Create(1, nil)
	sess.SetWSConn(wsConn)

	// 绑定到Agent
	ag := s.agentMgr.Assign()
	ag.BindSession(sess)

	// 发送握手响应
	resp := []byte{0, 0, 0, 0} // code = 0 (成功)
	wsConn.WriteMessage(MsgTypeResponse, resp)

	// 消息循环
	for {
		conn.SetReadDeadline(time.Now().Add(30 * time.Second))
		msgType, data, err := wsConn.ReadMessage()
		if err != nil {
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

	// 清理
	ag.UnbindSession(sess.RID)
	s.sessionMgr.Remove(sess.RID)
}

// WSConnection WebSocket连接包装
type WSConnection struct {
	conn  *websocket.Conn
	mutex sync.Mutex
}

// NewWSConnection 创建WS连接
func NewWSConnection(conn *websocket.Conn) *WSConnection {
	return &WSConnection{conn: conn}
}

// ReadMessage 读取消息
func (c *WSConnection) ReadMessage() (msgType byte, payload []byte, err error) {
	// WebSocket读取二进制消息
	_, data, err := c.conn.ReadMessage()
	if err != nil {
		return 0, nil, err
	}

	if len(data) < 1 {
		return 0, nil, ErrInvalidMessage
	}

	// 解析: [MsgType(1B)][Payload]
	msgType = data[0]
	payload = data[1:]
	return msgType, payload, nil
}

// WriteMessage 写入消息
func (c *WSConnection) WriteMessage(msgType byte, payload []byte) error {
	c.mutex.Lock()
	defer c.mutex.Unlock()

	// 封装: [MsgType(1B)][Payload]
	data := make([]byte, 1+len(payload))
	data[0] = msgType
	copy(data[1:], payload)

	return c.conn.WriteMessage(websocket.BinaryMessage, data)
}

// Send 发送数据 (实现session.Conn接口)
func (c *WSConnection) Send(data []byte) error {
	c.mutex.Lock()
	defer c.mutex.Unlock()

	return c.conn.WriteMessage(websocket.BinaryMessage, data)
}

// Close 关闭连接
func (c *WSConnection) Close() error {
	return c.conn.Close()
}

// RemoteAddr 远程地址
func (c *WSConnection) RemoteAddr() string {
	return c.conn.RemoteAddr().String()
}
