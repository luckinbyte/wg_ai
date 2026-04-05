package battle

import (
	"testing"
)

// TestBattleTypeString 测试战斗类型字符串转换
func TestBattleTypeString(t *testing.T) {
	tests := []struct {
		battleType BattleType
		expected   string
	}{
		{BattleTypeMonster, "monster"},
		{BattleTypeMonsterCity, "monster_city"},
		{BattleTypeResource, "resource"},
		{BattleTypeCity, "city"},
		{BattleType(999), "unknown(999)"},
	}

	for _, test := range tests {
		if test.battleType.String() != test.expected {
			t.Errorf("BattleType(%d).String() = %s, expected %s",
				test.battleType, test.battleType.String(), test.expected)
		}
	}
}

// TestBattleStatusString 测试战斗状态字符串转换
func TestBattleStatusString(t *testing.T) {
	tests := []struct {
		status   BattleStatus
		expected string
	}{
		{BattleStatusPending, "pending"},
		{BattleStatusFighting, "fighting"},
		{BattleStatusFinished, "finished"},
		{BattleStatus(999), "unknown(999)"},
	}

	for _, test := range tests {
		if test.status.String() != test.expected {
			t.Errorf("BattleStatus(%d).String() = %s, expected %s",
				test.status, test.status.String(), test.expected)
		}
	}
}

// TestBattleSide 测试战斗方
func TestBattleSide(t *testing.T) {
	side := NewBattleSide(100, SideTypePlayer)

	if side.ID != 100 {
		t.Errorf("Expected ID=100, got %d", side.ID)
	}
	if side.Type != SideTypePlayer {
		t.Errorf("Expected Type=player, got %s", side.Type)
	}

	// 设置士兵
	side.SetSoldiers(map[int]int{
		1001: 500, // 步兵
		1002: 300, // 骑兵
		1003: 200, // 弓兵
	})

	total := side.GetTotalSoldiers()
	if total != 1000 {
		t.Errorf("Expected total soldiers=1000, got %d", total)
	}
}

// TestCalcArmyPower 测试战力计算
func TestCalcArmyPower(t *testing.T) {
	soldiers := map[int]int{
		1001: 100, // 100步兵
	}

	// 无加成
	power := CalcArmyPower(soldiers, nil)
	expected := int64(100 * 10) // 100 * 10战力
	if power != expected {
		t.Errorf("Expected power=%d, got %d", expected, power)
	}

	// 有加成
	buffs := map[string]float64{
		"attack":  0.1,  // 10%攻击加成
		"defense": 0.05, // 5%防御加成
	}
	powerWithBuffs := CalcArmyPower(soldiers, buffs)
	if powerWithBuffs <= power {
		t.Errorf("Power with buffs should be greater than without")
	}
}

// TestCalcAttackPower 测试攻击力计算
func TestCalcAttackPower(t *testing.T) {
	soldiers := map[int]int{
		1001: 100, // 100步兵
	}

	attack := CalcAttackPower(soldiers, nil)
	expected := int64(100 * 100) // 100士兵 * 100攻击
	if attack != expected {
		t.Errorf("Expected attack=%d, got %d", expected, attack)
	}
}

// TestCalcDefensePower 测试防御力计算
func TestCalcDefensePower(t *testing.T) {
	soldiers := map[int]int{
		1001: 100, // 100步兵
	}

	defense := CalcDefensePower(soldiers, nil)
	expected := int64(100 * 80) // 100士兵 * 80防御
	if defense != expected {
		t.Errorf("Expected defense=%d, got %d", expected, defense)
	}
}

// TestCalcDamage 测试伤害计算
func TestCalcDamage(t *testing.T) {
	engine := NewEngine()

	attack := int64(1000)
	defense := int64(500)

	// 多次测试验证随机因素
	for i := 0; i < 10; i++ {
		damage := engine.CalcDamage(attack, defense)
		// 伤害应该在 1800 ~ 2200 之间 (1000*1000/500 * 0.9~1.1)
		if damage < 1600 || damage > 2400 {
			t.Errorf("Damage %d out of expected range [1600, 2400]", damage)
		}
	}
}

// TestCalcCasualties 测试伤亡计算
func TestCalcCasualties(t *testing.T) {
	engine := NewEngine()

	side := NewBattleSide(1, SideTypePlayer)
	side.SetSoldiers(map[int]int{
		1001: 1000, // 1000步兵
	})

	// 计算总HP: 1000 * 1000 = 1,000,000
	damage := int64(100000) // 10%伤害

	engine.CalcCasualties(side, damage, BattleTypeMonster, true)

	totalDeath := side.GetTotalDeaths()
	totalSerious := side.GetTotalSeriousWound()
	totalMinor := side.GetTotalMinorWound()

	if totalDeath == 0 {
		t.Error("Expected some deaths")
	}
	if totalSerious == 0 {
		t.Error("Expected some serious wounds")
	}
	if totalMinor == 0 {
		t.Error("Expected some minor wounds")
	}

	t.Logf("Deaths: %d, Serious: %d, Minor: %d", totalDeath, totalSerious, totalMinor)
}

