package main

import (
	"fmt"
	"sync/atomic"
	"time"

	baseplugin "github.com/luckinbyte/wg_ai/plugin"
)

// 全局队列ID生成器
var globalQueueID int64

func generateQueueID() int64 {
	return atomic.AddInt64(&globalQueueID, 1)
}

// StartTrain 开始训练
func (m *Manager) StartTrain(data baseplugin.DataAccessor, soldierType, level, count int, isUpgrade bool) (*TrainQueueItem, error) {
	soldierID := MakeSoldierID(soldierType, level)
	cfg := GetSoldierConfig(soldierID)
	if cfg == nil {
		return nil, fmt.Errorf("soldier config not found: %d", soldierID)
	}

	m.queuesMutex.Lock()
	defer m.queuesMutex.Unlock()

	rid := m.getRIDFromData(data)
	queueData := m.getOrCreateTrainQueue(rid)

	// 检查该兵种是否已有训练
	now := time.Now().Unix()
	for _, item := range queueData.Items {
		if item.SoldierType == soldierType && item.FinishTime > now {
			return nil, fmt.Errorf("soldier type %d already training", soldierType)
		}
	}

	// 计算消耗
	cost := m.calcTrainCost(cfg, count, isUpgrade)

	// 检查并扣除资源
	if err := m.checkAndCostResources(data, cost); err != nil {
		return nil, err
	}

	// 如果是晋升，扣除低级士兵
	if isUpgrade {
		lowerLevel := level - 1
		if lowerLevel < 1 {
			m.refundResources(data, cost)
			return nil, fmt.Errorf("cannot upgrade level 1 soldiers")
		}
		lowerID := MakeSoldierID(soldierType, lowerLevel)
		if err := m.SubSoldiers(data, lowerID, count); err != nil {
			m.refundResources(data, cost)
			return nil, err
		}
	}

	// 创建训练队列项
	trainTime := int64(cfg.TrainTime) * int64(count)
	item := &TrainQueueItem{
		ID:          generateQueueID(),
		SoldierID:   soldierID,
		SoldierType: soldierType,
		Level:       level,
		Count:       count,
		StartTime:   now,
		FinishTime:  now + trainTime,
		IsUpgrade:   isUpgrade,
	}

	queueData.Items = append(queueData.Items, *item)
	return item, nil
}

// CancelTrain 取消训练
func (m *Manager) CancelTrain(data baseplugin.DataAccessor, queueID int64) error {
	m.queuesMutex.Lock()
	defer m.queuesMutex.Unlock()

	rid := m.getRIDFromData(data)
	queueData, ok := m.trainQueues[rid]
	if !ok {
		return fmt.Errorf("train queue not found")
	}

	var targetIndex int = -1
	for i, item := range queueData.Items {
		if item.ID == queueID {
			targetIndex = i
			cfg := GetSoldierConfig(item.SoldierID)
			if cfg != nil {
				cost := m.calcTrainCost(cfg, item.Count, item.IsUpgrade)
				cost.Food /= 2
				cost.Wood /= 2
				cost.Stone /= 2
				cost.Gold /= 2
				m.refundResources(data, cost)

				if item.IsUpgrade {
					lowerID := MakeSoldierID(item.SoldierType, item.Level-1)
					m.AddSoldiers(data, lowerID, item.Count)
				}
			}
			break
		}
	}

	if targetIndex == -1 {
		return fmt.Errorf("train item %d not found", queueID)
	}

	queueData.Items = append(queueData.Items[:targetIndex], queueData.Items[targetIndex+1:]...)
	return nil
}

// GetTrainQueue 获取训练队列
func (m *Manager) GetTrainQueue(data baseplugin.DataAccessor) ([]TrainQueueItem, error) {
	m.queuesMutex.RLock()
	defer m.queuesMutex.RUnlock()

	rid := m.getRIDFromData(data)
	if queueData, ok := m.trainQueues[rid]; ok {
		return queueData.Items, nil
	}
	return []TrainQueueItem{}, nil
}

// CompleteTrain 完成训练
func (m *Manager) CompleteTrain(data baseplugin.DataAccessor, queueID int64) (*TrainQueueItem, error) {
	m.queuesMutex.Lock()
	defer m.queuesMutex.Unlock()

	rid := m.getRIDFromData(data)
	queueData, ok := m.trainQueues[rid]
	if !ok {
		return nil, fmt.Errorf("train queue not found")
	}

	now := time.Now().Unix()
	var targetIndex int = -1
	var targetItem TrainQueueItem

	for i, item := range queueData.Items {
		if item.ID == queueID && item.FinishTime <= now {
			targetIndex = i
			targetItem = item
			break
		}
	}

	if targetIndex == -1 {
		return nil, fmt.Errorf("train item %d not completed or not found", queueID)
	}

	if err := m.AddSoldiers(data, targetItem.SoldierID, targetItem.Count); err != nil {
		return nil, err
	}

	queueData.Items = append(queueData.Items[:targetIndex], queueData.Items[targetIndex+1:]...)
	return &targetItem, nil
}

