package soldier

import (
	"fmt"
	"time"

	baseplugin "github.com/luckinbyte/wg_ai/plugin"
)

// ============ 医院系统 ============

// GetWounded 获取所有伤兵
func (m *Manager) GetWounded(data baseplugin.DataAccessor) (map[int]*SoldierData, error) {
	soldiers, err := m.GetSoldiers(data)
	if err != nil {
		return nil, err
	}
	wounded := make(map[int]*SoldierData)
	for id, s := range soldiers {
		if s.Wounded > 0 {
			wounded[id] = s
		}
	}
	return wounded, nil
}

// StartHeal 开始治疗
func (m *Manager) StartHeal(data baseplugin.DataAccessor, soldiers map[int]int) (*HealQueueItem, error) {
	// 1. 验证伤兵存在
	wounded, err := m.GetWounded(data)
	if err != nil {
		return nil, err
	}

	for soldierID, count := range soldiers {
		w, ok := wounded[soldierID]
		if !ok || w.Wounded < count {
			have := 0
			if ok {
				have = w.Wounded
			}
			return nil, fmt.Errorf("not enough wounded soldier %d, have %d, need %d",
				soldierID, have, count)
		}
	}

	// 2. 检查是否正在治疗
	m.queuesMutex.Lock()
	defer m.queuesMutex.Unlock()

	rid := m.getRIDFromData(data)
	if healQueue, ok := m.healQueues[rid]; ok && healQueue.Current != nil {
		if healQueue.Current.FinishTime > time.Now().Unix() {
			return nil, fmt.Errorf("healing in progress, wait for completion")
		}
	}

	// 3. 计算治疗消耗和时间
	var totalFood, totalWood int64
	var totalTime int

	for soldierID, count := range soldiers {
		cfg := GetSoldierConfig(soldierID)
		if cfg != nil {
			totalFood += cfg.HealFood * int64(count)
			totalWood += cfg.HealWood * int64(count)
			totalTime += cfg.HealTime * count
		}
	}

	// 4. 检查并扣除资源
	food := m.getResource(data, "food")
	wood := m.getResource(data, "wood")

	if food < totalFood {
		return nil, fmt.Errorf("not enough food for healing, have %d, need %d", food, totalFood)
	}
	if wood < totalWood {
		return nil, fmt.Errorf("not enough wood for healing, have %d, need %d", wood, totalWood)
	}

	// 扣除资源
	m.setResource(data, "food", food-totalFood)
	m.setResource(data, "wood", wood-totalWood)

	// 5. 从伤兵池移除
	for soldierID, count := range soldiers {
		if err := m.SubWounded(data, soldierID, count); err != nil {
			// 返还资源
			m.setResource(data, "food", food)
			m.setResource(data, "wood", wood)
			return nil, err
		}
	}

	// 6. 创建治疗队列
	now := time.Now().Unix()
	healQueue := &HealQueueItem{
		ID:         generateQueueID(),
		Soldiers:   soldiers,
		StartTime:  now,
		FinishTime: now + int64(totalTime),
	}

	// 7. 保存队列
	healData := &HealQueueData{Current: healQueue}
	m.healQueues[rid] = healData

	data.MarkDirty()
	return healQueue, nil
}

// CancelHeal 取消治疗
func (m *Manager) CancelHeal(data baseplugin.DataAccessor) error {
	m.queuesMutex.Lock()
	defer m.queuesMutex.Unlock()

	rid := m.getRIDFromData(data)
	healData, ok := m.healQueues[rid]
	if !ok || healData.Current == nil {
		return fmt.Errorf("no healing in progress")
	}

	// 1. 返还伤兵
	for soldierID, count := range healData.Current.Soldiers {
		m.AddWounded(data, soldierID, count)
	}

	// 2. 返还50%资源
	var totalFood, totalWood int64
	for soldierID, count := range healData.Current.Soldiers {
		cfg := GetSoldierConfig(soldierID)
		if cfg != nil {
			totalFood += cfg.HealFood * int64(count) / 2
			totalWood += cfg.HealWood * int64(count) / 2
		}
	}

	food := m.getResource(data, "food")
	wood := m.getResource(data, "wood")
	m.setResource(data, "food", food+totalFood)
	m.setResource(data, "wood", wood+totalWood)

	// 3. 清除队列
	healData.Current = nil
	data.MarkDirty()

	return nil
}

