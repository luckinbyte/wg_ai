package march

import (
	"fmt"
	"sync"
	"sync/atomic"
	"time"

	"github.com/luckinbyte/wg_ai/internal/battle"
	"github.com/luckinbyte/wg_ai/internal/scene"
)

// 全局军队ID生成器
var globalArmyID int64

// GenerateArmyID 生成唯一军队ID
func GenerateArmyID() int64 {
	return atomic.AddInt64(&globalArmyID, 1)
}

// Manager 行军管理器
type Manager struct {
	armies         map[int64]*Army   // armyID -> Army
	playerArmies   map[int64][]int64 // playerID -> []armyID
	marchingArmies map[int64]*Army   // 正在行军的军队
	mutex          sync.RWMutex

	sceneMgr       *scene.Manager
	walker         *WalkSimulator
	battleEng      *battle.Engine
	soldierConsumer SoldierConsumer  // 士兵消费者接口
}

// NewManager 创建行军管理器
func NewManager(sceneMgr *scene.Manager) *Manager {
	m := &Manager{
		armies:         make(map[int64]*Army),
		playerArmies:   make(map[int64][]int64),
		marchingArmies: make(map[int64]*Army),
		sceneMgr:       sceneMgr,
		battleEng:      battle.NewEngine(),
	}

	// 创建移动模拟器
	m.walker = NewWalkSimulator(m)

	return m
}

// SetSoldierConsumer 设置士兵消费者
func (m *Manager) SetSoldierConsumer(consumer SoldierConsumer) {
	m.mutex.Lock()
	defer m.mutex.Unlock()
	m.soldierConsumer = consumer
}

// Start 启动管理器
func (m *Manager) Start() {
	m.walker.Start()
}

// Stop 停止管理器
func (m *Manager) Stop() {
	m.walker.Stop()
}

// CreateArmy 创建军队
func (m *Manager) CreateArmy(ownerID, heroID int64, soldiers map[int]int, pos scene.Vector2, sceneID int64) (*Army, error) {
	m.mutex.Lock()
	defer m.mutex.Unlock()

	// 检查玩家军队数量上限
	playerArmies := m.playerArmies[ownerID]
	if len(playerArmies) >= 5 { // 最多5支军队
		return nil, fmt.Errorf("player %d has reached max army count", ownerID)
	}

	// 创建军队
	id := GenerateArmyID()
	army := NewArmy(id, ownerID, heroID, soldiers)
	army.Position = pos
	army.SceneID = sceneID

	// 存储军队
	m.armies[id] = army
	m.playerArmies[ownerID] = append(m.playerArmies[ownerID], id)

	return army, nil
}

// CreateArmyWithData 创建军队并消耗士兵 (带数据访问)
func (m *Manager) CreateArmyWithData(data DataAccessor, ownerID, heroID int64, soldiers map[int]int, pos scene.Vector2, sceneID int64) (*Army, error) {
	m.mutex.Lock()
	defer m.mutex.Unlock()

	// 检查士兵消费者
	if m.soldierConsumer == nil {
		return nil, fmt.Errorf("soldier consumer not set")
	}

	// 检查玩家军队数量上限
	playerArmies := m.playerArmies[ownerID]
	if len(playerArmies) >= 5 {
		return nil, fmt.Errorf("player %d has reached max army count", ownerID)
	}

	// 检查是否有足够士兵
	if !m.soldierConsumer.HasEnoughSoldiers(data, soldiers) {
		return nil, fmt.Errorf("not enough soldiers")
	}

	// 扣除士兵
	for soldierID, count := range soldiers {
		if err := m.soldierConsumer.SubSoldiers(data, soldierID, count); err != nil {
			// 回滚已扣除的士兵
			for sid, c := range soldiers {
				if sid == soldierID {
					break
				}
				m.soldierConsumer.AddSoldiers(data, sid, c)
			}
			return nil, fmt.Errorf("failed to consume soldiers: %v", err)
		}
	}

	// 创建军队
	id := GenerateArmyID()
	army := NewArmy(id, ownerID, heroID, soldiers)
	army.Position = pos
	army.SceneID = sceneID

	// 存储军队
	m.armies[id] = army
	m.playerArmies[ownerID] = append(m.playerArmies[ownerID], id)

	return army, nil
}

