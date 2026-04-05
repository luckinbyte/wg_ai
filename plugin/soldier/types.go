package main

// SoldierType 兵种类型
type SoldierType int

const (
	SoldierTypeInfantry SoldierType = 1 // 步兵
	SoldierTypeCavalry  SoldierType = 2 // 骑兵
	SoldierTypeArcher   SoldierType = 3 // 弓兵
	SoldierTypeSiege    SoldierType = 4 // 攻城
)

// String 返回兵种名称
func (t SoldierType) String() string {
	names := []string{"", "步兵", "骑兵", "弓兵", "攻城"}
	if int(t) < len(names) {
		return names[t]
	}
	return "未知"
}

// SoldierConfig 士兵配置
type SoldierConfig struct {
	ID      int    `yaml:"id" json:"id"`
	Type    int    `yaml:"type" json:"type"`
	Level   int    `yaml:"level" json:"level"`
	Name    string `yaml:"name" json:"name"`
	Attack  int64  `yaml:"attack" json:"attack"`
	Defense int64  `yaml:"defense" json:"defense"`
	HP      int64  `yaml:"hp" json:"hp"`
	Speed   int64  `yaml:"speed" json:"speed"`
	Load    int64  `yaml:"load" json:"load"`
	Power   int64  `yaml:"power" json:"power"`

	// 训练消耗
	CostFood  int64 `yaml:"cost_food" json:"cost_food"`
	CostWood  int64 `yaml:"cost_wood" json:"cost_wood"`
	CostStone int64 `yaml:"cost_stone" json:"cost_stone"`
	CostGold  int64 `yaml:"cost_gold" json:"cost_gold"`
	TrainTime int   `yaml:"train_time" json:"train_time"`

	// 治疗消耗
	HealFood int64 `yaml:"heal_food" json:"heal_food"`
	HealWood int64 `yaml:"heal_wood" json:"heal_wood"`
	HealTime int   `yaml:"heal_time" json:"heal_time"`
}

// SoldierData 玩家士兵数据
type SoldierData struct {
	ID      int `json:"id"`      // 士兵ID
	Type    int `json:"type"`    // 兵种类型
	Level   int `json:"level"`   // 兵种等级
	Count   int `json:"count"`   // 健康数量
	Wounded int `json:"wounded"` // 受伤数量
}
