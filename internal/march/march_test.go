package march

import (
	"testing"
	"time"

	"github.com/luckinbyte/wg_ai/internal/scene"
)

// TestArmyStatus 测试军队状态
func TestArmyStatus(t *testing.T) {
	tests := []struct {
		status   ArmyStatus
		expected string
	}{
		{ArmyStatusIdle, "idle"},
		{ArmyStatusMarching, "marching"},
		{ArmyStatusCollecting, "collecting"},
		{ArmyStatusBattle, "battle"},
		{ArmyStatusStationing, "stationing"},
	}

	for _, test := range tests {
		if test.status.String() != test.expected {
			t.Errorf("ArmyStatus(%d).String() = %s, expected %s",
				test.status, test.status.String(), test.expected)
		}
	}
}

// TestMarchType 测试行军类型
func TestMarchType(t *testing.T) {
	tests := []struct {
		marchType MarchType
		expected  string
	}{
		{MarchTypeCollect, "collect"},
		{MarchTypeAttack, "attack"},
		{MarchTypeReinforce, "reinforce"},
		{MarchTypeReturn, "return"},
	}

	for _, test := range tests {
		if test.marchType.String() != test.expected {
			t.Errorf("MarchType(%d).String() = %s, expected %s",
				test.marchType, test.marchType.String(), test.expected)
		}
	}
}

// TestNewArmy 测试创建军队
func TestNewArmy(t *testing.T) {
	soldiers := map[int]int{1: 1000}
	army := NewArmy(1, 100, 200, soldiers)

	if army.ID != 1 {
		t.Errorf("Expected ID=1, got %d", army.ID)
	}
	if army.OwnerID != 100 {
		t.Errorf("Expected OwnerID=100, got %d", army.OwnerID)
	}
	if army.HeroID != 200 {
		t.Errorf("Expected HeroID=200, got %d", army.HeroID)
	}
	if army.GetTotalSoldiers() != 1000 {
		t.Errorf("Expected total soldiers=1000, got %d", army.GetTotalSoldiers())
	}
	if !army.IsIdle() {
		t.Error("New army should be idle")
	}
}

// TestArmyIsMarching 测试行军状态判断
func TestArmyIsMarching(t *testing.T) {
	army := NewArmy(1, 100, 200, map[int]int{1: 1000})

	if army.IsMarching() {
		t.Error("Idle army should not be marching")
	}

	army.Status = ArmyStatusMarching
	if !army.IsMarching() {
		t.Error("Marching army should be marching")
	}

	army.Status = ArmyStatusCollecting
	if !army.IsMarching() {
		t.Error("Collecting army should be marching")
	}
}

// TestArmyCanMarch 测试是否可以行军
func TestArmyCanMarch(t *testing.T) {
	army := NewArmy(1, 100, 200, map[int]int{1: 1000})

	if !army.CanMarch() {
		t.Error("Idle army should be able to march")
	}

	army.Status = ArmyStatusMarching
	if army.CanMarch() {
		t.Error("Marching army should not be able to march again")
	}
}

// TestArmyStartMarch 测试开始行军
func TestArmyStartMarch(t *testing.T) {
	army := NewArmy(1, 100, 200, map[int]int{1: 1000})
	army.Position = scene.Vector2{X: 0, Y: 0}

	targetPos := scene.Vector2{X: 1000, Y: 0}
	path := []scene.Vector2{army.Position, targetPos}
	speed := 100.0 // 100单位/秒

	army.StartMarch(MarchTypeCollect, 1, targetPos, path, speed)

	if !army.IsMarching() {
		t.Error("Army should be marching after StartMarch")
	}
	if army.March == nil {
		t.Fatal("March data should not be nil")
	}
	if army.March.Type != MarchTypeCollect {
		t.Errorf("Expected march type=collect, got %s", army.March.Type)
	}
	if army.March.TargetID != 1 {
		t.Errorf("Expected target_id=1, got %d", army.March.TargetID)
	}
	if army.March.Speed != speed {
		t.Errorf("Expected speed=%f, got %f", speed, army.March.Speed)
	}
}