// DeleteArmy 删除军队
func (m *Manager) DeleteArmy(armyID int64) error {
	m.mutex.Lock()
	defer m.mutex.Unlock()

	army, exists := m.armies[armyID]
	if !exists {
		return fmt.Errorf("army %d not found", armyID)
	}

	// 不能删除行军中的军队
	if army.IsMarching() {
		return fmt.Errorf("army %d is marching, cannot delete", armyID)
	}

	// 从玩家军队列表移除
	m.removeFromPlayerArmies(army.OwnerID, armyID)

	// 删除军队
	delete(m.armies, armyID)
	delete(m.marchingArmies, armyID)

	return nil
}

// DeleteArmyWithData 删除军队并归还士兵
func (m *Manager) DeleteArmyWithData(data DataAccessor, armyID int64) error {
	m.mutex.Lock()
	defer m.mutex.Unlock()

	army, exists := m.armies[armyID]
	if !exists {
		return fmt.Errorf("army %d not found", armyID)
	}

	// 不能删除行军中的军队
	if army.IsMarching() {
		return fmt.Errorf("army %d is marching, cannot delete", armyID)
	}

	// 归还士兵
	if m.soldierConsumer != nil && len(army.Soldiers) > 0 {
		for soldierID, count := range army.Soldiers {
			if err := m.soldierConsumer.AddSoldiers(data, soldierID, count); err != nil {
				// 记录错误但继续删除军队
				fmt.Printf("warning: failed to return soldiers %d: %v\n", soldierID, err)
			}
		}
	}

	// 从玩家军队列表移除
	m.removeFromPlayerArmies(army.OwnerID, armyID)

	// 删除军队
	delete(m.armies, armyID)
	delete(m.marchingArmies, armyID)

	return nil
}

// GetArmy 获取军队
func (m *Manager) GetArmy(armyID int64) *Army {
	m.mutex.RLock()
	defer m.mutex.RUnlock()
	return m.armies[armyID]
}

// GetPlayerArmies 获取玩家所有军队
func (m *Manager) GetPlayerArmies(ownerID int64) []*Army {
	m.mutex.RLock()
	defer m.mutex.RUnlock()

	armyIDs := m.playerArmies[ownerID]
	armies := make([]*Army, 0, len(armyIDs))
	for _, id := range armyIDs {
		if army, ok := m.armies[id]; ok {
			armies = append(armies, army)
		}
	}
	return armies
}

// GetMarchingArmies 获取所有行军中的军队
func (m *Manager) GetMarchingArmies() []*Army {
	m.mutex.RLock()
	defer m.mutex.RUnlock()

	armies := make([]*Army, 0, len(m.marchingArmies))
	for _, army := range m.marchingArmies {
		armies = append(armies, army)
	}
	return armies
}

// StartMarch 开始行军
func (m *Manager) StartMarch(armyID int64, marchType MarchType, targetID int64) error {
	m.mutex.Lock()
	defer m.mutex.Unlock()

	army, exists := m.armies[armyID]
	if !exists {
		return fmt.Errorf("army %d not found", armyID)
	}

	// 检查军队状态
	if !army.CanMarch() {
		return fmt.Errorf("army %d cannot march, status: %s", armyID, army.Status)
	}

	// 获取场景
	s := m.sceneMgr.GetScene(army.SceneID)
	if s == nil {
		return fmt.Errorf("scene %d not found", army.SceneID)
	}

	// 获取目标
	target := s.GetObjectManager().GetObject(targetID)
	if target == nil {
		return fmt.Errorf("target %d not found", targetID)
	}

	// 计算路径 (简化: 直线路径)
	path := []scene.Vector2{army.Position, target.Position}

	// 计算行军速度
	speed := m.calcMarchSpeed(army)

	// 开始行军
	army.StartMarch(marchType, targetID, target.Position, path, speed)

	// 添加到行军列表
	m.marchingArmies[armyID] = army

	// 添加到移动模拟器
	m.walker.AddArmy(army)

	return nil
}

