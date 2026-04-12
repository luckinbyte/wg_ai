package agent

import (
	"sync/atomic"
)

type Manager struct {
	agents     []*Agent
	roundRobin uint32
}

func NewManager(agentCount, queueSize int) *Manager {
	m := &Manager{
		agents: make([]*Agent, agentCount),
	}
	for i := 0; i < agentCount; i++ {
		agent := New(i, queueSize)
		m.agents[i] = agent
		go agent.Run()
	}
	return m
}

func (m *Manager) Assign() *Agent {
	idx := atomic.AddUint32(&m.roundRobin, 1) - 1
	return m.agents[idx%uint32(len(m.agents))]
}

func (m *Manager) Get(id int) *Agent {
	if id >= 0 && id < len(m.agents) {
		return m.agents[id]
	}
	return nil
}

func (m *Manager) Stop() {
	for _, a := range m.agents {
		a.Stop()
	}
}

func (m *Manager) SetFallback(fn HandlerFunc) {
	for _, a := range m.agents {
		a.SetFallback(fn)
	}
}
