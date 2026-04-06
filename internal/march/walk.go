package march

import (
	"sync"
	"time"

	"github.com/luckinbyte/wg_ai/internal/scene"
)

// WalkSimulator 移动模拟器
type WalkSimulator struct {
	mgr      *Manager
	interval time.Duration
	stopCh   chan struct{}
	running  bool
	mutex    sync.RWMutex
}

// NewWalkSimulator 创建移动模拟器
func NewWalkSimulator(mgr *Manager) *WalkSimulator {
	return &WalkSimulator{
		mgr:      mgr,
		interval: 100 * time.Millisecond, // 100ms更新一次
		stopCh:   make(chan struct{}),
	}
}

// Start 启动模拟器
func (w *WalkSimulator) Start() {
	w.mutex.Lock()
	defer w.mutex.Unlock()

	if w.running {
		return
	}

	w.running = true
	go w.tickLoop()
}

// Stop 停止模拟器
func (w *WalkSimulator) Stop() {
	w.mutex.Lock()
	defer w.mutex.Unlock()

	if !w.running {
		return
	}

	w.running = false
	close(w.stopCh)
}

// AddArmy 添加军队到模拟器
func (w *WalkSimulator) AddArmy(army *Army) {
	// 军队已经在marchingArmies中管理
}

// RemoveArmy 从模拟器移除军队
func (w *WalkSimulator) RemoveArmy(armyID int64) {
	// 军队已经在marchingArmies中管理
}

// tickLoop 主循环
func (w *WalkSimulator) tickLoop() {
	ticker := time.NewTicker(w.interval)
	defer ticker.Stop()

	for {
		select {
		case <-w.stopCh:
			return
		case <-ticker.C:
			w.tick()
		}
	}
}

// tick 每帧更新
func (w *WalkSimulator) tick() {
	armies := w.mgr.GetMarchingArmies()

	for _, army := range armies {
		if army.March == nil {
			continue
		}

		w.updateArmyPosition(army)

		// 检查是否到达
		if w.checkArrival(army) {
			w.mgr.OnArmyArrival(army)
		}
	}
}

// updateArmyPosition 更新军队位置
func (w *WalkSimulator) updateArmyPosition(army *Army) {
	if army.March == nil || len(army.March.Path) < 2 {
		return
	}

	// 计算进度
	progress := army.GetProgress()

	// 插值计算当前位置
	start := army.March.Path[0]
	end := army.March.Path[len(army.March.Path)-1]

	army.Position = scene.Vector2{
		X: start.X + (end.X-start.X)*progress,
		Y: start.Y + (end.Y-start.Y)*progress,
	}
}

// checkArrival 检查是否到达
func (w *WalkSimulator) checkArrival(army *Army) bool {
	if army.March == nil {
		return false
	}

	now := time.Now().UnixMilli()
	return now >= army.March.ArrivalTime
}
