package city

import (
	"os"
	"sync"

	"gopkg.in/yaml.v3"
)

type buildingConfigsStore struct {
	byID   map[int]*BuildingConfig
	byType map[int]map[int]*BuildingConfig
	mutex  sync.RWMutex
}

var buildingConfigs = &buildingConfigsStore{
	byID:   make(map[int]*BuildingConfig),
	byType: make(map[int]map[int]*BuildingConfig),
}

func LoadBuildingConfig(path string) error {
	data, err := os.ReadFile(path)
	if err != nil {
		return err
	}

	var cfg struct {
		Buildings []*BuildingConfig `yaml:"buildings"`
	}
	if err := yaml.Unmarshal(data, &cfg); err != nil {
		return err
	}

	buildingConfigs.mutex.Lock()
	defer buildingConfigs.mutex.Unlock()

	buildingConfigs.byID = make(map[int]*BuildingConfig)
	buildingConfigs.byType = make(map[int]map[int]*BuildingConfig)
	for _, b := range cfg.Buildings {
		buildingConfigs.byID[b.ID] = b
		if buildingConfigs.byType[b.Type] == nil {
			buildingConfigs.byType[b.Type] = make(map[int]*BuildingConfig)
		}
		buildingConfigs.byType[b.Type][b.Level] = b
	}
	return nil
}

func GetBuildingConfig(id int) *BuildingConfig {
	buildingConfigs.mutex.RLock()
	defer buildingConfigs.mutex.RUnlock()
	return buildingConfigs.byID[id]
}

func GetBuildingConfigByType(buildingType, level int) *BuildingConfig {
	buildingConfigs.mutex.RLock()
	defer buildingConfigs.mutex.RUnlock()
	if levels, ok := buildingConfigs.byType[buildingType]; ok {
		return levels[level]
	}
	return nil
}

func GetAllBuildingConfigs() []*BuildingConfig {
	buildingConfigs.mutex.RLock()
	defer buildingConfigs.mutex.RUnlock()
	result := make([]*BuildingConfig, 0, len(buildingConfigs.byID))
	for _, cfg := range buildingConfigs.byID {
		result = append(result, cfg)
	}
	return result
}

func MakeBuildingID(buildingType, level int) int {
	return buildingType*100 + level
}
