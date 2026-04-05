package main

import (
	"os"
	"sync"

	"gopkg.in/yaml.v3"
)

// SoldierConfigs 士兵配置管理
type SoldierConfigs struct {
	byID   map[int]*SoldierConfig
	byType map[int]map[int]*SoldierConfig
	mutex  sync.RWMutex
}

var soldierConfigs = &SoldierConfigs{
	byID:   make(map[int]*SoldierConfig),
	byType: make(map[int]map[int]*SoldierConfig),
}

// LoadSoldierConfig 加载士兵配置
func LoadSoldierConfig(path string) error {
	data, err := os.ReadFile(path)
	if err != nil {
		return err
	}

	var cfg struct {
		Soldiers []*SoldierConfig `yaml:"soldiers"`
	}
	if err := yaml.Unmarshal(data, &cfg); err != nil {
		return err
	}

	soldierConfigs.mutex.Lock()
	defer soldierConfigs.mutex.Unlock()

	soldierConfigs.byID = make(map[int]*SoldierConfig)
	soldierConfigs.byType = make(map[int]map[int]*SoldierConfig)

	for _, s := range cfg.Soldiers {
		soldierConfigs.byID[s.ID] = s
		if soldierConfigs.byType[s.Type] == nil {
			soldierConfigs.byType[s.Type] = make(map[int]*SoldierConfig)
		}
		soldierConfigs.byType[s.Type][s.Level] = s
	}
	return nil
}

// GetSoldierConfig 获取士兵配置
func GetSoldierConfig(id int) *SoldierConfig {
	soldierConfigs.mutex.RLock()
	defer soldierConfigs.mutex.RUnlock()
	return soldierConfigs.byID[id]
}

// GetSoldierConfigByType 根据类型和等级获取配置
func GetSoldierConfigByType(soldierType, level int) *SoldierConfig {
	soldierConfigs.mutex.RLock()
	defer soldierConfigs.mutex.RUnlock()
	if t, ok := soldierConfigs.byType[soldierType]; ok {
		return t[level]
	}
	return nil
}

// GetAllSoldierConfigs 获取所有配置
func GetAllSoldierConfigs() []*SoldierConfig {
	soldierConfigs.mutex.RLock()
	defer soldierConfigs.mutex.RUnlock()
	configs := make([]*SoldierConfig, 0, len(soldierConfigs.byID))
	for _, c := range soldierConfigs.byID {
		configs = append(configs, c)
	}
	return configs
}

// GetSoldierConfigsByType 获取指定类型的所有等级配置
func GetSoldierConfigsByType(soldierType int) []*SoldierConfig {
	soldierConfigs.mutex.RLock()
	defer soldierConfigs.mutex.RUnlock()
	if t, ok := soldierConfigs.byType[soldierType]; ok {
		configs := make([]*SoldierConfig, 0, len(t))
		for _, c := range t {
			configs = append(configs, c)
		}
		return configs
	}
	return nil
}

// MakeSoldierID 生成士兵ID
func MakeSoldierID(soldierType, level int) int {
	return soldierType*100 + level
}

// ParseSoldierID 解析士兵ID
func ParseSoldierID(id int) (soldierType, level int) {
	level = id % 100
	soldierType = id / 100
	return
}
