package scene

import (
	"testing"
)

// TestAOIPositionToGrid 测试坐标转格子
func TestAOIPositionToGrid(t *testing.T) {
	aoi := NewAOI(1000, 1000, 50)

	tests := []struct {
		pos      Vector2
		expected [2]int
	}{
		{Vector2{X: 0, Y: 0}, [2]int{0, 0}},
		{Vector2{X: 49, Y: 49}, [2]int{0, 0}},
		{Vector2{X: 50, Y: 50}, [2]int{1, 1}},
		{Vector2{X: 250, Y: 250}, [2]int{5, 5}},
		{Vector2{X: 999, Y: 999}, [2]int{19, 19}},
		{Vector2{X: -10, Y: -10}, [2]int{0, 0}},         // 边界外
		{Vector2{X: 10000, Y: 10000}, [2]int{19, 19}},   // 边界外
	}

	for _, test := range tests {
		gx, gy := aoi.posToGrid(test.pos)
		if gx != test.expected[0] || gy != test.expected[1] {
			t.Errorf("posToGrid(%v) = (%d, %d), expected %v",
				test.pos, gx, gy, test.expected)
		}
	}
}

// TestAOIGetNeighbors 测试获取九宫格
func TestAOIGetNeighbors(t *testing.T) {
	aoi := NewAOI(1000, 1000, 50)

	// 中心格子
	grids := aoi.GetNeighbors(5, 5)
	if len(grids) != 9 {
		t.Errorf("GetNeighbors(5, 5) = %d grids, expected 9", len(grids))
	}

	// 角落格子 (0, 0) - 只有4个格子
	grids = aoi.GetNeighbors(0, 0)
	if len(grids) != 4 {
		t.Errorf("GetNeighbors(0, 0) = %d grids, expected 4", len(grids))
	}

	// 边缘格子 (0, 5) - 只有6个格子
	grids = aoi.GetNeighbors(0, 5)
	if len(grids) != 6 {
		t.Errorf("GetNeighbors(0, 5) = %d grids, expected 6", len(grids))
	}
}

// TestAOIEnterLeave 测试实体进入和离开
func TestAOIEnterLeave(t *testing.T) {
	aoi := NewAOI(1000, 1000, 50)

	// 创建两个实体，在同一格子内
	entity1 := NewEntity(1, EntityTypePlayer, Vector2{X: 10, Y: 10})
	entity2 := NewEntity(2, EntityTypePlayer, Vector2{X: 20, Y: 20})

	// 实体1进入
	events1 := aoi.Enter(entity1)
	if len(events1) != 0 {
		t.Errorf("First entity enter should have no events, got %d", len(events1))
	}

	// 实体2进入，应该看到实体1
	events2 := aoi.Enter(entity2)
	if len(events2) != 2 {
		t.Errorf("Second entity enter should have 2 events, got %d", len(events2))
	}

	// 检查事件
	enterCount := 0
	for _, e := range events2 {
		if e.Type == EventEnter {
			enterCount++
		}
	}
	if enterCount != 2 {
		t.Errorf("Expected 2 enter events, got %d", enterCount)
	}

	// 实体1离开，应该通知实体2
	events3 := aoi.Leave(entity1)
	if len(events3) != 2 {
		t.Errorf("Entity leave should have 2 events, got %d", len(events3))
	}
}

// TestAOIMoveSameGrid 测试同一格子内移动
func TestAOIMoveSameGrid(t *testing.T) {
	aoi := NewAOI(1000, 1000, 50)

	entity := NewEntity(1, EntityTypePlayer, Vector2{X: 10, Y: 10})
	aoi.Enter(entity)

	// 在同一格子内移动
	events := aoi.Move(entity, Vector2{X: 40, Y: 40})
	if len(events) != 0 {
		t.Errorf("Move within same grid should have no events, got %d", len(events))
	}
}

