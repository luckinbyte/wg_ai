package battle

import (
	"fmt"
)

// BattleType 战斗类型
type BattleType int

const (
	BattleTypeMonster     BattleType = 100 // 野蛮人
	BattleTypeMonsterCity BattleType = 101 // 野蛮人城寨
	BattleTypeResource    BattleType = 201 // 资源点争夺
	BattleTypeCity        BattleType = 202 // 城池攻防
)

// String 返回战斗类型名称
func (bt BattleType) String() string {
	switch bt {
	case BattleTypeMonster:
		return "monster"
	case BattleTypeMonsterCity:
		return "monster_city"
	case BattleTypeResource:
		return "resource"
	case BattleTypeCity:
		return "city"
	default:
		return fmt.Sprintf("unknown(%d)", int(bt))
	}
}

// BattleStatus 战斗状态
type BattleStatus int

const (
	BattleStatusPending  BattleStatus = iota // 等待中
	BattleStatusFighting                     // 战斗中
	BattleStatusFinished                     // 已结束
)

// String 返回战斗状态名称
func (bs BattleStatus) String() string {
	switch bs {
	case BattleStatusPending:
		return "pending"
	case BattleStatusFighting:
		return "fighting"
	case BattleStatusFinished:
		return "finished"
	default:
		return fmt.Sprintf("unknown(%d)", int(bs))
	}
}

// SideType 战斗方类型
type SideType string

const (
	SideTypePlayer  SideType = "player"
	SideTypeMonster SideType = "monster"
	SideTypeNPC     SideType = "npc"
)

// BattleSide 战斗方
type BattleSide struct {
	ID       int64       // 玩家ID/怪物ID
	Type     SideType    // 类型
	HeroID   int64       // 英雄ID (玩家方)
	Soldiers map[int]int // 士兵类型 -> 数量
	Power    int64       // 总战力

	// 伤亡
	MinorWound   map[int]int // 轻伤
	SeriousWound map[int]int // 重伤
	Death        map[int]int // 死亡
}

// NewBattleSide 创建战斗方
func NewBattleSide(id int64, sideType SideType) *BattleSide {
	return &BattleSide{
		ID:           id,
		Type:         sideType,
		Soldiers:     make(map[int]int),
		MinorWound:   make(map[int]int),
		SeriousWound: make(map[int]int),
		Death:        make(map[int]int),
	}
}

// SetSoldiers 设置士兵
func (bs *BattleSide) SetSoldiers(soldiers map[int]int) {
	bs.Soldiers = soldiers
}

// GetTotalSoldiers 获取总士兵数
func (bs *BattleSide) GetTotalSoldiers() int {
	total := 0
	for _, count := range bs.Soldiers {
		total += count
	}
	return total
}

// GetTotalDeaths 获取总死亡数
func (bs *BattleSide) GetTotalDeaths() int {
	total := 0
	for _, count := range bs.Death {
		total += count
	}
	return total
}

// GetTotalSeriousWound 获取总重伤数
func (bs *BattleSide) GetTotalSeriousWound() int {
	total := 0
	for _, count := range bs.SeriousWound {
		total += count
	}
	return total
}

// GetTotalMinorWound 获取总轻伤数
func (bs *BattleSide) GetTotalMinorWound() int {
	total := 0
	for _, count := range bs.MinorWound {
		total += count
	}
	return total
}

// ApplyCasualties 应用伤亡 (从士兵中扣除)
func (bs *BattleSide) ApplyCasualties() {
	// 死亡士兵直接移除
	for soldierType, count := range bs.Death {
		bs.Soldiers[soldierType] -= count
		if bs.Soldiers[soldierType] < 0 {
			bs.Soldiers[soldierType] = 0
		}
	}

	// 重伤和轻伤士兵保留,但标记状态
	// TODO: 后续可以在医院系统中处理
}

// ItemReward 物品奖励
type ItemReward struct {
	ItemID int64
	Count  int
}

// BattleRewards 战斗奖励
type BattleRewards struct {
	HeroExp int64        // 英雄经验
	Food    int64        // 粮食
	Wood    int64        // 木材
	Stone   int64        // 石头
	Gold    int64        // 金币
	Items   []ItemReward // 物品列表
}

// BattleResult 战斗结果
type BattleResult struct {
	ID        int64
	Type      BattleType
	Status    BattleStatus
	Attacker  *BattleSide
	Defender  *BattleSide
	Winner    string // "attacker" / "defender" / "draw"
	Duration  int    // 回合数
	StartTime int64
	EndTime   int64
	Rewards   *BattleRewards
}

// NewBattleResult 创建战斗结果
func NewBattleResult(battleType BattleType, attacker, defender *BattleSide) *BattleResult {
	return &BattleResult{
		ID:       0,
		Type:     battleType,
		Status:   BattleStatusPending,
		Attacker: attacker,
		Defender: defender,
		Duration: 0,
	}
}

// IsAttackerWin 攻击方是否获胜
func (br *BattleResult) IsAttackerWin() bool {
	return br.Winner == "attacker"
}

// IsDefenderWin 防守方是否获胜
func (br *BattleResult) IsDefenderWin() bool {
	return br.Winner == "defender"
}

// IsDraw 是否平局
func (br *BattleResult) IsDraw() bool {
	return br.Winner == "draw"
}