// TestArmyFinishMarch 测试结束行军
func TestArmyFinishMarch(t *testing.T) {
	army := NewArmy(1, 100, 200, map[int]int{1: 1000})
	army.Position = scene.Vector2{X: 0, Y: 0}

	targetPos := scene.Vector2{X: 1000, Y: 0}
	path := []scene.Vector2{army.Position, targetPos}
	army.StartMarch(MarchTypeCollect, 1, targetPos, path, 100)

	army.FinishMarch()

	if !army.IsIdle() {
		t.Error("Army should be idle after FinishMarch")
	}
	if army.March != nil {
		t.Error("March data should be nil after FinishMarch")
	}
	if army.Position.X != 1000 {
		t.Errorf("Position should be at target, got %f", army.Position.X)
	}
}

// TestArmyGetProgress 测试获取行军进度
func TestArmyGetProgress(t *testing.T) {
	army := NewArmy(1, 100, 200, map[int]int{1: 1000})

	// 没有行军数据时进度应为1.0
	if army.GetProgress() != 1.0 {
		t.Errorf("Progress should be 1.0 when not marching, got %f", army.GetProgress())
	}

	// 开始行军
	army.Position = scene.Vector2{X: 0, Y: 0}
	targetPos := scene.Vector2{X: 1000, Y: 0}
	path := []scene.Vector2{army.Position, targetPos}
	army.StartMarch(MarchTypeCollect, 1, targetPos, path, 100)

	// 刚开始行军时进度应该接近0
	progress := army.GetProgress()
	if progress < 0 || progress > 1 {
		t.Errorf("Progress should be between 0 and 1, got %f", progress)
	}
}

// TestArmyPower 测试战力计算
func TestArmyCalcPower(t *testing.T) {
	army := NewArmy(1, 100, 200, map[int]int{1: 1000})
	power := army.CalcPower()

	expected := int64(1000 * 10) // 士兵数 * 10
	if power != expected {
		t.Errorf("Expected power=%d, got %d", expected, power)
	}
}

// TestArmyLoadCapacity 测试负重计算
func TestArmyCalcLoadCapacity(t *testing.T) {
	army := NewArmy(1, 100, 200, map[int]int{1: 1000})
	capacity := army.CalcLoadCapacity()

	expected := int64(1000 * 100) // 士兵数 * 100
	if capacity != expected {
		t.Errorf("Expected capacity=%d, got %d", expected, capacity)
	}
}

// TestArmyLoad 测试负重操作
func TestArmyLoad(t *testing.T) {
	army := NewArmy(1, 100, 200, map[int]int{1: 100})

	// 添加负重
	army.AddLoad(scene.ResourceFood, 1000)
	army.AddLoad(scene.ResourceWood, 500)

	if army.LoadFood != 1000 {
		t.Errorf("Expected LoadFood=1000, got %d", army.LoadFood)
	}
	if army.LoadWood != 500 {
		t.Errorf("Expected LoadWood=500, got %d", army.LoadWood)
	}

	// 获取当前负重
	currentLoad := army.GetCurrentLoad()
	if currentLoad != 1500 {
		t.Errorf("Expected currentLoad=1500, got %d", currentLoad)
	}

	// 检查能否装载更多
	if !army.CanLoadMore(5000) {
		t.Error("Should be able to load 5000 more")
	}
	if army.CanLoadMore(50000) {
		t.Error("Should not be able to load 50000 more")
	}

	// 清空负重
	army.ClearLoad()
	if army.GetCurrentLoad() != 0 {
		t.Errorf("Expected currentLoad=0 after clear, got %d", army.GetCurrentLoad())
	}
}