// GetCompletedTrains 获取已完成的训练
func (m *Manager) GetCompletedTrains(data baseplugin.DataAccessor) []TrainQueueItem {
	m.queuesMutex.RLock()
	defer m.queuesMutex.RUnlock()

	now := time.Now().Unix()
	var completed []TrainQueueItem

	rid := m.getRIDFromData(data)
	if queueData, ok := m.trainQueues[rid]; ok {
		for _, item := range queueData.Items {
			if item.FinishTime <= now {
				completed = append(completed, item)
			}
		}
	}
	return completed
}

// GetTrainProgress 获取训练进度
func (m *Manager) GetTrainProgress(data baseplugin.DataAccessor, queueID int64) float64 {
	m.queuesMutex.RLock()
	defer m.queuesMutex.RUnlock()

	rid := m.getRIDFromData(data)
	if queueData, ok := m.trainQueues[rid]; ok {
		for _, item := range queueData.Items {
			if item.ID == queueID {
				now := time.Now().Unix()
				if now >= item.FinishTime {
					return 1.0
				}
				if item.StartTime >= item.FinishTime {
					return 1.0
				}
				elapsed := now - item.StartTime
				total := item.FinishTime - item.StartTime
				if total <= 0 {
					return 1.0
				}
				return float64(elapsed) / float64(total)
			}
		}
	}
	return 0
}

// ============ 辅助方法 ============

func (m *Manager) calcTrainCost(cfg *SoldierConfig, count int, isUpgrade bool) *ResourceCost {
	cost := &ResourceCost{
		Food:  cfg.CostFood * int64(count),
		Wood:  cfg.CostWood * int64(count),
		Stone: cfg.CostStone * int64(count),
		Gold:  cfg.CostGold * int64(count),
	}

	if isUpgrade {
		if lowerCfg := GetSoldierConfigByType(cfg.Type, cfg.Level-1); lowerCfg != nil {
			cost.Food -= lowerCfg.CostFood * int64(count)
			cost.Wood -= lowerCfg.CostWood * int64(count)
			cost.Stone -= lowerCfg.CostStone * int64(count)
			cost.Gold -= lowerCfg.CostGold * int64(count)
		}
	}

	if cost.Food < 0 {
		cost.Food = 0
	}
	if cost.Wood < 0 {
		cost.Wood = 0
	}
	if cost.Stone < 0 {
		cost.Stone = 0
	}
	if cost.Gold < 0 {
		cost.Gold = 0
	}

	return cost
}

func (m *Manager) checkAndCostResources(data baseplugin.DataAccessor, cost *ResourceCost) error {
	food := m.getResource(data, "food")
	wood := m.getResource(data, "wood")
	stone := m.getResource(data, "stone")
	gold := m.getResource(data, "gold")

	if food < cost.Food {
		return fmt.Errorf("not enough food, have %d, need %d", food, cost.Food)
	}
	if wood < cost.Wood {
		return fmt.Errorf("not enough wood, have %d, need %d", wood, cost.Wood)
	}
	if stone < cost.Stone {
		return fmt.Errorf("not enough stone, have %d, need %d", stone, cost.Stone)
	}
	if gold < cost.Gold {
		return fmt.Errorf("not enough gold, have %d, need %d", gold, cost.Gold)
	}

	m.setResource(data, "food", food-cost.Food)
	m.setResource(data, "wood", wood-cost.Wood)
	m.setResource(data, "stone", stone-cost.Stone)
	m.setResource(data, "gold", gold-cost.Gold)
	data.MarkDirty()

	return nil
}

func (m *Manager) refundResources(data baseplugin.DataAccessor, cost *ResourceCost) {
	food := m.getResource(data, "food")
	wood := m.getResource(data, "wood")
	stone := m.getResource(data, "stone")
	gold := m.getResource(data, "gold")

	m.setResource(data, "food", food+cost.Food)
	m.setResource(data, "wood", wood+cost.Wood)
	m.setResource(data, "stone", stone+cost.Stone)
	m.setResource(data, "gold", gold+cost.Gold)
	data.MarkDirty()
}
