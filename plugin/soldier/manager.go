package main

import (
	"encoding/json"
	"fmt"
	"sync"
	"time"

	baseplugin "github.com/yourorg/wg_ai/plugin"
)

// ResourceCost 资源消耗
type ResourceCost struct {
	Food  int64
	Wood  int64
	Stone int64
	Gold  int64
}

// TrainQueueItem 训练队列项
type TrainQueueItem struct {
	ID          int64 `json:"id"`
	SoldierID   int   `json:"soldier_id"`
	SoldierType int   `json:"soldier_type"`
	Level       int   `json:"level"`
	Count       int   `json:"count"`
	StartTime   int64 `json:"start_time"`
	FinishTime  int64 `json:"finish_time"`
	IsUpgrade   bool  `json:"is_upgrade"`
}

// TrainQueueData 训练队列存储
type TrainQueueData struct {
	Items []TrainQueueItem `json:"items"`
}

// HealQueueItem 治疗队列项
type HealQueueItem struct {
	ID         int64       `json:"id"`
	Soldiers   map[int]int `json:"soldiers"`
	StartTime  int64       `json:"start_time"`
	FinishTime int64       `json:"finish_time"`
}

// HealQueueData 治疗队列存储
type HealQueueData struct {
	Current *HealQueueItem `json:"current"`
}

// Manager 士兵管理器
type Manager struct {
	trainQueues map[int64]*TrainQueueData
	healQueues  map[int64]*HealQueueData
	queuesMutex sync.RWMutex

	trainTicker *time.Ticker
	healTicker  *time.Ticker
	stopCh      chan struct{}
	running     bool
	mutex       sync.Mutex
}

// NewManager 创建士兵管理器
func NewManager() *Manager {
	return &Manager{
		trainQueues: make(map[int64]*TrainQueueData),
		healQueues:  make(map[int64]*HealQueueData),
		stopCh:      make(chan struct{}),
	}
}

// Start 启动管理器
func (m *Manager) Start() {
	m.mutex.Lock()
	defer m.mutex.Unlock()
	if m.running {
		return
	}
	m.running = true
	m.trainTicker = time.NewTicker(1 * time.Second)
	m.healTicker = time.NewTicker(1 * time.Second)
	go m.tickLoop()
}

// Stop 停止管理器
func (m *Manager) Stop() {
	m.mutex.Lock()
	defer m.mutex.Unlock()
	if !m.running {
		return
	}
	m.running = false
	close(m.stopCh)
	if m.trainTicker != nil {
		m.trainTicker.Stop()
	}
	if m.healTicker != nil {
		m.healTicker.Stop()
	}
}

func (m *Manager) tickLoop() {
	for {
		select {
		case <-m.stopCh:
			return
		case <-m.trainTicker.C:
		case <-m.healTicker.C:
		}
	}
}

// ============ 士兵存取 ============

// GetSoldiers 获取玩家所有士兵
func (m *Manager) GetSoldiers(data baseplugin.DataAccessor) (map[int]*SoldierData, error) {
	raw, err := data.GetArray("soldiers")
	if err != nil || raw == nil {
		return make(map[int]*SoldierData), nil
	}

	if soldiers, ok := raw.(map[int]*SoldierData); ok {
		return soldiers, nil
	}

	if rawMap, ok := raw.(map[string]any); ok {
		soldiers := make(map[int]*SoldierData)
		for k, v := range rawMap {
			var id int
			fmt.Sscanf(k, "%d", &id)
			if bytes, err := json.Marshal(v); err == nil {
				var sd SoldierData
				if json.Unmarshal(bytes, &sd) == nil {
					soldiers[id] = &sd
				}
			}
		}
		return soldiers, nil
	}

	return make(map[int]*SoldierData), nil
}

// GetSoldierCount 获取指定士兵数量
func (m *Manager) GetSoldierCount(data baseplugin.DataAccessor, soldierID int) (int, error) {
	soldiers, err := m.GetSoldiers(data)
	if err != nil {
		return 0, err
	}
	if s, ok := soldiers[soldierID]; ok {
		return s.Count, nil
	}
	return 0, nil
}

// AddSoldiers 增加士兵
func (m *Manager) AddSoldiers(data baseplugin.DataAccessor, soldierID, count int) error {
	soldiers, err := m.GetSoldiers(data)
	if err != nil {
		return err
	}

	soldierType, level := ParseSoldierID(soldierID)
	if s, ok := soldiers[soldierID]; ok {
		s.Count += count
	} else {
		soldiers[soldierID] = &SoldierData{
			ID:    soldierID,
			Type:  soldierType,
			Level: level,
			Count: count,
		}
	}
	data.MarkDirty()
	return nil
}

