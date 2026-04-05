package battle

import (
	"fmt"
	"sync"
)

// BattleLossConfig 战斗伤亡配置
type BattleLossConfig struct {
	BattleType         BattleType // 战斗类型
	IsAttacker         bool       // 是否攻击方
	SeriousInjuryRate int        // 重伤率 (千分比)
	DeathRate          int        // 死亡率 (千分比)
}

// SoldierConfig 士兵配置
type SoldierConfig struct {
	ID       int    // 士兵ID
	Name     string // 名称
	Power    int64  // 战力
	Attack   int64  // 攻击力
	Defense  int64  // 防御力
	HP       int64  // 生命值
	Speed    int64  // 速度
	Load     int64  // 负重
}

// 全局配置
var (
	battleLossConfigs = make(map[string]*BattleLossConfig) // key: "battleType_isAttacker"
	soldierConfigs    = make(map[int]*SoldierConfig)       // key: soldierID
	configMutex       sync.RWMutex
)

func init() {
	// 初始化默认配置
	initDefaultBattleLossConfigs()
	initDefaultSoldierConfigs()
}

// initDefaultBattleLossConfigs 初始化默认战斗伤亡配置
func initDefaultBattleLossConfigs() {
	// 打怪 - 攻击方
	addBattleLossConfig(BattleTypeMonster, true, 200, 50)
	// 打怪 - 防守方 (怪物)
	addBattleLossConfig(BattleTypeMonster, false, 1000, 1000)

	// 野蛮人城寨 - 攻击方
	addBattleLossConfig(BattleTypeMonsterCity, true, 250, 80)
	// 野蛮人城寨 - 防守方
	addBattleLossConfig(BattleTypeMonsterCity, false, 1000, 1000)

	// 资源点争夺 - 攻击方
	addBattleLossConfig(BattleTypeResource, true, 300, 100)
	// 资源点争夺 - 防守方
	addBattleLossConfig(BattleTypeResource, false, 400, 150)

	// 攻城 - 攻击方
	addBattleLossConfig(BattleTypeCity, true, 400, 150)
	// 攻城 - 防守方
	addBattleLossConfig(BattleTypeCity, false, 350, 120)
}

// initDefaultSoldierConfigs 初始化默认士兵配置
func initDefaultSoldierConfigs() {
	// 步兵
	addSoldierConfig(1001, "步兵", 10, 100, 80, 1000, 50, 100)
	// 骑兵
	addSoldierConfig(1002, "骑兵", 12, 120, 60, 800, 80, 80)
	// 弓兵
	addSoldierConfig(1003, "弓兵", 11, 110, 50, 600, 60, 60)
}

// addBattleLossConfig 添加战斗伤亡配置
func addBattleLossConfig(battleType BattleType, isAttacker bool, seriousRate, deathRate int) {
	configMutex.Lock()
	defer configMutex.Unlock()

	key := makeBattleLossKey(battleType, isAttacker)
	battleLossConfigs[key] = &BattleLossConfig{
		BattleType:         battleType,
		IsAttacker:         isAttacker,
		SeriousInjuryRate: seriousRate,
		DeathRate:          deathRate,
	}
}

// addSoldierConfig 添加士兵配置
func addSoldierConfig(id int, name string, power, attack, defense, hp, speed, load int64) {
	configMutex.Lock()
	defer configMutex.Unlock()

	soldierConfigs[id] = &SoldierConfig{
		ID:      id,
		Name:    name,
		Power:   power,
		Attack:  attack,
		Defense: defense,
		HP:      hp,
		Speed:   speed,
		Load:    load,
	}
}

// makeBattleLossKey 生成战斗伤亡配置key
func makeBattleLossKey(battleType BattleType, isAttacker bool) string {
	attacker := "0"
	if isAttacker {
		attacker = "1"
	}
	return fmt.Sprintf("%d_%s", int(battleType), attacker)
}

// GetBattleLossConfig 获取战斗伤亡配置
func GetBattleLossConfig(battleType BattleType, isAttacker bool) *BattleLossConfig {
	configMutex.RLock()
	defer configMutex.RUnlock()

	key := makeBattleLossKey(battleType, isAttacker)
	return battleLossConfigs[key]
}

// GetSoldierConfig 获取士兵配置
func GetSoldierConfig(soldierID int) *SoldierConfig {
	configMutex.RLock()
	defer configMutex.RUnlock()

	return soldierConfigs[soldierID]
}

// GetAllSoldierConfigs 获取所有士兵配置
func GetAllSoldierConfigs() map[int]*SoldierConfig {
	configMutex.RLock()
	defer configMutex.RUnlock()

	result := make(map[int]*SoldierConfig)
	for k, v := range soldierConfigs {
		result[k] = v
	}
	return result
}

// LoadBattleLossConfigs 从外部加载战斗伤亡配置
func LoadBattleLossConfigs(configs []BattleLossConfig) {
	configMutex.Lock()
	defer configMutex.Unlock()

	for _, cfg := range configs {
		key := makeBattleLossKey(cfg.BattleType, cfg.IsAttacker)
		battleLossConfigs[key] = &cfg
	}
}

// LoadSoldierConfigs 从外部加载士兵配置
func LoadSoldierConfigs(configs []SoldierConfig) {
	configMutex.Lock()
	defer configMutex.Unlock()

	for _, cfg := range configs {
		soldierConfigs[cfg.ID] = &cfg
	}
}
