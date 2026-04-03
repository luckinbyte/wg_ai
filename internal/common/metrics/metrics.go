package metrics

import (
	"sync/atomic"
)

type Metrics struct {
	connections int64
	messages    int64
	errors      int64
}

func NewMetrics() *Metrics {
	return &Metrics{}
}

func (m *Metrics) IncConnections() {
	atomic.AddInt64(&m.connections, 1)
}

func (m *Metrics) DecConnections() {
	atomic.AddInt64(&m.connections, -1)
}

func (m *Metrics) GetConnections() int64 {
	return atomic.LoadInt64(&m.connections)
}

func (m *Metrics) IncMessages() {
	atomic.AddInt64(&m.messages, 1)
}

func (m *Metrics) GetMessages() int64 {
	return atomic.LoadInt64(&m.messages)
}

func (m *Metrics) IncErrors() {
	atomic.AddInt64(&m.errors, 1)
}

func (m *Metrics) GetErrors() int64 {
	return atomic.LoadInt64(&m.errors)
}
