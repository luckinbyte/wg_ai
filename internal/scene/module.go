package scene

import (
	"encoding/json"

	baseplugin "github.com/yourorg/wg_ai/plugin"
)

// Module 场景模块 (实现 LogicModule 接口)
type Module struct {
	mgr *Manager
}

// NewModule 创建场景模块
func NewModule(mgr *Manager) *Module {
	return &Module{mgr: mgr}
}

// Name 返回模块名称
func (m *Module) Name() string {
	return "scene"
}

// Handle 处理场景请求
func (m *Module) Handle(ctx *baseplugin.LogicContext, method string, params map[string]any) (*baseplugin.LogicResult, error) {
	if m.mgr == nil {
		return baseplugin.Error(500, "scene manager not initialized"), nil
	}

	switch method {
	case "enter":
		return m.handleEnter(ctx, params)
	case "move":
		return m.handleMove(ctx, params)
	case "leave":
		return m.handleLeave(ctx, params)
	case "get_nearby":
		return m.handleGetNearby(ctx, params)
	case "get_scene_info":
		return m.handleGetSceneInfo(ctx, params)
	default:
		return nil, baseplugin.ErrMethodNotFound
	}
}

// handleEnter 进入场景
func (m *Module) handleEnter(ctx *baseplugin.LogicContext, params map[string]any) (*baseplugin.LogicResult, error) {
	// 获取参数
	sceneID, ok := params["scene_id"].(float64)
	if !ok {
		return baseplugin.Error(400, "scene_id required"), nil
	}
	x, _ := params["x"].(float64)
	y, _ := params["y"].(float64)

	// 获取或创建场景
	s := m.mgr.GetScene(int64(sceneID))
	if s == nil {
		// 创建默认场景 1000x1000, 格子大小50
		s = m.mgr.CreateScene(SceneConfig{
			ID:       int64(sceneID),
			Width:    1000,
			Height:   1000,
			GridSize: 50,
		})
	}

	// 创建玩家实体
	entity := CreatePlayerEntity(ctx.RID, Vector2{X: x, Y: y}, int64(sceneID))

	// 从玩家数据获取扩展信息
	if name, err := ctx.Data.GetField("name"); err == nil {
		entity.SetData("name", name)
	}
	if level, err := ctx.Data.GetField("level"); err == nil {
		entity.SetData("level", level)
	}

	// 加入场景
	events := s.AddEntity(entity)

	// 保存玩家当前位置
	ctx.Data.SetField("scene_id", int64(sceneID))
	ctx.Data.SetField("pos_x", x)
	ctx.Data.SetField("pos_y", y)
	ctx.Data.MarkDirty()

	// 构建响应
	result := map[string]any{
		"rid":      ctx.RID,
		"scene_id": int64(sceneID),
		"x":        x,
		"y":        y,
	}

	// 处理进入视野事件
	pushList := make([]baseplugin.PushData, 0)
	for _, event := range events {
		if event.Watcher == ctx.RID && event.Type == EventEnter {
			pushData, _ := json.Marshal(map[string]any{
				"event":  "enter",
				"entity": serializeEntity(event.Entity),
			})
			pushList = append(pushList, baseplugin.PushData{
				MsgID: 3010,
				Data:  pushData,
			})
		}
	}

	return &baseplugin.LogicResult{
		Code: 0,
		Data: result,
		Push: pushList,
	}, nil
}

// handleMove 移动
func (m *Module) handleMove(ctx *baseplugin.LogicContext, params map[string]any) (*baseplugin.LogicResult, error) {
	sceneID, ok := params["scene_id"].(float64)
	if !ok {
		return baseplugin.Error(400, "scene_id required"), nil
	}
	x, ok := params["x"].(float64)
	if !ok {
		return baseplugin.Error(400, "x required"), nil
	}
	y, ok := params["y"].(float64)
	if !ok {
		return baseplugin.Error(400, "y required"), nil
	}

	s := m.mgr.GetScene(int64(sceneID))
	if s == nil {
		return baseplugin.Error(404, "scene not found"), nil
	}

	// 移动实体
	events, err := s.MoveEntity(ctx.RID, Vector2{X: x, Y: y})
	if err != nil {
		return baseplugin.Error(400, err.Error()), nil
	}

	// 更新玩家位置
	ctx.Data.SetField("pos_x", x)
	ctx.Data.SetField("pos_y", y)
	ctx.Data.MarkDirty()

	// 构建响应
	result := map[string]any{
		"rid":      ctx.RID,
		"scene_id": int64(sceneID),
		"x":        x,
		"y":        y,
	}

	// 处理视野变化事件
	pushList := make([]baseplugin.PushData, 0)
	for _, event := range events {
		if event.Watcher == ctx.RID {
			eventType := "enter"
			if event.Type == EventLeave {
				eventType = "leave"
			}
			pushData, _ := json.Marshal(map[string]any{
				"event":  eventType,
				"entity": serializeEntity(event.Entity),
			})
			pushList = append(pushList, baseplugin.PushData{
				MsgID: 3010,
				Data:  pushData,
			})
		}
	}

	return &baseplugin.LogicResult{
		Code: 0,
		Data: result,
		Push: pushList,
	}, nil
}

