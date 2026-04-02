package session

import (
	"net"
	"sync"
	"time"
)

type Session struct {
	RID        int64
	UID        int64
	Conn       net.Conn
	CreatedAt  time.Time
	LastActive time.Time
	mutex      sync.RWMutex
}

func New(uid, rid int64, conn net.Conn) *Session {
	return &Session{
		UID:        uid,
		RID:        rid,
		Conn:       conn,
		CreatedAt:  time.Now(),
		LastActive: time.Now(),
	}
}

func (s *Session) Send(data []byte) error {
	s.mutex.Lock()
	defer s.mutex.Unlock()

	s.Conn.SetWriteDeadline(time.Now().Add(5 * time.Second))
	_, err := s.Conn.Write(data)
	return err
}

func (s *Session) Close() error {
	s.mutex.Lock()
	defer s.mutex.Unlock()
	if s.Conn != nil {
		return s.Conn.Close()
	}
	return nil
}

func (s *Session) UpdateActive() {
	s.mutex.Lock()
	defer s.mutex.Unlock()
	s.LastActive = time.Now()
}
