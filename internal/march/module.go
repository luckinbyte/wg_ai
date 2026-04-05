package march

import (
	"encoding/json"
	"fmt"
	"time"

	baseplugin "github.com/yourorg/wg_ai/plugin"
	"github.com/yourorg/wg_ai/internal/scene"
)

// Module 行军模块 (实现 LogicModule 接口)
type Module struct {
	mgr *Manager
}

// NewModule 创建行军模块
func NewModule(mgr *Manager) *Module {
	return &Module{mgr: mgr}
}

// Name 返回模块名称
func (m *Module) Name() string {
	return "march"
}

// Handle 处理请求
func (m *Module) Handle(ctx *baseplugin.LogicContext, method string, params map[string]any) (*baseplugin.LogicResult, error) {
	if m.mgr == nil {
		return baseplugin.Error(500, "march manager not initialized"), nil
	}

	switch method {
	case "create_army":
		return m.handleCreateArmy(ctx, params)
	case "delete_army":
		return m.handleDeleteArmy(ctx, params)
	case "get_armies":
		return m.handleGetArmies(ctx, params)
	case "start_march":
		return m.handleStartMarch(ctx, params)
	case "cancel_march":
		return m.handleCancelMarch(ctx, params)
	case "get_march_info":
		return m.handleGetMarchInfo(ctx, params)
	default:
		return nil, baseplugin.ErrMethodNotFound
	}
}

// handleCreateArmy 创建军队
func (m *Module) handleCreateArmy(ctx *baseplugin.LogicContext, params map[string]any) (*baseplugin.LogicResult, error) {
	// 获取参数
	heroID := int64(params["hero_id"].(float64))
	sceneID := int64(params["scene_id"].(float64))
	posX := params["x"].(float64)
	posY := params["y"].(float64)

	// 解析士兵参数
	soldiers := make(map[int]int)
	if soldiersRaw, ok := params["soldiers"]; ok {
		switch v := soldiersRaw.(type) {
		case map[string]any:
			for k, val := range v {
				var soldierID int
				fmt.Sscanf(k, "%d", &soldierID)
				if count, ok := val.(float64); ok {
					soldiers[soldierID] = int(count)
				}
			}
		case map[int]int:
			soldiers = v
		}
	}

	// 检查士兵数量
	totalSoldiers := 0
	for _, count := range soldiers {
		totalSoldiers += count
	}
	if totalSoldiers == 0 {
		return baseplugin.Error(400, "no soldiers specified"), nil
	}

	// 创建军队 (使用带数据访问的方法)
	army, err := m.mgr.CreateArmyWithData(ctx.Data, ctx.RID, heroID, soldiers, scene.Vector2{X: posX, Y: posY}, sceneID)
	if err != nil {
		return baseplugin.Error(400, err.Error()), nil
	}

	return baseplugin.Success(map[string]any{
		"army_id":  army.ID,
		"owner_id": army.OwnerID,
		"hero_id":  army.HeroID,
		"soldiers": army.Soldiers,
		"status":   army.Status.String(),
	}), nil
}

// handleDeleteArmy 解散军队
func (m *Module) handleDeleteArmy(ctx *baseplugin.LogicContext, params map[string]any) (*baseplugin.LogicResult, error) {
	armyID := int64(params["army_id"].(float64))

	// 验证军队归属
	army := m.mgr.GetArmy(armyID)
	if army == nil {
		return baseplugin.Error(404, "army not found"), nil
	}
	if army.OwnerID != ctx.RID {
		return baseplugin.Error(403, "not your army"), nil
	}

	// 删除军队并归还士兵
	err := m.mgr.DeleteArmyWithData(ctx.Data, armyID)
	if err != nil {
		return baseplugin.Error(400, err.Error()), nil
	}

	return baseplugin.Success(map[string]any{
		"army_id": armyID,
		"message": "army disbanded, soldiers returned",
	}), nil
}

// handleGetArmies 获取军队列表
func (m *Module) handleGetArmies(ctx *baseplugin.LogicContext, params map[string]any) (*baseplugin.LogicResult, error) {
	armies := m.mgr.GetPlayerArmies(ctx.RID)

	armyList := make([]map[string]any, 0, len(armies))
	for _, army := range armies {
		armyList = append(armyList, m.serializeArmy(army))
	}

	return baseplugin.Success(map[string]any{
		"armies": armyList,
		"count":  len(armyList),
	}), nil
}