// TestAOIMoveCrossGrid 测试跨格子移动
func TestAOIMoveCrossGrid(t *testing.T) {
	aoi := NewAOI(1000, 1000, 50)

	// 创建两个实体，在相邻格子
	entity1 := NewEntity(1, EntityTypePlayer, Vector2{X: 10, Y: 10})   // 格子 (0, 0)
	entity2 := NewEntity(2, EntityTypePlayer, Vector2{X: 110, Y: 110}) // 格子 (2, 2)

	aoi.Enter(entity1)
	aoi.Enter(entity2)

	// 它们应该互相看不到 (不在九宫格内)
	entities1 := aoi.GetEntitiesInAOI(entity1)
	if len(entities1) != 0 {
		t.Errorf("Entity1 should see 0 entities, got %d", len(entities1))
	}

	// 移动 entity1 到格子 (1, 1)，现在它们在九宫格内了
	events := aoi.Move(entity1, Vector2{X: 60, Y: 60})
	if len(events) != 2 {
		t.Errorf("Move to adjacent grid should have 2 events, got %d", len(events))
	}

	// 检查现在能互相看到
	entities1 = aoi.GetEntitiesInAOI(entity1)
	if len(entities1) != 1 {
		t.Errorf("Entity1 should see 1 entity now, got %d", len(entities1))
	}
}

// TestSceneAddRemove 测试场景添加/移除实体
func TestSceneAddRemove(t *testing.T) {
	scene := NewScene(SceneConfig{
		ID:       1,
		Width:    1000,
		Height:   1000,
		GridSize: 50,
	})

	entity1 := CreatePlayerEntity(1, Vector2{X: 100, Y: 100}, 1)
	entity2 := CreatePlayerEntity(2, Vector2{X: 110, Y: 110}, 1)

	// 添加实体1
	events1 := scene.AddEntity(entity1)
	if len(events1) != 0 {
		t.Errorf("First entity add should have no events, got %d", len(events1))
	}

	// 添加实体2，应该看到实体1
	events2 := scene.AddEntity(entity2)
	if len(events2) == 0 {
		t.Error("Second entity add should have events")
	}

	// 检查实体数量
	if scene.EntityCount() != 2 {
		t.Errorf("Scene should have 2 entities, got %d", scene.EntityCount())
	}

	// 获取实体
	e := scene.GetEntity(1)
	if e == nil || e.ID != 1 {
		t.Error("GetEntity failed")
	}

	// 移除实体
	events3 := scene.RemoveEntity(1)
	if len(events3) == 0 {
		t.Error("Entity remove should have events")
	}

	if scene.EntityCount() != 1 {
		t.Errorf("Scene should have 1 entity, got %d", scene.EntityCount())
	}
}

// TestSceneMove 测试场景内移动
func TestSceneMove(t *testing.T) {
	scene := NewScene(SceneConfig{
		ID:       1,
		Width:    1000,
		Height:   1000,
		GridSize: 50,
	})

	entity1 := CreatePlayerEntity(1, Vector2{X: 10, Y: 10}, 1)
	entity2 := CreatePlayerEntity(2, Vector2{X: 110, Y: 110}, 1)

	scene.AddEntity(entity1)
	scene.AddEntity(entity2)

	// 初始时互相看不到
	nearby := scene.GetEntitiesInAOI(1)
	if len(nearby) != 0 {
		t.Errorf("Entity1 should see 0 entities, got %d", len(nearby))
	}

	// 移动到能互相看到的位置
	events, err := scene.MoveEntity(1, Vector2{X: 60, Y: 60})
	if err != nil {
		t.Errorf("MoveEntity failed: %v", err)
	}

	// 应该有视野变化事件
	if len(events) == 0 {
		t.Error("Move should generate events")
	}

	// 现在应该能看到
	nearby = scene.GetEntitiesInAOI(1)
	if len(nearby) != 1 {
		t.Errorf("Entity1 should see 1 entity now, got %d", len(nearby))
	}
}

// TestSceneBoundary 测试边界检查
func TestSceneBoundary(t *testing.T) {
	scene := NewScene(SceneConfig{
		ID:       1,
		Width:    1000,
		Height:   1000,
		GridSize: 50,
	})

	entity := CreatePlayerEntity(1, Vector2{X: 100, Y: 100}, 1)
	scene.AddEntity(entity)

	// 移动到边界外
	_, err := scene.MoveEntity(1, Vector2{X: -10, Y: -10})
	if err == nil {
		t.Error("Move out of bounds should fail")
	}

	_, err = scene.MoveEntity(1, Vector2{X: 10000, Y: 10000})
	if err == nil {
		t.Error("Move out of bounds should fail")
	}
}

