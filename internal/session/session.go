package session

import (
	"net"
	"sync"
	"time"
)

// Conn 连接接口
type Conn interface {
	Send(data []byte) error
	Close() error
}

// Session 会话
type Session struct {
	RID        int64
	UID        int64
	Conn       net.Conn       // TCP连接 (兼容旧代码)
	wsConn     Conn           // WebSocket连接
	CreatedAt  time.Time
	LastActive time.Time
	mutex      sync.RWMutex
}

// New 创建会话
func New(uid, rid int64, conn net.Conn) *Session {
	return &Session{
		UID:        uid,
		RID:        rid,
		Conn:       conn,
		CreatedAt:  time.Now(),
		LastActive: time.Now(),
	}
}

// SetWSConn 设置WebSocket连接
func (s *Session) SetWSConn(conn Conn) {
	s.mutex.Lock()
	defer s.mutex.Unlock()
	s.wsConn = conn
}

// Send 发送数据
func (s *Session) Send(data []byte) error {
	s.mutex.Lock()
	defer s.mutex.Unlock()

	// 优先使用WS连接
	if s.wsConn != nil {
		return s.wsConn.Send(data)
	}

	// 使用TCP连接
	if s.Conn != nil {
		s.Conn.SetWriteDeadline(time.Now().Add(5 * time.Second))
		_, err := s.Conn.Write(data)
		return err
	}

	return nil
}

// Close 关闭会话
func (s *Session) Close() error {
	s.mutex.Lock()
	defer s.mutex.Unlock()

	var err error
	if s.wsConn != nil {
		err = s.wsConn.Close()
	}
	if s.Conn != nil {
		if tcpErr := s.Conn.Close(); tcpErr != nil {
			err = tcpErr
		}
	}
	return err
}

// UpdateActive 更新活跃时间
func (s *Session) UpdateActive() {
	s.mutex.Lock()
	defer s.mutex.Unlock()
	s.LastActive = time.Now()
}

// IsWS 是否是WebSocket连接
func (s *Session) IsWS() bool {
	s.mutex.RLock()
	defer s.mutex.RUnlock()
	return s.wsConn != nil
}
