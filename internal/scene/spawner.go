package scene

import (
	"math/rand"
	"sync"
	"time"
)

// SpawnerConfig 刷新配置
type SpawnerConfig struct {
	ResourceInterval    time.Duration // 资源刷新间隔
	MonsterInterval     time.Duration // 怪物刷新间隔
	MaxResourcesPerGrid int           // 每格最大资源数
	MaxMonstersPerGrid  int           // 每格最大怪物数
	ResourceAmountBase  int64         // 资源基础数量
	ResourceAmountPerLevel int64      // 每级增加数量
}

// DefaultSpawnerConfig 默认刷新配置
var DefaultSpawnerConfig = SpawnerConfig{
	ResourceInterval:      5 * time.Minute,
	MonsterInterval:       10 * time.Minute,
	MaxResourcesPerGrid:   3,
	MaxMonstersPerGrid:    2,
	ResourceAmountBase:    1000,
	ResourceAmountPerLevel: 500,
}

// Spawner 对象刷新器
type Spawner struct {
	scene       *Scene
	config      SpawnerConfig
	objMgr      *MapObjectManager
	rand        *rand.Rand

	resourceTimer *time.Timer
	monsterTimer  *time.Timer
	stopCh        chan struct{}
	mutex         sync.Mutex
	running       bool
}

// NewSpawner 创建刷新器
func NewSpawner(scene *Scene, objMgr *MapObjectManager, config SpawnerConfig) *Spawner {
	return &Spawner{
		scene:  scene,
		config: config,
		objMgr: objMgr,
		rand:   rand.New(rand.NewSource(time.Now().UnixNano())),
		stopCh: make(chan struct{}),
	}
}

// Start 启动刷新器
func (s *Spawner) Start() {
	s.mutex.Lock()
	defer s.mutex.Unlock()

	if s.running {
		return
	}

	s.running = true

	// 初始刷新
	s.spawnResources()
	s.spawnMonsters()

	// 启动定时器
	s.resourceTimer = time.AfterFunc(s.config.ResourceInterval, s.resourceTick)
	s.monsterTimer = time.AfterFunc(s.config.MonsterInterval, s.monsterTick)
}

// Stop 停止刷新器
func (s *Spawner) Stop() {
	s.mutex.Lock()
	defer s.mutex.Unlock()

	if !s.running {
		return
	}

	s.running = false
	close(s.stopCh)

	if s.resourceTimer != nil {
		s.resourceTimer.Stop()
	}
	if s.monsterTimer != nil {
		s.monsterTimer.Stop()
	}
}

// resourceTick 资源刷新定时器
func (s *Spawner) resourceTick() {
	select {
	case <-s.stopCh:
		return
	default:
	}

	s.spawnResources()

	s.mutex.Lock()
	if s.running {
		s.resourceTimer = time.AfterFunc(s.config.ResourceInterval, s.resourceTick)
	}
	s.mutex.Unlock()
}

// monsterTick 怪物刷新定时器
func (s *Spawner) monsterTick() {
	select {
	case <-s.stopCh:
		return
	default:
	}

	s.spawnMonsters()

	s.mutex.Lock()
	if s.running {
		s.monsterTimer = time.AfterFunc(s.config.MonsterInterval, s.monsterTick)
	}
	s.mutex.Unlock()
}

// spawnResources 刷新资源
func (s *Spawner) spawnResources() {
	// 获取当前资源数量
	currentCount := s.objMgr.ObjectCountByType(EntityTypeResource)

	// 计算需要刷新的资源数量
	gridCountX := s.scene.Width / s.scene.GridSize
	gridCountY := s.scene.Height / s.scene.GridSize
	totalGrids := gridCountX * gridCountY
	maxResources := totalGrids * s.config.MaxResourcesPerGrid

	if currentCount >= maxResources {
		return // 资源已满，不需要刷新
	}

	// 刷新数量 = 最大数量 - 当前数量 (分批刷新，每次最多刷新总数的10%)
	spawnCount := (maxResources - currentCount) / 10
	if spawnCount < 1 {
		spawnCount = 1
	}
	if spawnCount > 50 {
		spawnCount = 50
	}

	for i := 0; i < int(spawnCount); i++ {
		s.spawnOneResource()
	}
}