// CancelMarch 取消行军 (返回)
func (m *Manager) CancelMarch(armyID int64) error {
	m.mutex.Lock()
	defer m.mutex.Unlock()

	army, exists := m.armies[armyID]
	if !exists {
		return fmt.Errorf("army %d not found", armyID)
	}

	if !army.IsMarching() {
		return fmt.Errorf("army %d is not marching", armyID)
	}

	// 从移动模拟器移除
	m.walker.RemoveArmy(armyID)

	// 开始返回行军
	// TODO: 获取玩家城池位置
	returnPos := scene.Vector2{X: 100, Y: 100} // 临时: 固定返回点
	path := []scene.Vector2{army.Position, returnPos}
	speed := m.calcMarchSpeed(army)

	army.StartMarch(MarchTypeReturn, 0, returnPos, path, speed)
	m.walker.AddArmy(army)

	return nil
}

// ForceReturn 强制返回
func (m *Manager) ForceReturn(armyID int64) error {
	return m.CancelMarch(armyID)
}

// OnArmyArrival 军队到达回调 (由WalkSimulator调用)
func (m *Manager) OnArmyArrival(army *Army) {
	m.mutex.Lock()
	defer m.mutex.Unlock()

	if army.March == nil {
		return
	}

	switch army.March.Type {
	case MarchTypeCollect:
		m.handleCollectArrival(army)
	case MarchTypeAttack:
		m.handleAttackArrival(army)
	case MarchTypeReturn:
		m.handleReturnArrival(army)
	case MarchTypeReinforce:
		m.handleReinforceArrival(army)
	}
}

// handleCollectArrival 处理采集到达
func (m *Manager) handleCollectArrival(army *Army) {
	// 获取场景
	s := m.sceneMgr.GetScene(army.SceneID)
	if s == nil {
		army.FinishMarch()
		delete(m.marchingArmies, army.ID)
		return
	}

	// 获取资源点
	target := s.GetObjectManager().GetObject(army.March.TargetID)
	if target == nil {
		// 目标不存在，自动返回
		army.FinishMarch()
		delete(m.marchingArmies, army.ID)
		return
	}

	// 从行军列表移除
	delete(m.marchingArmies, army.ID)

	// 开始采集
	army.StartCollect(30 * time.Second) // 30秒采集时间

	// 更新资源点占领者
	objData := target.GetObjectData()
	if objData != nil {
		objData.OwnerID = army.OwnerID
	}
}

// handleAttackArrival 处理攻击到达
func (m *Manager) handleAttackArrival(army *Army) {
	// 获取场景
	s := m.sceneMgr.GetScene(army.SceneID)
	if s == nil {
		army.FinishMarch()
		delete(m.marchingArmies, army.ID)
		return
	}

	// 获取目标
	target := s.GetObjectManager().GetObject(army.March.TargetID)
	if target == nil {
		// 目标不存在，自动返回
		army.FinishMarch()
		delete(m.marchingArmies, army.ID)
		return
	}

	// 创建攻击方战斗数据
	attacker := battle.NewBattleSide(army.OwnerID, battle.SideTypePlayer)
	attacker.HeroID = army.HeroID
	attacker.SetSoldiers(army.Soldiers) // 直接使用军队的士兵配置

	// 创建防守方战斗数据
	defender := battle.NewBattleSide(target.ID, battle.SideTypeMonster)
	objData := target.GetObjectData()
	if objData != nil {
		// 根据目标等级计算怪物兵力
		monsterCount := objData.Level * 100
		defender.SetSoldiers(map[int]int{
			101: monsterCount, // 怪物使用步兵
		})
	} else {
		defender.SetSoldiers(map[int]int{
			101: 500, // 默认500士兵
		})
	}

	// 确定战斗类型
	battleType := battle.BattleTypeMonster
	if target.Type == scene.EntityTypeBuilding {
		battleType = battle.BattleTypeMonsterCity
	}

	// 执行战斗
	result := m.battleEng.StartBattle(battleType, attacker, defender, nil, nil)

	// 生成战报
	report := battle.GenerateReport(result)

	// 记录战报 (TODO: 存储到数据库或发送给客户端)
	reportJSON, _ := report.ToJSON()
	fmt.Printf("Battle Report: %s\n", string(reportJSON))

	// 处理战斗伤亡
	m.processBattleCasualties(army, result)

	// 处理战斗结果
	if result.IsAttackerWin() {
		// 攻击方胜利
		// TODO: 发放奖励到玩家背包
		if result.Rewards != nil {
			army.LoadFood += result.Rewards.Food
			army.LoadWood += result.Rewards.Wood
			army.LoadStone += result.Rewards.Stone
			army.LoadGold += result.Rewards.Gold
		}

		// 移除目标 (怪物被消灭)
		s.GetObjectManager().RemoveObject(army.March.TargetID)
	}

	// 检查军队是否还有士兵
	if army.GetTotalSoldiers() <= 0 {
		// 全军覆没,删除军队
		m.removeFromPlayerArmies(army.OwnerID, army.ID)
		delete(m.armies, army.ID)
		delete(m.marchingArmies, army.ID)
		return
	}

	// 结束行军
	army.FinishMarch()
	delete(m.marchingArmies, army.ID)

	// TODO: 发送战斗结果给客户端
}

