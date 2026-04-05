package scene

import (
	"sync"
	"sync/atomic"
)

// 全局对象ID生成器
var globalObjectID int64

// GenerateObjectID 生成唯一对象ID
func GenerateObjectID() int64 {
	return atomic.AddInt64(&globalObjectID, 1)
}

// MapObjectManager 地图对象管理器
type MapObjectManager struct {
	scene    *Scene
	objects  map[int64]*Entity        // objectID -> Entity
	byType   map[EntityType]map[int64]*Entity // 按类型索引
	mutex    sync.RWMutex
}

// NewMapObjectManager 创建地图对象管理器
func NewMapObjectManager(scene *Scene) *MapObjectManager {
	return &MapObjectManager{
		scene:   scene,
		objects: make(map[int64]*Entity),
		byType:  make(map[EntityType]map[int64]*Entity),
	}
}

// SpawnResource 生成资源点
func (m *MapObjectManager) SpawnResource(pos Vector2, resType ResourceType, level int, amount int64) *Entity {
	m.mutex.Lock()
	defer m.mutex.Unlock()

	id := GenerateObjectID()
	entity := &Entity{
		ID:       id,
		Type:     EntityTypeResource,
		Position: pos,
		SceneID:  m.scene.ID,
		Data:     make(map[string]any),
	}

	// 设置资源数据
	objData := &MapObjectData{
		Level:        level,
		ResourceType: resType,
		Amount:       amount,
		MaxAmount:    amount,
		OwnerID:      0,
	}
	entity.SetData("object_data", objData)

	// 存储对象
	m.objects[id] = entity
	m.addToTypeIndex(entity)

	// 加入场景AOI
	m.scene.aoi.Enter(entity)

	return entity
}

// SpawnMonster 生成怪物
func (m *MapObjectManager) SpawnMonster(pos Vector2, cfgID int, level int, power int64) *Entity {
	m.mutex.Lock()
	defer m.mutex.Unlock()

	id := GenerateObjectID()
	entity := &Entity{
		ID:       id,
		Type:     EntityTypeMonster,
		Position: pos,
		SceneID:  m.scene.ID,
		Data:     make(map[string]any),
	}

	// 设置怪物数据
	objData := &MapObjectData{
		ConfigID: cfgID,
		Level:    level,
		Amount:   power, // 用Amount存储战力
	}
	entity.SetData("object_data", objData)

	// 存储对象
	m.objects[id] = entity
	m.addToTypeIndex(entity)

	// 加入场景AOI
	m.scene.aoi.Enter(entity)

	return entity
}

// SpawnBuilding 生成建筑
func (m *MapObjectManager) SpawnBuilding(pos Vector2, cfgID int, ownerID int64) *Entity {
	m.mutex.Lock()
	defer m.mutex.Unlock()

	id := GenerateObjectID()
	entity := &Entity{
		ID:       id,
		Type:     EntityTypeBuilding,
		Position: pos,
		SceneID:  m.scene.ID,
		Data:     make(map[string]any),
	}

	// 设置建筑数据
	objData := &MapObjectData{
		ConfigID: cfgID,
		OwnerID:  ownerID,
	}
	entity.SetData("object_data", objData)

	// 存储对象
	m.objects[id] = entity
	m.addToTypeIndex(entity)

	// 加入场景AOI
	m.scene.aoi.Enter(entity)

	return entity
}

// RemoveObject 移除对象
func (m *MapObjectManager) RemoveObject(objectID int64) {
	m.mutex.Lock()
	defer m.mutex.Unlock()

	entity, exists := m.objects[objectID]
	if !exists {
		return
	}

	// 从AOI移除
	m.scene.aoi.Leave(entity)

	// 从索引移除
	delete(m.objects, objectID)
	m.removeFromTypeIndex(entity)
}