// handleStartMarch 开始行军
func (m *Module) handleStartMarch(ctx *baseplugin.LogicContext, params map[string]any) (*baseplugin.LogicResult, error) {
	armyID := int64(params["army_id"].(float64))
	marchType := MarchType(int(params["march_type"].(float64)))
	targetID := int64(params["target_id"].(float64))

	// 验证军队归属
	army := m.mgr.GetArmy(armyID)
	if army == nil {
		return baseplugin.Error(404, "army not found"), nil
	}
	if army.OwnerID != ctx.RID {
		return baseplugin.Error(403, "not your army"), nil
	}

	// 开始行军
	err := m.mgr.StartMarch(armyID, marchType, targetID)
	if err != nil {
		return baseplugin.Error(400, err.Error()), nil
	}

	// 返回行军信息
	army = m.mgr.GetArmy(armyID)

	return baseplugin.Success(map[string]any{
		"army_id":      armyID,
		"march_type":   marchType.String(),
		"target_id":    targetID,
		"arrival_time": army.March.ArrivalTime,
		"progress":     army.GetProgress(),
	}), nil
}

// handleCancelMarch 取消行军
func (m *Module) handleCancelMarch(ctx *baseplugin.LogicContext, params map[string]any) (*baseplugin.LogicResult, error) {
	armyID := int64(params["army_id"].(float64))

	// 验证军队归属
	army := m.mgr.GetArmy(armyID)
	if army == nil {
		return baseplugin.Error(404, "army not found"), nil
	}
	if army.OwnerID != ctx.RID {
		return baseplugin.Error(403, "not your army"), nil
	}

	// 取消行军
	err := m.mgr.CancelMarch(armyID)
	if err != nil {
		return baseplugin.Error(400, err.Error()), nil
	}

	return baseplugin.Success(map[string]any{
		"army_id": armyID,
		"message": "march cancelled, returning to base",
	}), nil
}

// handleGetMarchInfo 获取行军信息
func (m *Module) handleGetMarchInfo(ctx *baseplugin.LogicContext, params map[string]any) (*baseplugin.LogicResult, error) {
	armyID := int64(params["army_id"].(float64))

	army := m.mgr.GetArmy(armyID)
	if army == nil {
		return baseplugin.Error(404, "army not found"), nil
	}

	return baseplugin.Success(m.serializeArmy(army)), nil
}

// serializeArmy 序列化军队信息
func (m *Module) serializeArmy(army *Army) map[string]any {
	data := map[string]any{
		"id":       army.ID,
		"owner_id": army.OwnerID,
		"hero_id":  army.HeroID,
		"soldiers": army.Soldiers,
		"status":   army.Status.String(),
		"position": map[string]any{
			"x": army.Position.X,
			"y": army.Position.Y,
		},
		"scene_id":     army.SceneID,
		"power":        army.CalcPower(),
		"load_capacity": army.CalcLoadCapacity(),
		"current_load":  army.GetCurrentLoad(),
	}

	// 行军信息
	if army.March != nil {
		marchData := map[string]any{
			"type":         army.March.Type.String(),
			"target_id":    army.March.TargetID,
			"target_pos":   army.March.TargetPos,
			"start_time":   army.March.StartTime,
			"arrival_time": army.March.ArrivalTime,
			"speed":        army.March.Speed,
			"progress":     army.GetProgress(),
		}
		if army.March.CollectEndTime > 0 {
			marchData["collect_end_time"] = army.March.CollectEndTime
		}
		data["march"] = marchData
	}

	// 携带资源
	if army.GetCurrentLoad() > 0 {
		data["load"] = map[string]any{
			"food":  army.LoadFood,
			"wood":  army.LoadWood,
			"stone": army.LoadStone,
			"gold":  army.LoadGold,
		}
	}

	return data
}

// serializeMarchEvent 序列化行军事件 (用于推送)
func (m *Module) serializeMarchEvent(eventType string, army *Army) []byte {
	data := map[string]any{
		"event":     eventType,
		"army":      m.serializeArmy(army),
		"timestamp": fmt.Sprintf("%d", time.Now().UnixMilli()),
	}
	bytes, _ := json.Marshal(data)
	return bytes
}
