package march

import (
	"time"

	"github.com/luckinbyte/wg_ai/internal/scene"
)

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
	Type           MarchType       `json:"type"`             // 行军类型
	TargetID       int64           `json:"target_id"`        // 目标ID (资源点/怪物/玩家)
	TargetPos      scene.Vector2   `json:"target_pos"`       // 目标位置
	Path           []scene.Vector2 `json:"path"`             // 行军路径
	StartTime      int64           `json:"start_time"`       // 开始时间 (毫秒)
	ArrivalTime    int64           `json:"arrival_time"`     // 到达时间 (毫秒)
	Speed          float64         `json:"speed"`            // 行军速度 (单位/秒)
	CollectAmount  int64           `json:"collect_amount"`   // 已采集数量
	CollectEndTime int64           `json:"collect_end_time"` // 采集结束时间 (毫秒)
}

// Army 军队
type Army struct {
	ID       int64         `json:"id"`
	OwnerID  int64         `json:"owner_id"`  // 所属玩家ID
	HeroID   int64         `json:"hero_id"`   // 英雄ID
	Soldiers map[int]int   `json:"soldiers"`  // 士兵 {soldierID: count}
	Status   ArmyStatus    `json:"status"`    // 军队状态
	Position scene.Vector2 `json:"position"`  // 当前位置
	SceneID  int64         `json:"scene_id"`  // 场景ID

	// 行军数据 (nil表示未行军)
	March *MarchData `json:"march,omitempty"`

	// 采集资源 (采集完成后携带)
	LoadFood  int64 `json:"load_food"`  // 携带粮食
	LoadWood  int64 `json:"load_wood"`  // 携带木材
	LoadStone int64 `json:"load_stone"` // 携带石材
	LoadGold  int64 `json:"load_gold"`  // 携带金币

	// 内部数据
	entityID int64 // 场景实体ID (行军时使用)
}

// GetTotalSoldiers 获取士兵总数
func (a *Army) GetTotalSoldiers() int {
	total := 0
	for _, count := range a.Soldiers {
		total += count
	}
	return total
}

// IsMarching 是否正在行军
func (a *Army) IsMarching() bool {
	return a.Status == ArmyStatusMarching || a.Status == ArmyStatusCollecting
}

// NewArmy 创建新军队
func NewArmy(id, ownerID, heroID int64, soldiers map[int]int) *Army {
	if soldiers == nil {
		soldiers = make(map[int]int)
	}
	return &Army{
		ID:       id,
		OwnerID:  ownerID,
		HeroID:   heroID,
		Soldiers: soldiers,
		Status:   ArmyStatusIdle,
	}
}

// IsIdle 是否空闲
func (a *Army) IsIdle() bool {
	return a.Status == ArmyStatusIdle
}

// IsCollecting 是否采集中
func (a *Army) IsCollecting() bool {
	return a.Status == ArmyStatusCollecting
}

// CanMarch 是否可以开始行军
func (a *Army) CanMarch() bool {
	return a.Status == ArmyStatusIdle
}

// StartMarch 开始行军
func (a *Army) StartMarch(marchType MarchType, targetID int64, targetPos scene.Vector2, path []scene.Vector2, speed float64) {
	now := time.Now().UnixMilli()

	// 计算到达时间
	totalDist := 0.0
	for i := 0; i < len(path)-1; i++ {
		dx := path[i+1].X - path[i].X
		dy := path[i+1].Y - path[i].Y
		totalDist += sqrt(dx*dx + dy*dy)
	}
	travelTime := int64(totalDist / speed * 1000) // 毫秒

	a.Status = ArmyStatusMarching
	a.March = &MarchData{
		Type:        marchType,
		TargetID:    targetID,
		TargetPos:   targetPos,
		Path:        path,
		StartTime:   now,
		ArrivalTime: now + travelTime,
		Speed:       speed,
	}
}

// UpdatePosition 更新行军位置
func (a *Army) UpdatePosition(pos scene.Vector2) {
	a.Position = pos
}

// FinishMarch 结束行军
func (a *Army) FinishMarch() {
	if a.March != nil {
		a.Position = a.March.TargetPos
	}
	a.Status = ArmyStatusIdle
	a.March = nil
}

// StartCollect 开始采集
func (a *Army) StartCollect(duration time.Duration) {
	a.Status = ArmyStatusCollecting
	if a.March != nil {
		a.March.CollectEndTime = time.Now().UnixMilli() + duration.Milliseconds()
	}
}

// FinishCollect 完成采集
func (a *Army) FinishCollect() {
	a.Status = ArmyStatusIdle
	if a.March != nil {
		a.March.CollectAmount = 0
		a.March.CollectEndTime = 0
	}
}

// GetProgress 获取行军进度 (0.0 - 1.0)
func (a *Army) GetProgress() float64 {
	if a.March == nil {
		return 1.0
	}

	now := time.Now().UnixMilli()
	if now >= a.March.ArrivalTime {
		return 1.0
	}

	elapsed := now - a.March.StartTime
	total := a.March.ArrivalTime - a.March.StartTime
	if total <= 0 {
		return 1.0
	}

	return float64(elapsed) / float64(total)
}

// CalcPower 计算战力
func (a *Army) CalcPower() int64 {
	// 简单战力计算: 士兵总数 * 10 + 英雄加成
	basePower := int64(a.GetTotalSoldiers() * 10)
	// TODO: 加入英雄加成和各兵种战力
	return basePower
}

// CalcLoadCapacity 计算负重上限
func (a *Army) CalcLoadCapacity() int64 {
	// 简单负重计算: 士兵总数 * 100
	return int64(a.GetTotalSoldiers() * 100)
}

// GetCurrentLoad 获取当前负重
func (a *Army) GetCurrentLoad() int64 {
	return a.LoadFood + a.LoadWood + a.LoadStone + a.LoadGold
}

// CanLoadMore 是否还能装载更多
func (a *Army) CanLoadMore(amount int64) bool {
	return a.GetCurrentLoad()+amount <= a.CalcLoadCapacity()
}

// AddLoad 添加负重
func (a *Army) AddLoad(resourceType scene.ResourceType, amount int64) {
	switch resourceType {
	case scene.ResourceFood:
		a.LoadFood += amount
	case scene.ResourceWood:
		a.LoadWood += amount
	case scene.ResourceStone:
		a.LoadStone += amount
	case scene.ResourceGold:
		a.LoadGold += amount
	}
}

// ClearLoad 清空负重
func (a *Army) ClearLoad() {
	a.LoadFood = 0
	a.LoadWood = 0
	a.LoadStone = 0
	a.LoadGold = 0
}

// sqrt 平方根
func sqrt(x float64) float64 {
	if x <= 0 {
		return 0
	}
	z := x
	for i := 0; i < 10; i++ {
		z = (z + x/z) / 2
	}
	return z
}
