package soldier

import (
	"fmt"

	baseplugin "github.com/luckinbyte/wg_ai/plugin"
)

// SoldierLogic 士兵逻辑模块
type SoldierLogic struct {
	mgr *Manager
}

// NewSoldierLogic 创建士兵逻辑模块
func NewSoldierLogic(mgr *Manager) *SoldierLogic {
	return &SoldierLogic{mgr: mgr}
}

// Name 返回模块名称
func (l *SoldierLogic) Name() string {
	return "soldier"
}

// Handle 处理请求
func (l *SoldierLogic) Handle(ctx *baseplugin.LogicContext, method string, params map[string]any) (*baseplugin.LogicResult, error) {
	if l.mgr == nil {
		return baseplugin.Error(500, "soldier manager not initialized"), nil
	}

	switch method {
	case "list":
		return l.handleList(ctx, params)
	case "get":
		return l.handleGet(ctx, params)
	case "train":
		return l.handleTrain(ctx, params)
	case "cancel_train":
		return l.handleCancelTrain(ctx, params)
	case "train_queue":
		return l.handleTrainQueue(ctx, params)
	case "complete_train":
		return l.handleCompleteTrain(ctx, params)
	case "heal":
		return l.handleHeal(ctx, params)
	case "cancel_heal":
		return l.handleCancelHeal(ctx, params)
	case "heal_queue":
		return l.handleHealQueue(ctx, params)
	case "complete_heal":
		return l.handleCompleteHeal(ctx, params)
	case "dismiss":
		return l.handleDismiss(ctx, params)
	case "configs":
		return l.handleConfigs(ctx, params)
	case "stats":
		return l.handleStats(ctx, params)
	default:
		return nil, baseplugin.ErrMethodNotFound
	}
}

// handleList 获取所有士兵列表
func (l *SoldierLogic) handleList(ctx *baseplugin.LogicContext, params map[string]any) (*baseplugin.LogicResult, error) {
	soldiers, err := l.mgr.GetSoldiers(ctx.Data)
	if err != nil {
		return baseplugin.Error(500, err.Error()), nil
	}

	// 转换为列表格式
	list := make([]map[string]any, 0, len(soldiers))
	for _, s := range soldiers {
		list = append(list, map[string]any{
			"id":      s.ID,
			"type":    s.Type,
			"level":   s.Level,
			"count":   s.Count,
			"wounded": s.Wounded,
		})
	}

	return baseplugin.Success(map[string]any{
		"soldiers": list,
		"total":    len(list),
	}), nil
}

// handleGet 获取指定士兵信息
func (l *SoldierLogic) handleGet(ctx *baseplugin.LogicContext, params map[string]any) (*baseplugin.LogicResult, error) {
	soldierID := int(params["soldier_id"].(float64))

	soldiers, err := l.mgr.GetSoldiers(ctx.Data)
	if err != nil {
		return baseplugin.Error(500, err.Error()), nil
	}

	s, ok := soldiers[soldierID]
	if !ok {
		return baseplugin.Error(404, "soldier not found"), nil
	}

	cfg := GetSoldierConfig(soldierID)
	result := map[string]any{
		"id":      s.ID,
		"type":    s.Type,
		"level":   s.Level,
		"count":   s.Count,
		"wounded": s.Wounded,
	}

	if cfg != nil {
		result["config"] = map[string]any{
			"name":    cfg.Name,
			"attack":  cfg.Attack,
			"defense": cfg.Defense,
			"hp":      cfg.HP,
			"power":   cfg.Power,
		}
	}

	return baseplugin.Success(result), nil
}