// TestManagerCreateScene 测试场景管理器
func TestManagerCreateScene(t *testing.T) {
	mgr := NewManager()

	// 创建场景
	scene := mgr.CreateScene(SceneConfig{
		ID:       1,
		Width:    1000,
		Height:   1000,
		GridSize: 50,
	})

	if scene == nil {
		t.Fatal("CreateScene failed")
	}

	if mgr.SceneCount() != 1 {
		t.Errorf("Manager should have 1 scene, got %d", mgr.SceneCount())
	}

	// 重复创建应该返回 nil
	scene2 := mgr.CreateScene(SceneConfig{
		ID:       1,
		Width:    1000,
		Height:   1000,
		GridSize: 50,
	})

	if scene2 != nil {
		t.Error("Duplicate scene creation should return nil")
	}
}

// TestManagerGetRemoveScene 测试获取和移除场景
func TestManagerGetRemoveScene(t *testing.T) {
	mgr := NewManager()

	mgr.CreateScene(SceneConfig{ID: 1, Width: 1000, Height: 1000, GridSize: 50})
	mgr.CreateScene(SceneConfig{ID: 2, Width: 1000, Height: 1000, GridSize: 50})

	// 获取场景
	scene := mgr.GetScene(1)
	if scene == nil || scene.ID != 1 {
		t.Error("GetScene failed")
	}

	// 获取不存在的场景
	scene = mgr.GetScene(999)
	if scene != nil {
		t.Error("GetScene for non-existent scene should return nil")
	}

	// 移除场景
	mgr.RemoveScene(1)
	if mgr.SceneCount() != 1 {
		t.Errorf("Manager should have 1 scene, got %d", mgr.SceneCount())
	}

	scene = mgr.GetScene(1)
	if scene != nil {
		t.Error("Removed scene should not be found")
	}
}

// TestManagerEntityOperations 测试管理器的实体操作
func TestManagerEntityOperations(t *testing.T) {
	mgr := NewManager()
	mgr.CreateScene(SceneConfig{ID: 1, Width: 1000, Height: 1000, GridSize: 50})

	entity := CreatePlayerEntity(1, Vector2{X: 100, Y: 100}, 1)

	// 添加实体
	events, err := mgr.AddEntityToScene(1, entity)
	if err != nil {
		t.Errorf("AddEntityToScene failed: %v", err)
	}
	_ = events

	// 获取实体
	e, err := mgr.GetEntityFromScene(1, 1)
	if err != nil || e == nil {
		t.Error("GetEntityFromScene failed")
	}

	// 移动实体
	events, err = mgr.MoveEntityInScene(1, 1, Vector2{X: 200, Y: 200})
	if err != nil {
		t.Errorf("MoveEntityInScene failed: %v", err)
	}
	_ = events

	// 移除实体
	events, err = mgr.RemoveEntityFromScene(1, 1)
	if err != nil {
		t.Errorf("RemoveEntityFromScene failed: %v", err)
	}
	_ = events

	// 不存在的场景
	_, err = mgr.AddEntityToScene(999, entity)
	if err == nil {
		t.Error("AddEntityToScene to non-existent scene should fail")
	}
}

// TestEntityTypes 测试实体类型
func TestEntityTypes(t *testing.T) {
	tests := []struct {
		t        EntityType
		expected string
	}{
		{EntityTypePlayer, "player"},
		{EntityTypeNPC, "npc"},
		{EntityTypeMonster, "monster"},
		{EntityTypeResource, "resource"},
		{EntityTypeBuilding, "building"},
	}

	for _, test := range tests {
		if test.t.String() != test.expected {
			t.Errorf("EntityType(%d).String() = %s, expected %s",
				test.t, test.t.String(), test.expected)
		}
	}
}

// TestEntityData 测试实体扩展数据
func TestEntityData(t *testing.T) {
	entity := NewEntity(1, EntityTypePlayer, Vector2{X: 100, Y: 100})

	entity.SetData("name", "test")
	entity.SetData("level", 10)

	name, ok := entity.GetData("name")
	if !ok || name != "test" {
		t.Error("GetData failed for name")
	}

	level, ok := entity.GetData("level")
	if !ok || level != 10 {
		t.Error("GetData failed for level")
	}

	_, ok = entity.GetData("notexist")
	if ok {
		t.Error("GetData for non-existent key should return false")
	}
}

// TestStats 测试统计信息
func TestStats(t *testing.T) {
	aoi := NewAOI(1000, 1000, 50)
	stats := aoi.GetStats()

	if stats.GridCountX != 20 || stats.GridCountY != 20 {
		t.Errorf("Grid count wrong: %dx%d", stats.GridCountX, stats.GridCountY)
	}

	if stats.TotalGrids != 400 {
		t.Errorf("Total grids = %d, expected 400", stats.TotalGrids)
	}
}