// GetObject 获取对象
func (m *MapObjectManager) GetObject(objectID int64) *Entity {
	m.mutex.RLock()
	defer m.mutex.RUnlock()
	return m.objects[objectID]
}

// GetObjectsByType 获取指定类型的所有对象
func (m *MapObjectManager) GetObjectsByType(objType EntityType) []*Entity {
	m.mutex.RLock()
	defer m.mutex.RUnlock()

	entities := m.byType[objType]
	if entities == nil {
		return nil
	}

	result := make([]*Entity, 0, len(entities))
	for _, e := range entities {
		result = append(result, e)
	}
	return result
}

// GetObjectsInRadius 获取指定半径内的对象
func (m *MapObjectManager) GetObjectsInRadius(pos Vector2, radius float64) []*Entity {
	m.mutex.RLock()
	defer m.mutex.RUnlock()

	result := make([]*Entity, 0)
	for _, entity := range m.objects {
		dist := Distance(pos, entity.Position)
		if dist <= radius {
			result = append(result, entity)
		}
	}
	return result
}

// GetNearestObject 获取最近的对象
func (m *MapObjectManager) GetNearestObject(pos Vector2, objType EntityType, maxRadius float64) *Entity {
	m.mutex.RLock()
	defer m.mutex.RUnlock()

	entities := m.byType[objType]
	if entities == nil {
		return nil
	}

	var nearest *Entity
	minDist := maxRadius

	for _, entity := range entities {
		dist := Distance(pos, entity.Position)
		if dist < minDist {
			minDist = dist
			nearest = entity
		}
	}

	return nearest
}

// UpdateObjectOwner 更新对象占领者
func (m *MapObjectManager) UpdateObjectOwner(objectID int64, ownerID int64) bool {
	m.mutex.Lock()
	defer m.mutex.Unlock()

	entity, exists := m.objects[objectID]
	if !exists {
		return false
	}

	objData := getObjectData(entity)
	if objData == nil {
		return false
	}

	objData.OwnerID = ownerID
	return true
}

// UpdateObjectAmount 更新对象数量 (资源采集用)
func (m *MapObjectManager) UpdateObjectAmount(objectID int64, amount int64) bool {
	m.mutex.Lock()
	defer m.mutex.Unlock()

	entity, exists := m.objects[objectID]
	if !exists {
		return false
	}

	objData := getObjectData(entity)
	if objData == nil {
		return false
	}

	objData.Amount = amount
	return true
}

// ObjectCount 获取对象数量
func (m *MapObjectManager) ObjectCount() int {
	m.mutex.RLock()
	defer m.mutex.RUnlock()
	return len(m.objects)
}

// ObjectCountByType 获取指定类型对象数量
func (m *MapObjectManager) ObjectCountByType(objType EntityType) int {
	m.mutex.RLock()
	defer m.mutex.RUnlock()

	entities := m.byType[objType]
	if entities == nil {
		return 0
	}
	return len(entities)
}

// addToTypeIndex 添加到类型索引
func (m *MapObjectManager) addToTypeIndex(entity *Entity) {
	if m.byType[entity.Type] == nil {
		m.byType[entity.Type] = make(map[int64]*Entity)
	}
	m.byType[entity.Type][entity.ID] = entity
}

// removeFromTypeIndex 从类型索引移除
func (m *MapObjectManager) removeFromTypeIndex(entity *Entity) {
	entities := m.byType[entity.Type]
	if entities != nil {
		delete(entities, entity.ID)
	}
}

// getObjectData 获取对象的MapObjectData
func getObjectData(entity *Entity) *MapObjectData {
	if entity.Data == nil {
		return nil
	}
	data, ok := entity.Data["object_data"]
	if !ok {
		return nil
	}
	objData, ok := data.(*MapObjectData)
	if !ok {
		return nil
	}
	return objData
}

// GetObjectData 获取对象的MapObjectData (导出)
func (entity *Entity) GetObjectData() *MapObjectData {
	return getObjectData(entity)
}