// handleLeave 离开场景
func (m *Module) handleLeave(ctx *baseplugin.LogicContext, params map[string]any) (*baseplugin.LogicResult, error) {
	sceneID, ok := params["scene_id"].(float64)
	if !ok {
		return baseplugin.Error(400, "scene_id required"), nil
	}

	s := m.mgr.GetScene(int64(sceneID))
	if s == nil {
		return baseplugin.Error(404, "scene not found"), nil
	}

	// 移除实体
	events := s.RemoveEntity(ctx.RID)

	// 清除玩家位置数据
	ctx.Data.SetField("scene_id", int64(0))
	ctx.Data.SetField("pos_x", float64(0))
	ctx.Data.SetField("pos_y", float64(0))
	ctx.Data.MarkDirty()

	// 处理离开视野事件
	pushList := make([]baseplugin.PushData, 0)
	for _, event := range events {
		if event.Watcher == ctx.RID && event.Type == EventLeave {
			pushData, _ := json.Marshal(map[string]any{
				"event":  "leave",
				"entity": serializeEntity(event.Entity),
			})
			pushList = append(pushList, baseplugin.PushData{
				MsgID: 3010,
				Data:  pushData,
			})
		}
	}

	return &baseplugin.LogicResult{
		Code: 0,
		Data: map[string]any{
			"rid":      ctx.RID,
			"scene_id": int64(sceneID),
		},
		Push: pushList,
	}, nil
}

// handleGetNearby 获取附近实体
func (m *Module) handleGetNearby(ctx *baseplugin.LogicContext, params map[string]any) (*baseplugin.LogicResult, error) {
	sceneID, ok := params["scene_id"].(float64)
	if !ok {
		return baseplugin.Error(400, "scene_id required"), nil
	}

	s := m.mgr.GetScene(int64(sceneID))
	if s == nil {
		return baseplugin.Error(404, "scene not found"), nil
	}

	// 获取视野内实体
	entities := s.GetEntitiesInAOI(ctx.RID)

	// 序列化实体列表
	nearbyList := make([]map[string]any, 0, len(entities))
	for _, e := range entities {
		nearbyList = append(nearbyList, serializeEntity(e))
	}

	return baseplugin.Success(map[string]any{
		"nearby": nearbyList,
		"count":  len(nearbyList),
	}), nil
}

// handleGetSceneInfo 获取场景信息
func (m *Module) handleGetSceneInfo(ctx *baseplugin.LogicContext, params map[string]any) (*baseplugin.LogicResult, error) {
	sceneID, ok := params["scene_id"].(float64)
	if !ok {
		return baseplugin.Error(400, "scene_id required"), nil
	}

	s := m.mgr.GetScene(int64(sceneID))
	if s == nil {
		return baseplugin.Error(404, "scene not found"), nil
	}

	info := s.GetInfo()
	return baseplugin.Success(map[string]any{
		"scene_id":     info.ID,
		"width":        info.Width,
		"height":       info.Height,
		"grid_size":    info.GridSize,
		"entity_count": info.EntityCount,
	}), nil
}

// serializeEntity 序列化实体为map
func serializeEntity(e *Entity) map[string]any {
	data := map[string]any{
		"id":       e.ID,
		"type":     e.Type.String(),
		"x":        e.Position.X,
		"y":        e.Position.Y,
		"scene_id": e.SceneID,
	}
	// 合并扩展数据
	for k, v := range e.Data {
		data[k] = v
	}
	return data
}
