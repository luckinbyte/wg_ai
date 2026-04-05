package scene

import (
	"math"
)

// AOI 九宫格兴趣区域管理器
type AOI struct {
	width    int        // 格子数量(横向)
	height   int        // 格子数量(纵向)
	gridSize int        // 格子大小
	grids    [][]*Grid  // 二维格子数组
}

// NewAOI 创建AOI管理器
// sceneWidth, sceneHeight: 场景尺寸
// gridSize: 格子大小
func NewAOI(sceneWidth, sceneHeight, gridSize int) *AOI {
	// 计算格子数量
	gridCountX := (sceneWidth + gridSize - 1) / gridSize
	gridCountY := (sceneHeight + gridSize - 1) / gridSize

	aoi := &AOI{
		width:    gridCountX,
		height:   gridCountY,
		gridSize: gridSize,
		grids:    make([][]*Grid, gridCountX),
	}

	// 初始化所有格子
	for x := 0; x < gridCountX; x++ {
		aoi.grids[x] = make([]*Grid, gridCountY)
		for y := 0; y < gridCountY; y++ {
			aoi.grids[x][y] = NewGrid(x, y)
		}
	}

	return aoi
}

// posToGrid 坐标转格子索引
func (a *AOI) posToGrid(pos Vector2) (gx, gy int) {
	gx = int(pos.X) / a.gridSize
	gy = int(pos.Y) / a.gridSize

	// 边界检查
	if gx < 0 {
		gx = 0
	}
	if gx >= a.width {
		gx = a.width - 1
	}
	if gy < 0 {
		gy = 0
	}
	if gy >= a.height {
		gy = a.height - 1
	}

	return
}

// isValidGrid 检查格子坐标是否有效
func (a *AOI) isValidGrid(gx, gy int) bool {
	return gx >= 0 && gx < a.width && gy >= 0 && gy < a.height
}

// GetGrid 获取指定位置的格子
func (a *AOI) GetGrid(pos Vector2) *Grid {
	gx, gy := a.posToGrid(pos)
	return a.grids[gx][gy]
}

// GetGridByIndex 通过索引获取格子
func (a *AOI) GetGridByIndex(gx, gy int) *Grid {
	if !a.isValidGrid(gx, gy) {
		return nil
	}
	return a.grids[gx][gy]
}

// GetNeighbors 获取九宫格内的所有格子 (包括自身)
func (a *AOI) GetNeighbors(gx, gy int) []*Grid {
	grids := make([]*Grid, 0, 9)

	// 遍历九宫格
	for dx := -1; dx <= 1; dx++ {
		for dy := -1; dy <= 1; dy++ {
			nx, ny := gx+dx, gy+dy
			if a.isValidGrid(nx, ny) {
				grids = append(grids, a.grids[nx][ny])
			}
		}
	}

	return grids
}

// GetNeighborsByPos 通过坐标获取九宫格格子
func (a *AOI) GetNeighborsByPos(pos Vector2) []*Grid {
	gx, gy := a.posToGrid(pos)
	return a.GetNeighbors(gx, gy)
}

// Enter 实体进入AOI，返回需要通知的观察者事件
func (a *AOI) Enter(entity *Entity) []AOIEvent {
	events := make([]AOIEvent, 0)

	// 获取实体所在格子
	grid := a.GetGrid(entity.Position)
	grid.AddEntity(entity)

	// 获取九宫格内的所有实体
	neighbors := a.GetNeighborsByPos(entity.Position)
	for _, g := range neighbors {
		for _, other := range g.Entities {
			if other.ID == entity.ID {
				continue
			}
			// 新实体进入其他实体的视野
			events = append(events, AOIEvent{
				Type:    EventEnter,
				Entity:  entity,
				Watcher: other.ID,
			})
			// 其他实体进入新实体的视野 (如果是观察者)
			if other.Type == EntityTypePlayer {
				events = append(events, AOIEvent{
					Type:    EventEnter,
					Entity:  other,
					Watcher: entity.ID,
				})
			}
		}
	}

	return events
}

// Leave 实体离开AOI，返回需要通知的观察者事件
func (a *AOI) Leave(entity *Entity) []AOIEvent {
	events := make([]AOIEvent, 0)

	// 获取实体所在格子
	grid := a.GetGrid(entity.Position)

	// 获取九宫格内的所有实体
	neighbors := a.GetNeighborsByPos(entity.Position)
	for _, g := range neighbors {
		for _, other := range g.Entities {
			if other.ID == entity.ID {
				continue
			}
			// 实体离开其他实体的视野
			events = append(events, AOIEvent{
				Type:    EventLeave,
				Entity:  entity,
				Watcher: other.ID,
			})
			// 其他实体离开实体的视野 (如果是观察者)
			if other.Type == EntityTypePlayer {
				events = append(events, AOIEvent{
					Type:    EventLeave,
					Entity:  other,
					Watcher: entity.ID,
				})
			}
		}
	}

	// 从格子中移除
	grid.RemoveEntity(entity.ID)

	return events
}

