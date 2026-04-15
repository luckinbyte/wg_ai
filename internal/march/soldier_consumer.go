package march

import (
	baseplugin "github.com/luckinbyte/wg_ai/plugin"
)

// DataAccessor 数据访问器别名
type DataAccessor = baseplugin.DataAccessor

// SoldierConsumer 由 baseplugin 包定义，士兵插件实现
type SoldierConsumer = baseplugin.SoldierConsumer

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