// handleTrain 开始训练
func (l *SoldierLogic) handleTrain(ctx *baseplugin.LogicContext, params map[string]any) (*baseplugin.LogicResult, error) {
	soldierType := int(params["type"].(float64))
	level := int(params["level"].(float64))
	count := int(params["count"].(float64))
	isUpgrade := false
	if v, ok := params["is_upgrade"]; ok {
		isUpgrade = v.(bool)
	}

	// 参数校验
	if soldierType < 1 || soldierType > 4 {
		return baseplugin.Error(400, "invalid soldier type"), nil
	}
	if level < 1 || level > 5 {
		return baseplugin.Error(400, "invalid soldier level"), nil
	}
	if count <= 0 {
		return baseplugin.Error(400, "count must be positive"), nil
	}

	item, err := l.mgr.StartTrain(ctx.Data, soldierType, level, count, isUpgrade)
	if err != nil {
		return baseplugin.Error(400, err.Error()), nil
	}

	return baseplugin.Success(map[string]any{
		"queue_id":     item.ID,
		"soldier_id":   item.SoldierID,
		"soldier_type": item.SoldierType,
		"level":        item.Level,
		"count":        item.Count,
		"start_time":   item.StartTime,
		"finish_time":  item.FinishTime,
		"is_upgrade":   item.IsUpgrade,
	}), nil
}

// handleCancelTrain 取消训练
func (l *SoldierLogic) handleCancelTrain(ctx *baseplugin.LogicContext, params map[string]any) (*baseplugin.LogicResult, error) {
	queueID := int64(params["queue_id"].(float64))

	if err := l.mgr.CancelTrain(ctx.Data, queueID); err != nil {
		return baseplugin.Error(400, err.Error()), nil
	}

	return baseplugin.Success(map[string]any{
		"queue_id": queueID,
		"message":  "train cancelled",
	}), nil
}

// handleTrainQueue 获取训练队列
func (l *SoldierLogic) handleTrainQueue(ctx *baseplugin.LogicContext, params map[string]any) (*baseplugin.LogicResult, error) {
	items, err := l.mgr.GetTrainQueue(ctx.Data)
	if err != nil {
		return baseplugin.Error(500, err.Error()), nil
	}

	// 获取已完成的训练
	completed := l.mgr.GetCompletedTrains(ctx.Data)

	queue := make([]map[string]any, 0, len(items))
	for _, item := range items {
		queue = append(queue, map[string]any{
			"id":           item.ID,
			"soldier_id":   item.SoldierID,
			"soldier_type": item.SoldierType,
			"level":        item.Level,
			"count":        item.Count,
			"start_time":   item.StartTime,
			"finish_time":  item.FinishTime,
			"is_upgrade":   item.IsUpgrade,
		})
	}

	return baseplugin.Success(map[string]any{
		"queue":     queue,
		"completed": len(completed),
	}), nil
}

// handleCompleteTrain 完成训练
func (l *SoldierLogic) handleCompleteTrain(ctx *baseplugin.LogicContext, params map[string]any) (*baseplugin.LogicResult, error) {
	// 获取已完成的训练
	completed := l.mgr.GetCompletedTrains(ctx.Data)
	if len(completed) == 0 {
		return baseplugin.Error(400, "no completed training"), nil
	}

	// 完成所有已完成的训练
	results := make([]map[string]any, 0, len(completed))
	for _, item := range completed {
		if _, err := l.mgr.CompleteTrain(ctx.Data, item.ID); err != nil {
			continue
		}
		results = append(results, map[string]any{
			"queue_id":   item.ID,
			"soldier_id": item.SoldierID,
			"count":      item.Count,
		})
	}

	return baseplugin.Success(map[string]any{
		"completed": results,
		"count":     len(results),
	}), nil
}

