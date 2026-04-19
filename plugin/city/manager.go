package city

import (
	"fmt"
	"sync"
	"sync/atomic"
	"time"

	"github.com/luckinbyte/wg_ai/internal/scene"
	baseplugin "github.com/luckinbyte/wg_ai/plugin"
)

var buildQueueSeq int64

type Manager struct {
	sceneMgr *scene.Manager
	mutex    sync.RWMutex
}

func NewManager(sceneMgr *scene.Manager) *Manager {
	return &Manager{sceneMgr: sceneMgr}
}

func (m *Manager) GetCity(data baseplugin.DataAccessor) (*CityData, error) {
	raw, err := data.GetArray("city")
	if err != nil || raw == nil {
		return nil, nil
	}
	city, ok := raw.(*CityData)
	if !ok {
		return nil, fmt.Errorf("invalid city data")
	}
	return city, nil
}

func (m *Manager) InitCity(data baseplugin.DataAccessor, x, y float64) (*CityData, error) {
	city, err := m.GetCity(data)
	if err != nil {
		return nil, err
	}
	if city != nil {
		// 城池数据已存在，重新将建筑实体 Spawn 回大地图
		m.spawnCityToScene(city, data)
		return city, nil
	}
	if m.sceneMgr == nil {
		return nil, fmt.Errorf("scene manager not initialized")
	}

	s := m.sceneMgr.GetScene(1)
	if s == nil {
		s = m.sceneMgr.CreateScene(scene.SceneConfig{ID: 1, Width: 1000, Height: 1000, GridSize: 50})
	}

	cfg := GetBuildingConfigByType(int(BuildingTypeCastle), 1)
	if cfg == nil {
		return nil, fmt.Errorf("castle config not found")
	}

	entity := s.GetObjectManager().SpawnBuilding(scene.Vector2{X: x, Y: y}, cfg.ID, getRIDFromData(data))
	objData := entity.GetObjectData()
	if objData != nil {
		objData.Level = 1
	}
	entity.SetData("name", cfg.Name)
	entity.SetData("building_type", cfg.Type)
	entity.SetData("building_level", cfg.Level)

	city = &CityData{
		CityID:   entity.ID,
		Position: Position{X: x, Y: y},
		Buildings: map[int]*BuildingData{
			int(BuildingTypeCastle): {
				Type:     int(BuildingTypeCastle),
				Level:    1,
				HP:       cfg.HP,
				EntityID: entity.ID,
			},
		},
		BuildQueue: []BuildQueueItem{},
	}
	if err := data.SetArray("city", city); err != nil {
		return nil, err
	}

	m.ensureDefaultResources(data)
	return city, nil
}

// spawnCityToScene 将已有城池的建筑实体重新生成到大地图上
func (m *Manager) spawnCityToScene(city *CityData, data baseplugin.DataAccessor) {
	if m.sceneMgr == nil {
		return
	}
	s := m.sceneMgr.GetScene(1)
	if s == nil {
		return
	}
	rid := getRIDFromData(data)
	pos := scene.Vector2{X: city.Position.X, Y: city.Position.Y}

	for _, b := range city.Buildings {
		cfg := GetBuildingConfigByType(b.Type, b.Level)
		if cfg == nil {
			continue
		}
		entity := s.GetObjectManager().SpawnBuilding(pos, cfg.ID, rid)
		objData := entity.GetObjectData()
		if objData != nil {
			objData.Level = b.Level
		}
		entity.SetData("building_type", b.Type)
		entity.SetData("building_level", b.Level)
		// 更新 EntityID 以保持一致
		b.EntityID = entity.ID
	}
	city.CityID = city.Buildings[int(BuildingTypeCastle)].EntityID
}

func (m *Manager) UpgradeBuilding(data baseplugin.DataAccessor, buildingType int) (*BuildQueueItem, error) {
	city, err := m.GetCity(data)
	if err != nil {
		return nil, err
	}
	if city == nil {
		return nil, fmt.Errorf("city not initialized")
	}
	if len(city.BuildQueue) > 0 {
		return nil, fmt.Errorf("build queue is full")
	}

	currentLevel := 0
	if building, ok := city.Buildings[buildingType]; ok {
		currentLevel = building.Level
	}
	targetLevel := currentLevel + 1
	cfg := GetBuildingConfigByType(buildingType, targetLevel)
	if cfg == nil {
		return nil, fmt.Errorf("building config not found")
	}
	if err := m.consumeResources(data, cfg); err != nil {
		return nil, err
	}

	now := time.Now().Unix()
	item := BuildQueueItem{
		ID:           atomic.AddInt64(&buildQueueSeq, 1),
		BuildingType: buildingType,
		TargetLevel:  targetLevel,
		StartTime:    now,
		FinishTime:   now + int64(cfg.BuildTime),
	}
	city.BuildQueue = append(city.BuildQueue, item)
	data.MarkDirty()
	return &item, nil
}

func (m *Manager) CancelBuild(data baseplugin.DataAccessor, queueID int64) error {
	city, err := m.GetCity(data)
	if err != nil {
		return err
	}
	if city == nil {
		return fmt.Errorf("city not initialized")
	}
	for i, item := range city.BuildQueue {
		if item.ID != queueID {
			continue
		}
		cfg := GetBuildingConfigByType(item.BuildingType, item.TargetLevel)
		if cfg != nil {
			m.refundResources(data, cfg, 50)
		}
		city.BuildQueue = append(city.BuildQueue[:i], city.BuildQueue[i+1:]...)
		data.MarkDirty()
		return nil
	}
	return fmt.Errorf("build queue not found")
}

