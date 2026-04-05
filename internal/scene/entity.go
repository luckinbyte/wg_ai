package scene

import (
	"fmt"
	"sync/atomic"
)

// 全局实体ID生成器
var globalEntityID int64

// GenerateEntityID 生成唯一实体ID
func GenerateEntityID() int64 {
	return atomic.AddInt64(&globalEntityID, 1)
}

// NewEntity 创建新实体
func NewEntity(id int64, entityType EntityType, pos Vector2) *Entity {
	return &Entity{
		ID:       id,
		Type:     entityType,
		Position: pos,
		Data:     make(map[string]any),
	}
}

// NewEntityWithScene 创建带场景ID的实体
func NewEntityWithScene(id int64, entityType EntityType, pos Vector2, sceneID int64) *Entity {
	return &Entity{
		ID:       id,
		Type:     entityType,
		Position: pos,
		SceneID:  sceneID,
		Data:     make(map[string]any),
	}
}

// CreatePlayerEntity 创建玩家实体
func CreatePlayerEntity(rid int64, pos Vector2, sceneID int64) *Entity {
	return &Entity{
		ID:       rid,
		Type:     EntityTypePlayer,
		Position: pos,
		SceneID:  sceneID,
		Data:     make(map[string]any),
	}
}

// CreateNPCEntity 创建NPC实体
func CreateNPCEntity(npcID int64, pos Vector2, sceneID int64, cfgID int) *Entity {
	return &Entity{
		ID:       npcID,
		Type:     EntityTypeNPC,
		Position: pos,
		SceneID:  sceneID,
		Data: map[string]any{
			"cfg_id": cfgID,
		},
	}
}

// CreateMonsterEntity 创建怪物实体
func CreateMonsterEntity(monsterID int64, pos Vector2, sceneID int64, cfgID int) *Entity {
	return &Entity{
		ID:       monsterID,
		Type:     EntityTypeMonster,
		Position: pos,
		SceneID:  sceneID,
		Data: map[string]any{
			"cfg_id": cfgID,
		},
	}
}

// String 返回实体字符串表示
func (e *Entity) String() string {
	return fmt.Sprintf("Entity{ID:%d, Type:%s, Pos:(%.1f,%.1f), Scene:%d}",
		e.ID, e.Type, e.Position.X, e.Position.Y, e.SceneID)
}