// handleReturnArrival 处理返回到达
func (m *Manager) handleReturnArrival(army *Army) {
	// 返回完成
	army.FinishMarch()
	delete(m.marchingArmies, army.ID)

	// TODO: 将携带的资源加入玩家背包
}

// handleReinforceArrival 处理支援到达
func (m *Manager) handleReinforceArrival(army *Army) {
	// TODO: 实现支援逻辑
	army.FinishMarch()
	delete(m.marchingArmies, army.ID)
}

// processBattleCasualties 处理战斗伤亡
// 使用战斗引擎计算的伤亡结果
func (m *Manager) processBattleCasualties(army *Army, result *battle.BattleResult) {
	if result.Attacker == nil {
		return
	 }

        // 处理每种士兵的伤亡
        for soldierID, originalCount := range army.Soldiers {
                // 获取攻击方伤亡数据
                deathCount := result.Attacker.Death[soldierID]
                woundCount := result.Attacker.SeriousWound[soldierID]

                if deathCount+woundCount > originalCount {
                    deathCount = originalCount
                    woundCount = 0
                }

                remaining := originalCount - deathCount - woundCount
                if remaining < 0 {
                    remaining = 0
                }

                // 更新军队士兵数量
                if remaining > 0 {
                    army.Soldiers[soldierID] = remaining
                } else {
                    delete(army.Soldiers, soldierID)
                }

                // 记录伤亡信息 (用于战报)
                fmt.Printf("Soldier %d: original=%d, died=%d, wounded=%d, remaining=%d\n",
                        soldierID, originalCount, deathCount, woundCount, remaining)
        }
}

// removeFromPlayerArmies 从玩家军队列表移除
func (m *Manager) removeFromPlayerArmies(ownerID, armyID int64) {
	armyIDs := m.playerArmies[ownerID]
	for i, id := range armyIDs {
		if id == armyID {
			m.playerArmies[ownerID] = append(armyIDs[:i], armyIDs[i+1:]...)
			break
		}
	}
}

// calcMarchSpeed 计算行军速度
func (m *Manager) calcMarchSpeed(army *Army) float64 {
	// 基础速度: 100 单位/秒
	baseSpeed := 100.0

	// TODO: 加入科技加成、英雄加成等

	return baseSpeed
}

// ArmyCount 获取军队总数
func (m *Manager) ArmyCount() int {
	m.mutex.RLock()
	defer m.mutex.RUnlock()
	return len(m.armies)
}

// MarchingCount 获取行军中军队数量
func (m *Manager) MarchingCount() int {
	m.mutex.RLock()
	defer m.mutex.RUnlock()
	return len(m.marchingArmies)
}

// GetStats 获取统计信息
func (m *Manager) GetStats() map[string]any {
	m.mutex.RLock()
	defer m.mutex.RUnlock()

	return map[string]any{
		"army_count":     len(m.armies),
		"marching_count": len(m.marchingArmies),
		"player_count":   len(m.playerArmies),
	}
}