func (m *Manager) GetBuildQueue(data baseplugin.DataAccessor) ([]BuildQueueItem, error) {
	city, err := m.GetCity(data)
	if err != nil {
		return nil, err
	}
	if city == nil {
		return nil, nil
	}
	return city.BuildQueue, nil
}

func (m *Manager) CompleteReadyBuilds(data baseplugin.DataAccessor) ([]BuildQueueItem, error) {
	city, err := m.GetCity(data)
	if err != nil {
		return nil, err
	}
	if city == nil || len(city.BuildQueue) == 0 {
		return nil, nil
	}

	now := time.Now().Unix()
	completed := make([]BuildQueueItem, 0)
	remaining := make([]BuildQueueItem, 0, len(city.BuildQueue))
	for _, item := range city.BuildQueue {
		if item.FinishTime > now {
			remaining = append(remaining, item)
			continue
		}
		cfg := GetBuildingConfigByType(item.BuildingType, item.TargetLevel)
		if cfg == nil {
			continue
		}
		building, ok := city.Buildings[item.BuildingType]
		if !ok {
			building = &BuildingData{Type: item.BuildingType}
			city.Buildings[item.BuildingType] = building
		}
		building.Level = item.TargetLevel
		building.HP = cfg.HP
		completed = append(completed, item)

		if item.BuildingType == int(BuildingTypeCastle) {
			s := m.sceneMgr.GetScene(1)
			if s != nil {
				entity := s.GetObjectManager().GetObject(city.CityID)
				if entity != nil {
					if objData := entity.GetObjectData(); objData != nil {
						objData.ConfigID = cfg.ID
						objData.Level = cfg.Level
					}
					entity.SetData("name", cfg.Name)
					entity.SetData("building_level", cfg.Level)
				}
			}
		}
	}
	city.BuildQueue = remaining
	if len(completed) > 0 {
		data.MarkDirty()
	}
	return completed, nil
}

func (m *Manager) GetProduction(data baseplugin.DataAccessor) (int64, int64, int64, int64, error) {
	city, err := m.GetCity(data)
	if err != nil {
		return 0, 0, 0, 0, err
	}
	if city == nil {
		return 0, 0, 0, 0, nil
	}
	var food, wood, stone, gold int64
	for buildingType, building := range city.Buildings {
		cfg := GetBuildingConfigByType(buildingType, building.Level)
		if cfg == nil {
			continue
		}
		food += cfg.ProdFood
		wood += cfg.ProdWood
		stone += cfg.ProdStone
		gold += cfg.ProdGold
	}
	return food, wood, stone, gold, nil
}

func (m *Manager) getResource(data baseplugin.DataAccessor, key string) int64 {
	if val, err := data.GetField(key); err == nil {
		switch v := val.(type) {
		case int64:
			return v
		case int:
			return int64(v)
		case float64:
			return int64(v)
		}
	}
	return 0
}

func (m *Manager) setResource(data baseplugin.DataAccessor, key string, val int64) {
	data.SetField(key, val)
}

func (m *Manager) ensureDefaultResources(data baseplugin.DataAccessor) {
	if m.getResource(data, "food") == 0 {
		m.setResource(data, "food", 1000)
	}
	if m.getResource(data, "wood") == 0 {
		m.setResource(data, "wood", 1000)
	}
	if m.getResource(data, "stone") == 0 {
		m.setResource(data, "stone", 500)
	}
	if m.getResource(data, "gold") == 0 {
		m.setResource(data, "gold", 200)
	}
}

func (m *Manager) consumeResources(data baseplugin.DataAccessor, cfg *BuildingConfig) error {
	food := m.getResource(data, "food")
	wood := m.getResource(data, "wood")
	stone := m.getResource(data, "stone")
	gold := m.getResource(data, "gold")
	if food < cfg.CostFood || wood < cfg.CostWood || stone < cfg.CostStone || gold < cfg.CostGold {
		return fmt.Errorf("not enough resources")
	}
	m.setResource(data, "food", food-cfg.CostFood)
	m.setResource(data, "wood", wood-cfg.CostWood)
	m.setResource(data, "stone", stone-cfg.CostStone)
	m.setResource(data, "gold", gold-cfg.CostGold)
	return nil
}

func (m *Manager) refundResources(data baseplugin.DataAccessor, cfg *BuildingConfig, percent int64) {
	m.setResource(data, "food", m.getResource(data, "food")+cfg.CostFood*percent/100)
	m.setResource(data, "wood", m.getResource(data, "wood")+cfg.CostWood*percent/100)
	m.setResource(data, "stone", m.getResource(data, "stone")+cfg.CostStone*percent/100)
	m.setResource(data, "gold", m.getResource(data, "gold")+cfg.CostGold*percent/100)
}

func getRIDFromData(data baseplugin.DataAccessor) int64 {
	if adapter, ok := data.(baseplugin.DataAdapter); ok {
		return adapter.RID()
	}
	if ridAny, err := data.GetField("rid"); err == nil {
		switch v := ridAny.(type) {
		case int64:
			return v
		case int:
			return int64(v)
		case float64:
			return int64(v)
		}
	}
	return 0
}