// SubSoldiers 减少士兵
func (m *Manager) SubSoldiers(data baseplugin.DataAccessor, soldierID, count int) error {
	soldiers, err := m.GetSoldiers(data)
	if err != nil {
		return err
	}

	s, ok := soldiers[soldierID]
	if !ok {
		return fmt.Errorf("soldier %d not found", soldierID)
	}
	if s.Count < count {
		return fmt.Errorf("not enough soldiers, have %d, need %d", s.Count, count)
	}

	s.Count -= count
	if s.Count <= 0 && s.Wounded <= 0 {
		delete(soldiers, soldierID)
	}
	data.MarkDirty()
	return nil
}

// HasEnoughSoldiers 检查是否有足够士兵
func (m *Manager) HasEnoughSoldiers(data baseplugin.DataAccessor, required map[int]int) bool {
	soldiers, err := m.GetSoldiers(data)
	if err != nil {
		return false
	}
	for soldierID, count := range required {
		if s, ok := soldiers[soldierID]; !ok || s.Count < count {
			return false
		}
	}
	return true
}

// ============ 伤兵管理 ============

// AddWounded 添加伤兵
func (m *Manager) AddWounded(data baseplugin.DataAccessor, soldierID, count int) error {
	soldiers, err := m.GetSoldiers(data)
	if err != nil {
		return err
	}

	soldierType, level := ParseSoldierID(soldierID)
	if s, ok := soldiers[soldierID]; ok {
		s.Wounded += count
	} else {
		soldiers[soldierID] = &SoldierData{
			ID:      soldierID,
			Type:    soldierType,
			Level:   level,
			Wounded: count,
		}
	}
	data.MarkDirty()
	return nil
}

// SubWounded 减少伤兵
func (m *Manager) SubWounded(data baseplugin.DataAccessor, soldierID, count int) error {
	soldiers, err := m.GetSoldiers(data)
	if err != nil {
		return err
	}

	s, ok := soldiers[soldierID]
	if !ok {
		return fmt.Errorf("wounded soldier %d not found", soldierID)
	}
	if s.Wounded < count {
		return fmt.Errorf("not enough wounded, have %d, need %d", s.Wounded, count)
	}

	s.Wounded -= count
	if s.Count <= 0 && s.Wounded <= 0 {
		delete(soldiers, soldierID)
	}
	data.MarkDirty()
	return nil
}

// ============ 统计 ============

// GetTotalPower 获取总战力
func (m *Manager) GetTotalPower(data baseplugin.DataAccessor) (int64, error) {
	soldiers, err := m.GetSoldiers(data)
	if err != nil {
		return 0, err
	}
	var total int64
	for _, s := range soldiers {
		if cfg := GetSoldierConfig(s.ID); cfg != nil {
			total += cfg.Power * int64(s.Count)
		}
	}
	return total, nil
}

// GetTotalCount 获取士兵总数
func (m *Manager) GetTotalCount(data baseplugin.DataAccessor) (int, error) {
	soldiers, err := m.GetSoldiers(data)
	if err != nil {
		return 0, err
	}
	var total int
	for _, s := range soldiers {
		total += s.Count
	}
	return total, nil
}

// GetTotalWounded 获取伤兵总数
func (m *Manager) GetTotalWounded(data baseplugin.DataAccessor) (int, error) {
	soldiers, err := m.GetSoldiers(data)
	if err != nil {
		return 0, err
	}
	var total int
	for _, s := range soldiers {
		total += s.Wounded
	}
	return total, nil
}

// ============ 辅助方法 ============

func (m *Manager) getRIDFromData(data baseplugin.DataAccessor) int64 {
	if ridAny, err := data.GetField("rid"); err == nil {
		switch v := ridAny.(type) {
		case int64:
			return v
		case int:
			return int64(v)
		case float64:
			return int64(v)
		}
	}
	return 0
}

func (m *Manager) getResource(data baseplugin.DataAccessor, key string) int64 {
	if val, err := data.GetField(key); err == nil {
		switch v := val.(type) {
		case int64:
			return v
		case int:
			return int64(v)
		case float64:
			return int64(v)
		}
	}
	return 0
}

func (m *Manager) setResource(data baseplugin.DataAccessor, key string, val int64) {
	data.SetField(key, val)
}

func (m *Manager) getOrCreateTrainQueue(rid int64) *TrainQueueData {
	if q, ok := m.trainQueues[rid]; ok {
		return q
	}
	q := &TrainQueueData{Items: []TrainQueueItem{}}
	m.trainQueues[rid] = q
	return q
}
