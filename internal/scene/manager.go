package scene

import (
	"fmt"
	"sync"
)

// Manager 场景管理器
type Manager struct {
	scenes map[int64]*Scene
	mutex sync.RWMutex
}

// NewManager 创建场景管理器
func NewManager() *Manager {
	return &Manager{
		scenes: make(map[int64]*Scene),
	}
}

// CreateScene 创建场景
func (m *Manager) CreateScene(cfg SceneConfig) *Scene {
	m.mutex.Lock()
	defer m.mutex.Unlock()

	// 检查是否已存在
	if _, exists := m.scenes[cfg.ID]; exists {
		return nil
	}

	// 创建场景
	scene := NewScene(cfg)
	m.scenes[cfg.ID] = scene

	return scene
}

// GetScene 获取场景
func (m *Manager) GetScene(id int64) *Scene {
	m.mutex.RLock()
	defer m.mutex.RUnlock()
	return m.scenes[id]
}

// RemoveScene 移除场景
func (m *Manager) RemoveScene(id int64) {
	m.mutex.Lock()
	defer m.mutex.Unlock()

	scene, exists := m.scenes[id]
	if !exists {
		return
	}

	// 移除所有实体
	for _, entity := range scene.entities {
		scene.aoi.Leave(entity)
	}

	delete(m.scenes, id)
}

// GetScenes 获取所有场景
func (m *Manager) GetScenes() []*Scene {
	m.mutex.RLock()
	defer m.mutex.RUnlock()

	scenes := make([]*Scene, 0, len(m.scenes))
	for _, s := range m.scenes {
		scenes = append(scenes, s)
	}
	return scenes
}

// SceneCount 获取场景数量
func (m *Manager) SceneCount() int {
	m.mutex.RLock()
	defer m.mutex.RUnlock()
	return len(m.scenes)
}

// AddEntityToScene 添加实体到指定场景
func (m *Manager) AddEntityToScene(sceneID int64, entity *Entity) ([]AOIEvent, error) {
	scene := m.GetScene(sceneID)
	if scene == nil {
		return nil, fmt.Errorf("scene %d not found", sceneID)
	}
	return scene.AddEntity(entity), nil
}

// RemoveEntityFromScene 从指定场景移除实体
func (m *Manager) RemoveEntityFromScene(sceneID int64, entityID int64) ([]AOIEvent, error) {
	scene := m.GetScene(sceneID)
	if scene == nil {
		return nil, fmt.Errorf("scene %d not found", sceneID)
	}
	return scene.RemoveEntity(entityID), nil
}

// MoveEntityInScene 在指定场景移动实体
func (m *Manager) MoveEntityInScene(sceneID int64, entityID int64, newPos Vector2) ([]AOIEvent, error) {
	scene := m.GetScene(sceneID)
	if scene == nil {
		return nil, fmt.Errorf("scene %d not found", sceneID)
	}
	return scene.MoveEntity(entityID, newPos)
}

// GetEntityFromScene 从指定场景获取实体
func (m *Manager) GetEntityFromScene(sceneID int64, entityID int64) (*Entity, error) {
	scene := m.GetScene(sceneID)
	if scene == nil {
		return nil, fmt.Errorf("scene %d not found", sceneID)
	}
	entity := scene.GetEntity(entityID)
	if entity == nil {
		return nil, fmt.Errorf("entity %d not found in scene %d", entityID, sceneID)
	}
	return entity, nil
}

// GetNearbyEntities 获取附近实体
func (m *Manager) GetNearbyEntities(sceneID int64, entityID int64) ([]*Entity, error) {
	scene := m.GetScene(sceneID)
	if scene == nil {
		return nil, fmt.Errorf("scene %d not found", sceneID)
	}
	entities := scene.GetEntitiesInAOI(entityID)
	return entities, nil
}

// ManagerStats 统计信息
type ManagerStats struct {
	SceneCount   int          `json:"scene_count"`
	SceneDetails []SceneInfo  `json:"scene_details"`
}

// GetStats 获取统计信息
func (m *Manager) GetStats() ManagerStats {
	m.mutex.RLock()
	defer m.mutex.RUnlock()

	details := make([]SceneInfo, 0, len(m.scenes))
	for _, s := range m.scenes {
		details = append(details, s.GetInfo())
	}

	return ManagerStats{
		SceneCount:   len(m.scenes),
		SceneDetails: details,
	}
}

// String 返回管理器字符串表示
func (m *Manager) String() string {
	return fmt.Sprintf("SceneManager{Scenes:%d}", len(m.scenes))
}