// CompleteHeal 完成治疗
func (m *Manager) CompleteHeal(data baseplugin.DataAccessor) (*HealQueueItem, error) {
	m.queuesMutex.Lock()
	defer m.queuesMutex.Unlock()

	rid := m.getRIDFromData(data)
	healData, ok := m.healQueues[rid]
	if !ok || healData.Current == nil {
		return nil, fmt.Errorf("no healing in progress")
	}

	if healData.Current.FinishTime > time.Now().Unix() {
		return nil, fmt.Errorf("healing not completed yet")
	}

	// 1. 将士兵恢复为健康状态
	for soldierID, count := range healData.Current.Soldiers {
		if err := m.AddSoldiers(data, soldierID, count); err != nil {
			return nil, err
		}
	}

	// 2. 保存完成的队列信息
	completed := healData.Current

	// 3. 清除队列
	healData.Current = nil
	data.MarkDirty()

	return completed, nil
}

// GetHealQueue 获取治疗队列
func (m *Manager) GetHealQueue(data baseplugin.DataAccessor) (*HealQueueItem, error) {
	m.queuesMutex.RLock()
	defer m.queuesMutex.RUnlock()

	rid := m.getRIDFromData(data)
	if healData, ok := m.healQueues[rid]; ok {
		return healData.Current, nil
	}
	return nil, nil
}

// IsHealing 是否正在治疗
func (m *Manager) IsHealing(data baseplugin.DataAccessor) bool {
	m.queuesMutex.RLock()
	defer m.queuesMutex.RUnlock()

	rid := m.getRIDFromData(data)
	if healData, ok := m.healQueues[rid]; ok && healData.Current != nil {
		return healData.Current.FinishTime > time.Now().Unix()
	}
	return false
}

// GetHealProgress 获取治疗进度 (0.0 - 1.0)
func (m *Manager) GetHealProgress(data baseplugin.DataAccessor) float64 {
	m.queuesMutex.RLock()
	defer m.queuesMutex.RUnlock()

	rid := m.getRIDFromData(data)
	if healData, ok := m.healQueues[rid]; ok && healData.Current != nil {
		now := time.Now().Unix()
		if now >= healData.Current.FinishTime {
			return 1.0
		}
		if healData.Current.StartTime >= healData.Current.FinishTime {
			return 1.0
		}
		elapsed := now - healData.Current.StartTime
		total := healData.Current.FinishTime - healData.Current.StartTime
		return float64(elapsed) / float64(total)
	}
	return 0
}

// ============ 战斗伤亡处理 ============

// ProcessBattleCasualties 处理战斗伤亡
// deathRate: 死亡比例, woundRate: 重伤比例, minorRate: 轻伤比例
func (m *Manager) ProcessBattleCasualties(data baseplugin.DataAccessor, soldiers map[int]int,
	deathRate, woundRate, minorRate float64) (death, wounded map[int]int, err error) {

	death = make(map[int]int)
	wounded = make(map[int]int)

	for soldierID, count := range soldiers {
		// 计算各类型伤亡
		deathCount := int(float64(count) * deathRate)
		woundCount := int(float64(count) * woundRate)
		minorCount := int(float64(count) * minorRate)

		// 确保不超过总数
		if deathCount+woundCount+minorCount > count {
			deathCount = count / 3
			woundCount = count / 3
			minorCount = count - deathCount - woundCount
		}

		// 记录死亡
		if deathCount > 0 {
			death[soldierID] = deathCount
		}

		// 记录重伤
		if woundCount > 0 {
			wounded[soldierID] = woundCount
			m.AddWounded(data, soldierID, woundCount)
		}

		// 轻伤: 直接从健康士兵中扣除，但保留在军队中
		// (这里简化处理，轻伤也进入医院)
		if minorCount > 0 {
			wounded[soldierID] = woundCount + minorCount
			m.AddWounded(data, soldierID, minorCount)
		}

		// 扣除总损失
		totalLoss := deathCount + woundCount + minorCount
		if totalLoss > 0 {
			if err := m.SubSoldiers(data, soldierID, totalLoss); err != nil {
				return nil, nil, err
			}
		}
	}

	return death, wounded, nil
}

// AddSoldiers 批量增加士兵 (用于测试或GM命令)
func (m *Manager) AddSoldiersBatch(data baseplugin.DataAccessor, soldiers map[int]int) error {
	for soldierID, count := range soldiers {
		if err := m.AddSoldiers(data, soldierID, count); err != nil {
			return err
		}
	}
	return nil
}
