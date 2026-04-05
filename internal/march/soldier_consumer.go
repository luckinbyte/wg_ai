package march

import (
	baseplugin "github.com/yourorg/wg_ai/plugin"
)

// DataAccessor 数据访问器别名
type DataAccessor = baseplugin.DataAccessor

// SoldierConsumer 士兵消费者接口
// 由士兵插件实现，供军队模块调用
type SoldierConsumer interface {
	// HasEnoughSoldiers 检查是否有足够士兵
	HasEnoughSoldiers(data DataAccessor, required map[int]int) bool

	// SubSoldiers 扣除士兵 (创建军队时调用)
	SubSoldiers(data DataAccessor, soldierID, count int) error

	// AddSoldiers 归还士兵 (解散军队时调用)
	AddSoldiers(data DataAccessor, soldierID, count int) error

	// AddWounded 添加伤兵 (战斗伤亡时调用)
	AddWounded(data DataAccessor, soldierID, count int) error

	// GetSoldierCount 获取士兵数量
	GetSoldierCount(data DataAccessor, soldierID int) (int, error)
}

// SoldierCasualty 伤亡信息
type SoldierCasualty struct {
	SoldierID int // 士兵ID
	Died      int // 死亡数量
	Wounded   int // 受伤数量
}

// BattleCasualties 战斗伤亡结果
type BattleCasualties struct {
	Attacker []SoldierCasualty // 攻击方伤亡
	Defender []SoldierCasualty // 防守方伤亡
}
