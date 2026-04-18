package data

import (
	"sync"
)

// PlayerData 玩家数据容器
type PlayerData struct {
	RID     int64          `json:"rid"`     // 角色ID
	Base    map[string]any `json:"base"`    // 基础字段: name, level, exp, gold...
	Arrays  map[string]any `json:"arrays"`  // 数组字段: items, heroes, tasks...
	Dirty   bool           `json:"-"`       // 脏标记
	version uint64         `json:"-"`
	mutex   sync.RWMutex   `json:"-"`       // 读写锁
}

// NewPlayerData 创建空的玩家数据
func NewPlayerData(rid int64) *PlayerData {
	return &PlayerData{
		RID: rid,
		Base: map[string]any{
			"food":  int64(10000),
			"wood":  int64(10000),
			"stone": int64(5000),
			"gold":  int64(2000),
		},
		Arrays: make(map[string]any),
		Dirty:  true,
	}
}

// Lock 加写锁
func (p *PlayerData) Lock() {
	p.mutex.Lock()
}

// Unlock 解写锁
func (p *PlayerData) Unlock() {
	p.mutex.Unlock()
}

// RLock 加读锁
func (p *PlayerData) RLock() {
	p.mutex.RLock()
}

// RUnlock 解读锁
func (p *PlayerData) RUnlock() {
	p.mutex.RUnlock()
}

// GetField 获取基础字段
func (p *PlayerData) GetField(key string) any {
	p.mutex.RLock()
	defer p.mutex.RUnlock()
	return p.Base[key]
}

// SetField 设置基础字段
func (p *PlayerData) SetField(key string, value any) {
	p.mutex.Lock()
	defer p.mutex.Unlock()
	p.Base[key] = value
	p.markDirtyLocked()
}

// GetArray 获取数组字段
func (p *PlayerData) GetArray(key string) any {
	p.mutex.RLock()
	defer p.mutex.RUnlock()
	return p.Arrays[key]
}

// SetArray 设置数组字段
func (p *PlayerData) SetArray(key string, value any) {
	p.mutex.Lock()
	defer p.mutex.Unlock()
	p.Arrays[key] = value
	p.markDirtyLocked()
}

func (p *PlayerData) MarkDirty() {
	p.mutex.Lock()
	defer p.mutex.Unlock()
	p.markDirtyLocked()
}

func (p *PlayerData) SnapshotDirtyVersion() (bool, uint64) {
	p.mutex.RLock()
	defer p.mutex.RUnlock()
	return p.Dirty, p.version
}

func (p *PlayerData) ClearDirtyIfVersion(version uint64) bool {
	p.mutex.Lock()
	defer p.mutex.Unlock()
	if p.version != version {
		return false
	}
	p.Dirty = false
	return true
}

func (p *PlayerData) markDirtyLocked() {
	p.Dirty = true
	p.version++
}