// spawnOneResource 刷新单个资源点
func (s *Spawner) spawnOneResource() {
	// 随机位置
	pos := s.randomPosition()

	// 随机资源类型
	resTypes := []ResourceType{ResourceFood, ResourceWood, ResourceStone, ResourceGold}
	resType := resTypes[s.rand.Intn(len(resTypes))]

	// 随机等级 (1-5)
	level := s.rand.Intn(5) + 1

	// 计算资源数量
	amount := s.config.ResourceAmountBase + int64(level)*s.config.ResourceAmountPerLevel

	// 生成资源点
	s.objMgr.SpawnResource(pos, resType, level, amount)
}

// spawnMonsters 刷新怪物
func (s *Spawner) spawnMonsters() {
	// 获取当前怪物数量
	currentCount := s.objMgr.ObjectCountByType(EntityTypeMonster)

	// 计算需要刷新的怪物数量
	gridCountX := s.scene.Width / s.scene.GridSize
	gridCountY := s.scene.Height / s.scene.GridSize
	totalGrids := gridCountX * gridCountY
	maxMonsters := totalGrids * s.config.MaxMonstersPerGrid

	if currentCount >= maxMonsters {
		return // 怪物已满，不需要刷新
	}

	// 刷新数量
	spawnCount := (maxMonsters - currentCount) / 10
	if spawnCount < 1 {
		spawnCount = 1
	}
	if spawnCount > 20 {
		spawnCount = 20
	}

	for i := 0; i < int(spawnCount); i++ {
		s.spawnOneMonster()
	}
}

// spawnOneMonster 刷新单个怪物
func (s *Spawner) spawnOneMonster() {
	// 随机位置
	pos := s.randomPosition()

	// 随机怪物配置ID (1001-1010)
	cfgID := 1001 + s.rand.Intn(10)

	// 随机等级 (1-10)
	level := s.rand.Intn(10) + 1

	// 计算战力
	power := int64(level) * 100

	// 生成怪物
	entity := s.objMgr.SpawnMonster(pos, cfgID, level, power)
	if entity != nil {
		// 设置刷新时间
		objData := entity.GetObjectData()
		if objData != nil {
			objData.RefreshTime = time.Now().Unix()
		}
	}
}

// randomPosition 生成随机位置
func (s *Spawner) randomPosition() Vector2 {
	x := float64(s.rand.Intn(s.scene.Width))
	y := float64(s.rand.Intn(s.scene.Height))
	return Vector2{X: x, Y: y}
}

// ForceSpawnResource 强制刷新资源点 (GM命令用)
func (s *Spawner) ForceSpawnResource(pos Vector2, resType ResourceType, level int, amount int64) *Entity {
	return s.objMgr.SpawnResource(pos, resType, level, amount)
}

// ForceSpawnMonster 强制刷新怪物 (GM命令用)
func (s *Spawner) ForceSpawnMonster(pos Vector2, cfgID int, level int) *Entity {
	power := int64(level) * 100
	return s.objMgr.SpawnMonster(pos, cfgID, level, power)
}

// ClearResources 清除所有资源 (GM命令用)
func (s *Spawner) ClearResources() {
	resources := s.objMgr.GetObjectsByType(EntityTypeResource)
	for _, r := range resources {
		s.objMgr.RemoveObject(r.ID)
	}
}

// ClearMonsters 清除所有怪物 (GM命令用)
func (s *Spawner) ClearMonsters() {
	monsters := s.objMgr.GetObjectsByType(EntityTypeMonster)
	for _, m := range monsters {
		s.objMgr.RemoveObject(m.ID)
	}
}

// GetStats 获取刷新器统计信息
func (s *Spawner) GetStats() map[string]any {
	return map[string]any{
		"resource_count":   s.objMgr.ObjectCountByType(EntityTypeResource),
		"monster_count":    s.objMgr.ObjectCountByType(EntityTypeMonster),
		"running":          s.running,
		"resource_interval": s.config.ResourceInterval.String(),
		"monster_interval":  s.config.MonsterInterval.String(),
	}
}