// handleHeal 开始治疗
func (l *SoldierLogic) handleHeal(ctx *baseplugin.LogicContext, params map[string]any) (*baseplugin.LogicResult, error) {
	// 解析士兵列表
	soldiersRaw, ok := params["soldiers"].([]any)
	if !ok {
		return baseplugin.Error(400, "invalid soldiers parameter"), nil
	}

	soldiers := make(map[int]int)
	for _, v := range soldiersRaw {
		if sm, ok := v.(map[string]any); ok {
			soldierID := int(sm["soldier_id"].(float64))
			count := int(sm["count"].(float64))
			soldiers[soldierID] = count
		}
	}

	if len(soldiers) == 0 {
		return baseplugin.Error(400, "no soldiers to heal"), nil
	}

	item, err := l.mgr.StartHeal(ctx.Data, soldiers)
	if err != nil {
		return baseplugin.Error(400, err.Error()), nil
	}

	return baseplugin.Success(map[string]any{
		"queue_id":    item.ID,
		"soldiers":    item.Soldiers,
		"start_time":  item.StartTime,
		"finish_time": item.FinishTime,
	}), nil
}

// handleCancelHeal 取消治疗
func (l *SoldierLogic) handleCancelHeal(ctx *baseplugin.LogicContext, params map[string]any) (*baseplugin.LogicResult, error) {
	if err := l.mgr.CancelHeal(ctx.Data); err != nil {
		return baseplugin.Error(400, err.Error()), nil
	}

	return baseplugin.Success(map[string]any{
		"message": "healing cancelled",
	}), nil
}

// handleHealQueue 获取治疗队列
func (l *SoldierLogic) handleHealQueue(ctx *baseplugin.LogicContext, params map[string]any) (*baseplugin.LogicResult, error) {
	item, err := l.mgr.GetHealQueue(ctx.Data)
	if err != nil {
		return baseplugin.Error(500, err.Error()), nil
	}

	if item == nil {
		return baseplugin.Success(map[string]any{
			"queue":    nil,
			"progress": 0,
		}), nil
	}

	progress := l.mgr.GetHealProgress(ctx.Data)

	return baseplugin.Success(map[string]any{
		"queue": map[string]any{
			"id":          item.ID,
			"soldiers":    item.Soldiers,
			"start_time":  item.StartTime,
			"finish_time": item.FinishTime,
		},
		"progress": progress,
	}), nil
}

// handleCompleteHeal 完成治疗
func (l *SoldierLogic) handleCompleteHeal(ctx *baseplugin.LogicContext, params map[string]any) (*baseplugin.LogicResult, error) {
	item, err := l.mgr.CompleteHeal(ctx.Data)
	if err != nil {
		return baseplugin.Error(400, err.Error()), nil
	}

	return baseplugin.Success(map[string]any{
		"queue_id": item.ID,
		"soldiers": item.Soldiers,
		"message":  "healing completed",
	}), nil
}

// handleDismiss 解散士兵
func (l *SoldierLogic) handleDismiss(ctx *baseplugin.LogicContext, params map[string]any) (*baseplugin.LogicResult, error) {
	soldierID := int(params["soldier_id"].(float64))
	count := int(params["count"].(float64))

	if count <= 0 {
		return baseplugin.Error(400, "count must be positive"), nil
	}

	// 检查是否有足够士兵
	have, err := l.mgr.GetSoldierCount(ctx.Data, soldierID)
	if err != nil {
		return baseplugin.Error(500, err.Error()), nil
	}

	if have < count {
		return baseplugin.Error(400, fmt.Sprintf("not enough soldiers, have %d, need %d", have, count)), nil
	}

	// 扣除士兵
	if err := l.mgr.SubSoldiers(ctx.Data, soldierID, count); err != nil {
		return baseplugin.Error(500, err.Error()), nil
	}

	// 返还部分资源 (简化: 返还训练消耗的50%)
	cfg := GetSoldierConfig(soldierID)
	var refunded map[string]int64
	if cfg != nil {
		refunded = map[string]int64{
			"food":  cfg.CostFood * int64(count) / 2,
			"wood":  cfg.CostWood * int64(count) / 2,
			"stone": cfg.CostStone * int64(count) / 2,
			"gold":  cfg.CostGold * int64(count) / 2,
		}

		// 增加资源
		food := l.mgr.getResource(ctx.Data, "food")
		wood := l.mgr.getResource(ctx.Data, "wood")
		stone := l.mgr.getResource(ctx.Data, "stone")
		gold := l.mgr.getResource(ctx.Data, "gold")

		l.mgr.setResource(ctx.Data, "food", food+refunded["food"])
		l.mgr.setResource(ctx.Data, "wood", wood+refunded["wood"])
		l.mgr.setResource(ctx.Data, "stone", stone+refunded["stone"])
		l.mgr.setResource(ctx.Data, "gold", gold+refunded["gold"])
		ctx.Data.MarkDirty()
	}

	return baseplugin.Success(map[string]any{
		"soldier_id": soldierID,
		"dismissed":  count,
		"refunded":   refunded,
	}), nil
}

