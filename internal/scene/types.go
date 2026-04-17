package scene

// Vector2 2D坐标
type Vector2 struct {
	X float64 `json:"x"`
	Y float64 `json:"y"`
}

// EntityType 实体类型
type EntityType int

const (
	EntityTypePlayer EntityType = iota
	EntityTypeNPC
	EntityTypeMonster
	EntityTypeResource
	EntityTypeBuilding
	EntityTypeArmy
)

// String 返回实体类型名称
func (t EntityType) String() string {
	names := []string{"player", "npc", "monster", "resource", "building", "army"}
	if int(t) < len(names) {
		return names[t]
	}
	return "unknown"
}

// EventType AOI事件类型
type EventType int

const (
	EventEnter EventType = iota
	EventLeave
)

// String 返回事件类型名称
func (e EventType) String() string {
	if e == EventEnter {
		return "enter"
	}
	return "leave"
}

// Entity 场景实体
type Entity struct {
	ID       int64      `json:"id"`
	Type     EntityType `json:"type"`
	Position Vector2    `json:"position"`
	SceneID  int64      `json:"scene_id"`

	// 扩展数据 (如玩家名称、NPC配置ID等)
	Data map[string]any `json:"data,omitempty"`
}

// SetData 设置扩展数据
func (e *Entity) SetData(key string, value any) {
	if e.Data == nil {
		e.Data = make(map[string]any)
	}
	e.Data[key] = value
}

// GetData 获取扩展数据
func (e *Entity) GetData(key string) (any, bool) {
	if e.Data == nil {
		return nil, false
	}
	v, ok := e.Data[key]
	return v, ok
}

// Grid 格子
type Grid struct {
	X       int              // 格子X坐标
	Y       int              // 格子Y坐标
	Entities map[int64]*Entity // 格子内的实体
}

// NewGrid 创建新格子
func NewGrid(x, y int) *Grid {
	return &Grid{
		X:        x,
		Y:        y,
		Entities: make(map[int64]*Entity),
	}
}

// AddEntity 添加实体到格子
func (g *Grid) AddEntity(entity *Entity) {
	g.Entities[entity.ID] = entity
}

// RemoveEntity 从格子移除实体
func (g *Grid) RemoveEntity(entityID int64) {
	delete(g.Entities, entityID)
}

// GetEntities 获取格子内所有实体
func (g *Grid) GetEntities() []*Entity {
	entities := make([]*Entity, 0, len(g.Entities))
	for _, e := range g.Entities {
		entities = append(entities, e)
	}
	return entities
}

// EntityCount 获取格子内实体数量
func (g *Grid) EntityCount() int {
	return len(g.Entities)
}

// AOIEvent AOI事件
type AOIEvent struct {
	Type    EventType `json:"type"`    // 事件类型
	Entity  *Entity   `json:"entity"`  // 触发事件的实体
	Watcher int64     `json:"watcher"` // 观察者ID (看到此事件的玩家)
}

// SceneConfig 场景配置
type SceneConfig struct {
	ID       int64 `json:"id"`
	Width    int   `json:"width"`    // 场景宽度 (像素/单位)
	Height   int   `json:"height"`   // 场景高度 (像素/单位)
	GridSize int   `json:"grid_size"` // 格子大小
}

// ============ 地图对象相关类型 ============

// ResourceType 资源类型
type ResourceType int

const (
	ResourceFood ResourceType = iota
	ResourceWood
	ResourceStone
	ResourceGold
)

// String 返回资源类型名称
func (r ResourceType) String() string {
	names := []string{"food", "wood", "stone", "gold"}
	if int(r) < len(names) {
		return names[r]
	}
	return "unknown"
}

// MapObjectData 地图对象扩展数据
type MapObjectData struct {
	ConfigID     int          `json:"config_id"`     // 配置ID
	Level        int          `json:"level"`         // 等级
	ResourceType ResourceType `json:"resource_type"` // 资源类型 (资源点用)
	Amount       int64        `json:"amount"`        // 剩余数量
	MaxAmount    int64        `json:"max_amount"`    // 最大数量
	RefreshTime  int64        `json:"refresh_time"`  // 刷新时间戳
	OwnerID      int64        `json:"owner_id"`      // 占领者ID (0=无人占领)
}

// ArmyStatus 军队状态
type ArmyStatus int

const (
	ArmyStatusIdle ArmyStatus = iota // 空闲
	ArmyStatusMarching               // 行军中
	ArmyStatusCollecting             // 采集中
	ArmyStatusBattle                 // 战斗中
	ArmyStatusStationing             // 驻扎中
)

// String 返回军队状态名称
func (s ArmyStatus) String() string {
	names := []string{"idle", "marching", "collecting", "battle", "stationing"}
	if int(s) < len(names) {
		return names[s]
	}
	return "unknown"
}

// MarchType 行军类型
type MarchType int

const (
	MarchTypeCollect MarchType = iota // 采集
	MarchTypeAttack                   // 攻击
	MarchTypeReinforce                // 支援
	MarchTypeReturn                   // 返回
)

// String 返回行军类型名称
func (t MarchType) String() string {
	names := []string{"collect", "attack", "reinforce", "return"}
	if int(t) < len(names) {
		return names[t]
	}
	return "unknown"
}

// MarchData 行军数据
type MarchData struct {
	Type            MarchType `json:"type"`             // 行军类型
	TargetID        int64     `json:"target_id"`        // 目标ID (资源点/怪物/玩家)
	TargetPos       Vector2   `json:"target_pos"`       // 目标位置
	Path            []Vector2 `json:"path"`             // 行军路径
	StartTime       int64     `json:"start_time"`       // 开始时间 (毫秒)
	ArrivalTime     int64     `json:"arrival_time"`     // 到达时间 (毫秒)
	Speed           float64   `json:"speed"`            // 行军速度 (单位/秒)
	CollectAmount   int64     `json:"collect_amount"`   // 已采集数量
	CollectEndTime  int64     `json:"collect_end_time"` // 采集结束时间 (毫秒)
}

// ArmyData 军队扩展数据 (存储在Entity.Data中)
type ArmyData struct {
	OwnerID  int64   `json:"owner_id"`  // 所属玩家ID
	HeroID   int64   `json:"hero_id"`   // 英雄ID
	Soldiers int     `json:"soldiers"`  // 士兵数量
	Status   ArmyStatus `json:"status"` // 军队状态
	March    *MarchData `json:"march"`  // 行军数据 (nil表示未行军)
}
