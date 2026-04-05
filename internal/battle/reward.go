package battle

// RewardConfig 奖励配置
type RewardConfig struct {
	BattleType      BattleType // 战斗类型
	BaseHeroExp     int64      // 基础英雄经验
	BaseFood        int64      // 基础粮食
	BaseWood        int64      // 基础木材
	BaseStone       int64      // 基础石头
	BaseGold        int64      // 基础金币
	PowerMultiplier float64    // 战力系数
}

// 默认奖励配置
var defaultRewardConfigs = map[BattleType]*RewardConfig{
	BattleTypeMonster: {
		BattleType:      BattleTypeMonster,
		BaseHeroExp:     100,
		BaseFood:        500,
		BaseWood:        300,
		BaseStone:       100,
		BaseGold:        50,
		PowerMultiplier: 0.1,
	},
	BattleTypeMonsterCity: {
		BattleType:      BattleTypeMonsterCity,
		BaseHeroExp:     500,
		BaseFood:        2000,
		BaseWood:        1500,
		BaseStone:       800,
		BaseGold:        300,
		PowerMultiplier: 0.2,
	},
	BattleTypeResource: {
		BattleType:      BattleTypeResource,
		BaseHeroExp:     200,
		BaseFood:        1000,
		BaseWood:        800,
		BaseStone:       400,
		BaseGold:        100,
		PowerMultiplier: 0.15,
	},
	BattleTypeCity: {
		BattleType:      BattleTypeCity,
		BaseHeroExp:     1000,
		BaseFood:        5000,
		BaseWood:        4000,
		BaseStone:       2000,
		BaseGold:        1000,
		PowerMultiplier: 0.25,
	},
}

// CalcRewards 计算战斗奖励
func CalcRewards(battleType BattleType, loser *BattleSide, winner string) *BattleRewards {
	cfg, ok := defaultRewardConfigs[battleType]
	if !ok {
		cfg = defaultRewardConfigs[BattleTypeMonster]
	}

	// 基础奖励
	rewards := &BattleRewards{
		HeroExp: cfg.BaseHeroExp,
		Food:    cfg.BaseFood,
		Wood:    cfg.BaseWood,
		Stone:   cfg.BaseStone,
		Gold:    cfg.BaseGold,
		Items:   []ItemReward{},
	}

	// 根据败方战力加成
	if loser != nil && loser.Power > 0 {
		powerBonus := float64(loser.Power) * cfg.PowerMultiplier
		rewards.HeroExp += int64(powerBonus)
		rewards.Food += int64(powerBonus * 2)
		rewards.Wood += int64(powerBonus * 1.5)
		rewards.Stone += int64(powerBonus * 0.5)
		rewards.Gold += int64(powerBonus * 0.3)
	}

	// TODO: 添加物品奖励
	// rewards.Items = append(rewards.Items, ItemReward{ItemID: 1001, Count: 1})

	return rewards
}

// GetRewardConfig 获取奖励配置
func GetRewardConfig(battleType BattleType) *RewardConfig {
	if cfg, ok := defaultRewardConfigs[battleType]; ok {
		return cfg
	}
	return defaultRewardConfigs[BattleTypeMonster]
}

// SetRewardConfig 设置奖励配置 (用于热更新)
func SetRewardConfig(battleType BattleType, cfg *RewardConfig) {
	defaultRewardConfigs[battleType] = cfg
}
