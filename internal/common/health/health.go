package health

import (
	"sync"
)

type Checker struct {
	ready   bool
	healthy bool
	mutex   sync.RWMutex
}

func NewChecker() *Checker {
	return &Checker{}
}

func (c *Checker) IsReady() bool {
	c.mutex.RLock()
	defer c.mutex.RUnlock()
	return c.ready
}

func (c *Checker) SetReady(ready bool) {
	c.mutex.Lock()
	defer c.mutex.Unlock()
	c.ready = ready
}

func (c *Checker) IsHealthy() bool {
	c.mutex.RLock()
	defer c.mutex.RUnlock()
	return c.healthy
}

func (c *Checker) SetHealthy(healthy bool) {
	c.mutex.Lock()
	defer c.mutex.Unlock()
	c.healthy = healthy
}
