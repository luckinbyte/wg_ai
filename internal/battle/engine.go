package battle

import (
	"math/rand"
	"sync"
	"time"
)

// Engine 战斗引擎
type Engine struct {
	rand   *rand.Rand
	mutex  sync.Mutex
}

// NewEngine 创建战斗引擎
func NewEngine() *Engine {
	return &Engine{
		rand: rand.New(rand.NewSource(time.Now().UnixNano())),
	}
}

// StartBattle 开始战斗
func (e *Engine) StartBattle(battleType BattleType, attacker, defender *BattleSide, attackerBuffs, defenderBuffs map[string]float64) *BattleResult {
	e.mutex.Lock()
	defer e.mutex.Unlock()

	// 创建战斗结果
	result := NewBattleResult(battleType, attacker, defender)
	result.Status = BattleStatusFighting
	result.StartTime = time.Now().UnixMilli()

	// 1. 计算双方战力
	attackerPower := CalcSidePower(attacker, attackerBuffs)
	defenderPower := CalcSidePower(defender, defenderBuffs)

	attacker.Power = attackerPower
	defender.Power = defenderPower

	// 2. 计算双方攻击力和防御力
	attackerAttack := CalcSideAttack(attacker, attackerBuffs)
	attackerDefense := CalcSideDefense(attacker, attackerBuffs)
	defenderAttack := CalcSideAttack(defender, defenderBuffs)
	defenderDefense := CalcSideDefense(defender, defenderBuffs)

	// 3. 计算伤害
	attackerDamage := e.CalcDamage(attackerAttack, defenderDefense)
	defenderDamage := e.CalcDamage(defenderAttack, attackerDefense)

	// 4. 判定胜负 (战力对比 + 随机因素)
	result.Winner = e.determineWinner(attackerPower, defenderPower)

	// 胜利方伤害加成
	if result.Winner == "attacker" {
		attackerDamage = int64(float64(attackerDamage) * 1.2)
	} else if result.Winner == "defender" {
		defenderDamage = int64(float64(defenderDamage) * 1.2)
	}

	// 5. 计算伤亡
	e.CalcCasualties(attacker, defenderDamage, battleType, true)
	e.CalcCasualties(defender, attackerDamage, battleType, false)

	// 6. 应用伤亡
	attacker.ApplyCasualties()
	defender.ApplyCasualties()

	// 7. 结算战斗
	result.Duration = 1 // 简化为单回合
	result.EndTime = time.Now().UnixMilli()
	result.Status = BattleStatusFinished

	// 8. 计算奖励 (攻击方胜利时)
	if result.IsAttackerWin() {
		result.Rewards = CalcRewards(battleType, defender, result.Winner)
	}

	return result
}

// CalcDamage 计算伤害
// 公式: 攻击力 × 1000 / max(1, 防御力) × random(0.9~1.1)
func (e *Engine) CalcDamage(attackPower, defensePower int64) int64 {
	// 基础伤害
	damage := attackPower * 1000 / max(1, defensePower)

	// 加入随机因素 (±10%)
	randomFactor := 0.9 + e.rand.Float64()*0.2
	damage = int64(float64(damage) * randomFactor)

	return max(1, damage)
}

// CalcCasualties 计算伤亡
func (e *Engine) CalcCasualties(side *BattleSide, damage int64, battleType BattleType, isAttacker bool) {
	// 获取伤亡配置
	cfg := GetBattleLossConfig(battleType, isAttacker)
	if cfg == nil {
		cfg = &BattleLossConfig{
			SeriousInjuryRate: 300,
			DeathRate:          100,
		}
	}

	// 计算总士兵数
	totalSoldiers := side.GetTotalSoldiers()
	if totalSoldiers == 0 {
		return
	}

	// 计算总HP
	totalHP := CalcTotalHP(side.Soldiers)
	if totalHP == 0 {
		return
	}

	// 伤亡比例 = 伤害 / 总HP
	lossRate := float64(damage) / float64(totalHP)
	if lossRate > 1.0 {
		lossRate = 1.0 // 最多100%伤亡
	}

	// 按士兵类型分配伤亡
	for soldierType, count := range side.Soldiers {
		// 该类型士兵的总伤亡
		totalLoss := int(float64(count) * lossRate)

		// 分配: 死亡、重伤、轻伤
		death := int(float64(totalLoss) * float64(cfg.DeathRate) / 1000)
		serious := int(float64(totalLoss) * float64(cfg.SeriousInjuryRate) / 1000)
		minor := totalLoss - death - serious

		// 确保不超过总数
		death = min(death, count)
		serious = min(serious, count-death)
		minor = min(minor, count-death-serious)

		side.Death[soldierType] = death
		side.SeriousWound[soldierType] = serious
		side.MinorWound[soldierType] = minor
	}
}

// determineWinner 判定胜负
func (e *Engine) determineWinner(attackerPower, defenderPower int64) string {
	// 加入随机因素 (±10%)
	attackerFactor := 0.9 + e.rand.Float64()*0.2
	defenderFactor := 0.9 + e.rand.Float64()*0.2

	adjustedAttackerPower := float64(attackerPower) * attackerFactor
	adjustedDefenderPower := float64(defenderPower) * defenderFactor

	if adjustedAttackerPower > adjustedDefenderPower {
		return "attacker"
	} else if adjustedAttackerPower < adjustedDefenderPower {
		return "defender"
	}

	// 平局时随机判定
	if e.rand.Float64() < 0.5 {
		return "attacker"
	}
	return "defender"
}

// max 返回较大值
func max(a, b int64) int64 {
	if a > b {
		return a
	}
	return b
}

// min 返回较小值
func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}
