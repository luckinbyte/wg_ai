package scene

import (
	"fmt"
	"sync"
)

// Scene 场景
type Scene struct {
	ID       int64
	Width    int    // 场景宽度
	Height   int    // 场景高度
	GridSize int    // 格子大小
	aoi      *AOI   // AOI管理器

	entities map[int64]*Entity // 实体ID -> 实体
	mutex    sync.RWMutex

	// 地图对象管理
	objMgr   *MapObjectManager
	spawner  *Spawner
}

// NewScene 创建新场景
func NewScene(cfg SceneConfig) *Scene {
	s := &Scene{
		ID:       cfg.ID,
		Width:    cfg.Width,
		Height:   cfg.Height,
		GridSize: cfg.GridSize,
		entities: make(map[int64]*Entity),
	}

	// 使用默认格子大小
	if s.GridSize <= 0 {
		s.GridSize = 50
	}

	// 创建AOI
	s.aoi = NewAOI(s.Width, s.Height, s.GridSize)

	// 创建地图对象管理器
	s.objMgr = NewMapObjectManager(s)

	// 创建刷新器 (使用默认配置)
	s.spawner = NewSpawner(s, s.objMgr, DefaultSpawnerConfig)

	return s
}

// AddEntity 添加实体到场景
func (s *Scene) AddEntity(entity *Entity) []AOIEvent {
	s.mutex.Lock()
	defer s.mutex.Unlock()

	// 设置实体场景ID
	entity.SceneID = s.ID

	// 存储实体
	s.entities[entity.ID] = entity

	// 加入AOI并获取事件
	events := s.aoi.Enter(entity)

	return events
}

// RemoveEntity 从场景移除实体
func (s *Scene) RemoveEntity(entityID int64) []AOIEvent {
	s.mutex.Lock()
	defer s.mutex.Unlock()

	entity, exists := s.entities[entityID]
	if !exists {
		return nil
	}

	// 从AOI移除并获取事件
	events := s.aoi.Leave(entity)

	// 删除实体
	delete(s.entities, entityID)

	return events
}

// MoveEntity 移动实体
func (s *Scene) MoveEntity(entityID int64, newPos Vector2) ([]AOIEvent, error) {
	s.mutex.Lock()
	defer s.mutex.Unlock()

	entity, exists := s.entities[entityID]
	if !exists {
		return nil, fmt.Errorf("entity %d not found", entityID)
	}

	// 边界检查
	if newPos.X < 0 || newPos.X > float64(s.Width) ||
		newPos.Y < 0 || newPos.Y > float64(s.Height) {
		return nil, fmt.Errorf("position (%.1f, %.1f) out of bounds", newPos.X, newPos.Y)
	}

	// 移动并获取视野变化事件
	events := s.aoi.Move(entity, newPos)

	return events, nil
}

// GetEntity 获取实体
func (s *Scene) GetEntity(entityID int64) *Entity {
	s.mutex.RLock()
	defer s.mutex.RUnlock()
	return s.entities[entityID]
}

// GetEntities 获取所有实体
func (s *Scene) GetEntities() []*Entity {
	s.mutex.RLock()
	defer s.mutex.RUnlock()

	entities := make([]*Entity, 0, len(s.entities))
	for _, e := range s.entities {
		entities = append(entities, e)
	}
	return entities
}

// GetEntitiesInAOI 获取指定实体视野内的其他实体
func (s *Scene) GetEntitiesInAOI(entityID int64) []*Entity {
	s.mutex.RLock()
	defer s.mutex.RUnlock()

	entity, exists := s.entities[entityID]
	if !exists {
		return nil
	}

	return s.aoi.GetEntitiesInAOI(entity)
}

// GetNearbyPlayers 获取附近玩家
func (s *Scene) GetNearbyPlayers(entityID int64) []*Entity {
	s.mutex.RLock()
	defer s.mutex.RUnlock()

	entity, exists := s.entities[entityID]
	if !exists {
		return nil
	}

	entities := s.aoi.GetEntitiesInAOI(entity)
	players := make([]*Entity, 0)
	for _, e := range entities {
		if e.Type == EntityTypePlayer {
			players = append(players, e)
		}
	}
	return players
}

// EntityCount 获取实体数量
func (s *Scene) EntityCount() int {
	s.mutex.RLock()
	defer s.mutex.RUnlock()
	return len(s.entities)
}

// GetAOI 获取AOI管理器
func (s *Scene) GetAOI() *AOI {
	return s.aoi
}

// GetStats 获取场景统计信息
func (s *Scene) GetStats() map[string]any {
	s.mutex.RLock()
	defer s.mutex.RUnlock()

	return map[string]any{
		"scene_id":       s.ID,
		"width":          s.Width,
		"height":         s.Height,
		"grid_size":      s.GridSize,
		"entity_count":   len(s.entities),
		"aoi_stats":      s.aoi.GetStats(),
	}
}

// Broadcast 向视野内的玩家广播消息
// 返回需要接收消息的玩家ID列表
func (s *Scene) Broadcast(entityID int64, includeSelf bool) []int64 {
	s.mutex.RLock()
	defer s.mutex.RUnlock()

	entity, exists := s.entities[entityID]
	if !exists {
		return nil
	}

	// 获取视野内的所有实体
	entities := s.aoi.GetEntitiesInAOI(entity)

	// 筛选玩家
	playerIDs := make([]int64, 0)
	for _, e := range entities {
		if e.Type == EntityTypePlayer {
			playerIDs = append(playerIDs, e.ID)
		}
	}

	// 是否包含自己
	if includeSelf && entity.Type == EntityTypePlayer {
		playerIDs = append(playerIDs, entity.ID)
	}

	return playerIDs
}

// String 返回场景字符串表示
func (s *Scene) String() string {
	return fmt.Sprintf("Scene{ID:%d, Size:%dx%d, GridSize:%d, Entities:%d}",
		s.ID, s.Width, s.Height, s.GridSize, s.EntityCount())
}

// SceneInfo 场景信息
type SceneInfo struct {
	ID          int64 `json:"id"`
	Width       int   `json:"width"`
	Height      int   `json:"height"`
	GridSize    int   `json:"grid_size"`
	EntityCount int   `json:"entity_count"`
}

// GetInfo 获取场景信息
func (s *Scene) GetInfo() SceneInfo {
	s.mutex.RLock()
	defer s.mutex.RUnlock()

	return SceneInfo{
		ID:          s.ID,
		Width:       s.Width,
		Height:      s.Height,
		GridSize:    s.GridSize,
		EntityCount: len(s.entities),
	}
}

// GetObjectManager 获取地图对象管理器
func (s *Scene) GetObjectManager() *MapObjectManager {
	return s.objMgr
}

// GetSpawner 获取刷新器
func (s *Scene) GetSpawner() *Spawner {
	return s.spawner
}

// StartSpawner 启动刷新器
func (s *Scene) StartSpawner() {
	if s.spawner != nil {
		s.spawner.Start()
	}
}

// StopSpawner 停止刷新器
func (s *Scene) StopSpawner() {
	if s.spawner != nil {
		s.spawner.Stop()
	}
}
