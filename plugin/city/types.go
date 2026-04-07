package city

// BuildingType 建筑类型
type BuildingType int

const (
	BuildingTypeCastle    BuildingType = 1
	BuildingTypeBarracks  BuildingType = 2
	BuildingTypeFarm      BuildingType = 3
	BuildingTypeLumber    BuildingType = 4
	BuildingTypeQuarry    BuildingType = 5
	BuildingTypeHospital  BuildingType = 6
	BuildingTypeWarehouse BuildingType = 7
)

// BuildingConfig 建筑配置
type BuildingConfig struct {
	ID        int    `yaml:"id" json:"id"`
	Type      int    `yaml:"type" json:"type"`
	Level     int    `yaml:"level" json:"level"`
	Name      string `yaml:"name" json:"name"`
	HP        int64  `yaml:"hp" json:"hp"`
	BuildTime int    `yaml:"build_time" json:"build_time"`
	CostFood  int64  `yaml:"cost_food" json:"cost_food"`
	CostWood  int64  `yaml:"cost_wood" json:"cost_wood"`
	CostStone int64  `yaml:"cost_stone" json:"cost_stone"`
	CostGold  int64  `yaml:"cost_gold" json:"cost_gold"`
	ProdFood  int64  `yaml:"prod_food" json:"prod_food"`
	ProdWood  int64  `yaml:"prod_wood" json:"prod_wood"`
	ProdStone int64  `yaml:"prod_stone" json:"prod_stone"`
	ProdGold  int64  `yaml:"prod_gold" json:"prod_gold"`
	TrainSlot int    `yaml:"train_slot" json:"train_slot"`
	HealSlot  int    `yaml:"heal_slot" json:"heal_slot"`
	MaxArmy   int    `yaml:"max_army" json:"max_army"`
}

// Position 位置
type Position struct {
	X float64 `json:"x"`
	Y float64 `json:"y"`
}

// BuildingData 单个建筑数据
type BuildingData struct {
	Type     int   `json:"type"`
	Level    int   `json:"level"`
	HP       int64 `json:"hp"`
	EntityID int64 `json:"entity_id"`
}

// BuildQueueItem 建造队列项
type BuildQueueItem struct {
	ID           int64 `json:"id"`
	BuildingType int   `json:"building_type"`
	TargetLevel  int   `json:"target_level"`
	StartTime    int64 `json:"start_time"`
	FinishTime   int64 `json:"finish_time"`
}

// CityData 玩家城池数据
type CityData struct {
	CityID     int64                  `json:"city_id"`
	Position   Position               `json:"position"`
	Buildings  map[int]*BuildingData  `json:"buildings"`
	BuildQueue []BuildQueueItem       `json:"build_queue"`
}