// handleConfigs 获取士兵配置
func (l *SoldierLogic) handleConfigs(ctx *baseplugin.LogicContext, params map[string]any) (*baseplugin.LogicResult, error) {
	configs := GetAllSoldierConfigs()

	list := make([]map[string]any, 0, len(configs))
	for _, cfg := range configs {
		list = append(list, map[string]any{
			"id":         cfg.ID,
			"type":       cfg.Type,
			"level":      cfg.Level,
			"name":       cfg.Name,
			"attack":     cfg.Attack,
			"defense":    cfg.Defense,
			"hp":         cfg.HP,
			"speed":      cfg.Speed,
			"load":       cfg.Load,
			"power":      cfg.Power,
			"cost_food":  cfg.CostFood,
			"cost_wood":  cfg.CostWood,
			"cost_stone": cfg.CostStone,
			"cost_gold":  cfg.CostGold,
			"train_time": cfg.TrainTime,
		})
	}

	return baseplugin.Success(map[string]any{
		"configs": list,
		"total":   len(list),
	}), nil
}

// handleStats 获取士兵统计
func (l *SoldierLogic) handleStats(ctx *baseplugin.LogicContext, params map[string]any) (*baseplugin.LogicResult, error) {
	totalCount, _ := l.mgr.GetTotalCount(ctx.Data)
	totalWounded, _ := l.mgr.GetTotalWounded(ctx.Data)
	totalPower, _ := l.mgr.GetTotalPower(ctx.Data)

	return baseplugin.Success(map[string]any{
		"total_count":   totalCount,
		"total_wounded": totalWounded,
		"total_power":   totalPower,
	}), nil
}

// ============ 导出插件 ============

var initOnce bool
var globalMgr *Manager

// init 初始化
func init() {
	// 加载配置
	if err := LoadSoldierConfig("config/soldier.yaml"); err != nil {
		// 尝试相对路径
		LoadSoldierConfig("./config/soldier.yaml")
	}
}

// SoldierModule 导出的插件模块
var SoldierModule baseplugin.LogicModule

// GetSoldierModule 获取士兵模块 (用于外部初始化)
func GetSoldierModule() baseplugin.LogicModule {
	if SoldierModule == nil {
		globalMgr = NewManager()
		SoldierModule = NewSoldierLogic(globalMgr)
	}
	return SoldierModule
}

// GetSoldierManager 获取士兵管理器 (用于外部调用)
func GetSoldierManager() *Manager {
	if globalMgr == nil {
		globalMgr = NewManager()
	}
	return globalMgr
}

// GetSoldierConsumer 获取士兵消费者 (供行军模块通过 .so Lookup 调用)
func GetSoldierConsumer() baseplugin.SoldierConsumer {
	return GetSoldierManager()
}

// StartManager 启动管理器 (由服务器调用)
func StartManager() {
	if globalMgr != nil {
		globalMgr.Start()
	}
}

// StopManager 停止管理器
func StopManager() {
	if globalMgr != nil {
		globalMgr.Stop()
	}
}
