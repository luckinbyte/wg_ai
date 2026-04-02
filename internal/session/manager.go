package session

import (
	"net"
	"sync"
	"sync/atomic"
	"time"
)

type Manager struct {
	sessions map[int64]*Session
	mutex    sync.RWMutex
	ridSeq   int64
}

func NewManager() *Manager {
	return &Manager{
		sessions: make(map[int64]*Session),
		ridSeq:   time.Now().Unix(),
	}
}

func (m *Manager) Create(uid int64, conn net.Conn) *Session {
	m.mutex.Lock()
	defer m.mutex.Unlock()

	rid := atomic.AddInt64(&m.ridSeq, 1)
	sess := New(uid, rid, conn)
	m.sessions[rid] = sess
	return sess
}

func (m *Manager) Get(rid int64) *Session {
	m.mutex.RLock()
	defer m.mutex.RUnlock()
	return m.sessions[rid]
}

func (m *Manager) Remove(rid int64) {
	m.mutex.Lock()
	defer m.mutex.Unlock()
	delete(m.sessions, rid)
}

func (m *Manager) Count() int {
	m.mutex.RLock()
	defer m.mutex.RUnlock()
	return len(m.sessions)
}

func (m *Manager) GetAll() []*Session {
	m.mutex.RLock()
	defer m.mutex.RUnlock()
	result := make([]*Session, 0, len(m.sessions))
	for _, s := range m.sessions {
		result = append(result, s)
	}
	return result
}