// Move 实体移动，返回视野变化事件
func (a *AOI) Move(entity *Entity, newPos Vector2) []AOIEvent {
	events := make([]AOIEvent, 0)

	oldGx, oldGy := a.posToGrid(entity.Position)
	newGx, newGy := a.posToGrid(newPos)

	// 更新实体位置
	entity.Position = newPos

	// 如果格子没变，不需要处理视野变化
	if oldGx == newGx && oldGy == newGy {
		return events
	}

	// 获取旧九宫格和新九宫格
	oldNeighbors := a.GetNeighbors(oldGx, oldGy)
	newNeighbors := a.GetNeighbors(newGx, newGy)

	// 构建旧格子集合
	oldGridSet := make(map[[2]int]bool)
	for _, g := range oldNeighbors {
		oldGridSet[[2]int{g.X, g.Y}] = true
	}

	// 构建新格子集合
	newGridSet := make(map[[2]int]bool)
	for _, g := range newNeighbors {
		newGridSet[[2]int{g.X, g.Y}] = true
	}

	// 从旧格子移除，添加到新格子
	oldGrid := a.GetGridByIndex(oldGx, oldGy)
	newGrid := a.GetGridByIndex(newGx, newGy)

	if oldGrid != nil {
		oldGrid.RemoveEntity(entity.ID)
	}
	if newGrid != nil {
		newGrid.AddEntity(entity)
	}

	// 处理离开视野的实体 (在旧九宫格但不在新九宫格)
	for _, g := range oldNeighbors {
		if newGridSet[[2]int{g.X, g.Y}] {
			continue // 还在新视野内
		}
		for _, other := range g.Entities {
			if other.ID == entity.ID {
				continue
			}
			// 实体离开视野
			events = append(events, AOIEvent{
				Type:    EventLeave,
				Entity:  other,
				Watcher: entity.ID,
			})
			// 离开其他实体的视野
			events = append(events, AOIEvent{
				Type:    EventLeave,
				Entity:  entity,
				Watcher: other.ID,
			})
		}
	}

	// 处理进入视野的实体 (在新九宫格但不在旧九宫格)
	for _, g := range newNeighbors {
		if oldGridSet[[2]int{g.X, g.Y}] {
			continue // 已经在视野内
		}
		for _, other := range g.Entities {
			if other.ID == entity.ID {
				continue
			}
			// 实体进入视野
			events = append(events, AOIEvent{
				Type:    EventEnter,
				Entity:  other,
				Watcher: entity.ID,
			})
			// 进入其他实体的视野
			events = append(events, AOIEvent{
				Type:    EventEnter,
				Entity:  entity,
				Watcher: other.ID,
			})
		}
	}

	return events
}

// GetEntitiesInAOI 获取指定实体视野内的所有其他实体
func (a *AOI) GetEntitiesInAOI(entity *Entity) []*Entity {
	entities := make([]*Entity, 0)

	neighbors := a.GetNeighborsByPos(entity.Position)
	for _, g := range neighbors {
		for _, other := range g.Entities {
			if other.ID != entity.ID {
				entities = append(entities, other)
			}
		}
	}

	return entities
}

// GetEntitiesInAOIByPos 获取指定位置视野内的所有实体
func (a *AOI) GetEntitiesInAOIByPos(pos Vector2, excludeID int64) []*Entity {
	entities := make([]*Entity, 0)

	neighbors := a.GetNeighborsByPos(pos)
	for _, g := range neighbors {
		for _, other := range g.Entities {
			if other.ID != excludeID {
				entities = append(entities, other)
			}
		}
	}

	return entities
}

// GridInfo 格子信息 (用于调试)
type GridInfo struct {
	X          int      `json:"x"`
	Y          int      `json:"y"`
	EntityIDs  []int64  `json:"entity_ids"`
}

// GetGridInfo 获取格子信息 (用于调试)
func (a *AOI) GetGridInfo(gx, gy int) *GridInfo {
	grid := a.GetGridByIndex(gx, gy)
	if grid == nil {
		return nil
	}

	ids := make([]int64, 0, len(grid.Entities))
	for id := range grid.Entities {
		ids = append(ids, id)
	}

	return &GridInfo{
		X:         grid.X,
		Y:         grid.Y,
		EntityIDs: ids,
	}
}

// Stats 统计信息
type Stats struct {
	GridCountX   int `json:"grid_count_x"`
	GridCountY   int `json:"grid_count_y"`
	GridSize     int `json:"grid_size"`
	TotalGrids   int `json:"total_grids"`
	TotalEntities int `json:"total_entities"`
}

// GetStats 获取统计信息
func (a *AOI) GetStats() Stats {
	totalEntities := 0
	for x := 0; x < a.width; x++ {
		for y := 0; y < a.height; y++ {
			totalEntities += a.grids[x][y].EntityCount()
		}
	}

	return Stats{
		GridCountX:    a.width,
		GridCountY:    a.height,
		GridSize:      a.gridSize,
		TotalGrids:    a.width * a.height,
		TotalEntities: totalEntities,
	}
}

// Distance 计算两个位置的距离
func Distance(p1, p2 Vector2) float64 {
	dx := p1.X - p2.X
	dy := p1.Y - p2.Y
	return math.Sqrt(dx*dx + dy*dy)
}

// DistanceSquared 计算距离的平方 (避免开方，用于比较)
func DistanceSquared(p1, p2 Vector2) float64 {
	dx := p1.X - p2.X
	dy := p1.Y - p2.Y
	return dx*dx + dy*dy
}
