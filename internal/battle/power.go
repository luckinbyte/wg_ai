package battle

// CalcArmyPower 计算军队战力
// 公式: 士兵数 × 单兵战力 × 加成
func CalcArmyPower(soldiers map[int]int, buffs map[string]float64) int64 {
	var totalPower int64

	for soldierType, count := range soldiers {
		cfg := GetSoldierConfig(soldierType)
		if cfg == nil {
			continue
		}

		basePower := cfg.Power * int64(count)

		// 应用加成
		attackBuff := getBuff(buffs, "attack")
		defenseBuff := getBuff(buffs, "defense")

		// 最终战力 = 基础战力 × (1 + 攻击加成 + 防御加成)
		finalPower := float64(basePower) * (1 + attackBuff + defenseBuff)
		totalPower += int64(finalPower)
	}

	return totalPower
}

// CalcAttackPower 计算攻击力
// 公式: 士兵数 × 单兵攻击 × 加成
func CalcAttackPower(soldiers map[int]int, buffs map[string]float64) int64 {
	var totalAttack int64

	for soldierType, count := range soldiers {
		cfg := GetSoldierConfig(soldierType)
		if cfg == nil {
			continue
		}

		baseAttack := cfg.Attack * int64(count)
		attackBuff := getBuff(buffs, "attack")

		finalAttack := float64(baseAttack) * (1 + attackBuff)
		totalAttack += int64(finalAttack)
	}

	return totalAttack
}

// CalcDefensePower 计算防御力
// 公式: 士兵数 × 单兵防御 × 加成
func CalcDefensePower(soldiers map[int]int, buffs map[string]float64) int64 {
	var totalDefense int64

	for soldierType, count := range soldiers {
		cfg := GetSoldierConfig(soldierType)
		if cfg == nil {
			continue
		}

		baseDefense := cfg.Defense * int64(count)
		defenseBuff := getBuff(buffs, "defense")

		finalDefense := float64(baseDefense) * (1 + defenseBuff)
		totalDefense += int64(finalDefense)
	}

	return totalDefense
}

// CalcTotalHP 计算总生命值
func CalcTotalHP(soldiers map[int]int) int64 {
	var totalHP int64

	for soldierType, count := range soldiers {
		cfg := GetSoldierConfig(soldierType)
		if cfg == nil {
			continue
		}

		totalHP += cfg.HP * int64(count)
	}

	return totalHP
}

// getBuff 获取加成值
func getBuff(buffs map[string]float64, key string) float64 {
	if buffs == nil {
		return 0
	}
	return buffs[key]
}

// CalcSidePower 计算战斗方战力
func CalcSidePower(side *BattleSide, buffs map[string]float64) int64 {
	return CalcArmyPower(side.Soldiers, buffs)
}

// CalcSideAttack 计算战斗方攻击力
func CalcSideAttack(side *BattleSide, buffs map[string]float64) int64 {
	return CalcAttackPower(side.Soldiers, buffs)
}

// CalcSideDefense 计算战斗方防御力
func CalcSideDefense(side *BattleSide, buffs map[string]float64) int64 {
	return CalcDefensePower(side.Soldiers, buffs)
}