// TestArmyCollect 测试采集
func TestArmyCollect(t *testing.T) {
	army := NewArmy(1, 100, 200, map[int]int{1: 1000})

	// 先设置行军数据
	army.Position = scene.Vector2{X: 0, Y: 0}
	targetPos := scene.Vector2{X: 1000, Y: 0}
	path := []scene.Vector2{army.Position, targetPos}
	army.StartMarch(MarchTypeCollect, 1, targetPos, path, 100)

	// 开始采集
	army.StartCollect(30 * time.Second)

	if !army.IsCollecting() {
		t.Error("Army should be collecting")
	}
	if army.Status != ArmyStatusCollecting {
		t.Error("Status should be ArmyStatusCollecting")
	}

	// 完成采集
	army.FinishCollect()
	if !army.IsIdle() {
		t.Error("Army should be idle after finish collect")
	}
}

// TestSqrt 测试平方根
func TestSqrt(t *testing.T) {
	tests := []struct {
		x        float64
		expected float64
	}{
		{0, 0},
		{1, 1},
		{4, 2},
		{9, 3},
		{100, 10},
	}

	for _, test := range tests {
		result := sqrt(test.x)
		// 允许小误差
		if result < test.expected-0.0001 || result > test.expected+0.0001 {
			t.Errorf("sqrt(%f) = %f, expected %f", test.x, result, test.expected)
		}
	}
}

// TestManagerCreateArmy 测试创建军队
func TestManagerCreateArmy(t *testing.T) {
	// 创建场景管理器
	sceneMgr := scene.NewManager()
	sceneMgr.CreateScene(scene.SceneConfig{
		ID:       1,
		Width:    1000,
		Height:   1000,
		GridSize: 50,
	})

	// 创建行军管理器
	mgr := NewManager(sceneMgr)

	// 创建军队
	army, err := mgr.CreateArmy(100, 200, map[int]int{1: 1000}, scene.Vector2{X: 100, Y: 100}, 1)
	if err != nil {
		t.Fatalf("Failed to create army: %v", err)
	}

	if army.ID == 0 {
		t.Error("Army ID should not be 0")
	}
	if army.OwnerID != 100 {
		t.Errorf("Expected OwnerID=100, got %d", army.OwnerID)
	}

	// 获取军队
	gotArmy := mgr.GetArmy(army.ID)
	if gotArmy == nil {
		t.Error("GetArmy should return the army")
	}
	if gotArmy.ID != army.ID {
		t.Errorf("GetArmy returned wrong army")
	}

	// 获取玩家军队
	armies := mgr.GetPlayerArmies(100)
	if len(armies) != 1 {
		t.Errorf("Expected 1 army, got %d", len(armies))
	}

	// 测试军队数量上限
	for i := 0; i < 5; i++ {
		_, err = mgr.CreateArmy(100, 200, map[int]int{1: 1000}, scene.Vector2{X: 100, Y: 100}, 1)
		if err != nil {
			break
		}
	}
	// 第6支军队应该失败
	_, err = mgr.CreateArmy(100, 200, map[int]int{1: 1000}, scene.Vector2{X: 100, Y: 100}, 1)
	if err == nil {
		t.Error("Should fail to create 6th army")
	}
}

// TestManagerStats 测试统计信息
func TestManagerStats(t *testing.T) {
	sceneMgr := scene.NewManager()
	sceneMgr.CreateScene(scene.SceneConfig{
		ID:       1,
		Width:    1000,
		Height:   1000,
		GridSize: 50,
	})

	mgr := NewManager(sceneMgr)

	// 创建一些军队
	mgr.CreateArmy(100, 200, map[int]int{1: 1000}, scene.Vector2{X: 100, Y: 100}, 1)
	mgr.CreateArmy(101, 200, map[int]int{1: 1000}, scene.Vector2{X: 100, Y: 100}, 1)

	stats := mgr.GetStats()
	if stats["army_count"].(int) != 2 {
		t.Errorf("Expected army_count=2, got %d", stats["army_count"])
	}
	if stats["player_count"].(int) != 2 {
		t.Errorf("Expected player_count=2, got %d", stats["player_count"])
	}
}