// TestFullBattle 测试完整战斗流程
func TestFullBattle(t *testing.T) {
	engine := NewEngine()

	// 创建攻击方
	attacker := NewBattleSide(100, SideTypePlayer)
	attacker.HeroID = 1
	attacker.SetSoldiers(map[int]int{
		1001: 500, // 500步兵
		1002: 300, // 300骑兵
		1003: 200, // 200弓兵
	})

	// 创建防守方 (怪物)
	defender := NewBattleSide(1001, SideTypeMonster)
	defender.SetSoldiers(map[int]int{
		1001: 800, // 800步兵
	})

	// 开始战斗
	result := engine.StartBattle(BattleTypeMonster, attacker, defender, nil, nil)

	if result.Status != BattleStatusFinished {
		t.Errorf("Expected status=finished, got %s", result.Status)
	}

	if result.Winner != "attacker" && result.Winner != "defender" {
		t.Errorf("Invalid winner: %s", result.Winner)
	}

	if result.Duration != 1 {
		t.Errorf("Expected duration=1, got %d", result.Duration)
	}

	t.Logf("Battle result: Winner=%s, Duration=%d", result.Winner, result.Duration)
	t.Logf("Attacker power: %d, Defender power: %d", result.Attacker.Power, result.Defender.Power)
	t.Logf("Attacker deaths: %d, Defender deaths: %d",
		result.Attacker.GetTotalDeaths(), result.Defender.GetTotalDeaths())

	// 攻击方胜利时应该有奖励
	if result.IsAttackerWin() && result.Rewards == nil {
		t.Error("Expected rewards when attacker wins")
	}
}

// TestBattleReport 测试战报生成
func TestBattleReport(t *testing.T) {
	engine := NewEngine()

	attacker := NewBattleSide(100, SideTypePlayer)
	attacker.SetSoldiers(map[int]int{1001: 500})

	defender := NewBattleSide(1001, SideTypeMonster)
	defender.SetSoldiers(map[int]int{1001: 300})

	result := engine.StartBattle(BattleTypeMonster, attacker, defender, nil, nil)

	report := GenerateReport(result)

	if report.BattleType != "monster" {
		t.Errorf("Expected battle_type=monster, got %s", report.BattleType)
	}

	if report.Winner == "" {
		t.Error("Expected winner to be set")
	}

	// 测试JSON序列化
	jsonBytes, err := report.ToJSON()
	if err != nil {
		t.Errorf("Failed to serialize report: %v", err)
	}

	if len(jsonBytes) == 0 {
		t.Error("Expected non-empty JSON")
	}

	t.Logf("Report JSON length: %d", len(jsonBytes))
	t.Logf("Summary: %s", report.Summary())
}

// TestBattleRewards 测试奖励计算
func TestBattleRewards(t *testing.T) {
	loser := NewBattleSide(1001, SideTypeMonster)
	loser.Power = 10000

	rewards := CalcRewards(BattleTypeMonster, loser, "attacker")

	if rewards == nil {
		t.Fatal("Expected rewards to be non-nil")
	}

	if rewards.HeroExp == 0 {
		t.Error("Expected hero exp to be non-zero")
	}

	if rewards.Food == 0 {
		t.Error("Expected food to be non-zero")
	}

	t.Logf("Rewards: Exp=%d, Food=%d, Wood=%d, Stone=%d, Gold=%d",
		rewards.HeroExp, rewards.Food, rewards.Wood, rewards.Stone, rewards.Gold)
}

// TestConfigLoading 测试配置加载
func TestConfigLoading(t *testing.T) {
	// 测试获取战斗伤亡配置
	cfg := GetBattleLossConfig(BattleTypeMonster, true)
	if cfg == nil {
		t.Fatal("Expected battle loss config to be non-nil")
	}

	if cfg.SeriousInjuryRate == 0 {
		t.Error("Expected serious injury rate to be non-zero")
	}

	if cfg.DeathRate == 0 {
		t.Error("Expected death rate to be non-zero")
	}

	// 测试获取士兵配置
	soldierCfg := GetSoldierConfig(1001)
	if soldierCfg == nil {
		t.Fatal("Expected soldier config to be non-nil")
	}

	if soldierCfg.Name != "步兵" {
		t.Errorf("Expected soldier name=步兵, got %s", soldierCfg.Name)
	}
}

// TestApplyCasualties 测试应用伤亡
func TestApplyCasualties(t *testing.T) {
	side := NewBattleSide(1, SideTypePlayer)
	side.SetSoldiers(map[int]int{
		1001: 1000,
	})

	// 设置伤亡
	side.Death[1001] = 100        // 死亡100
	side.SeriousWound[1001] = 200 // 重伤200
	side.MinorWound[1001] = 300   // 轻伤300

	// 应用伤亡
	side.ApplyCasualties()

	// 剩余士兵应该是 1000 - 100 = 900
	if side.Soldiers[1001] != 900 {
		t.Errorf("Expected soldiers=900, got %d", side.Soldiers[1001])
	}
}
