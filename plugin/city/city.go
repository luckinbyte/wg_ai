package city

import (
	"fmt"

	"github.com/luckinbyte/wg_ai/internal/scene"
	baseplugin "github.com/luckinbyte/wg_ai/plugin"
)

var globalSceneMgr *scene.Manager

type Logic struct {
	mgr *Manager
}

func SetSceneManager(sceneMgr *scene.Manager) {
	globalSceneMgr = sceneMgr
}

func HasSceneManager() bool {
	return globalSceneMgr != nil
}

func NewModule(sceneMgr *scene.Manager) (baseplugin.LogicModule, error) {
	if err := loadBuildingConfig(); err != nil {
		return nil, err
	}
	SetSceneManager(sceneMgr)
	return &Logic{mgr: NewManager(sceneMgr)}, nil
}

func loadBuildingConfig() error {
	if err := LoadBuildingConfig("config/building.yaml"); err == nil {
		return nil
	}
	if err := LoadBuildingConfig("./config/building.yaml"); err == nil {
		return nil
	}
	return fmt.Errorf("load building config failed")
}

func (l *Logic) Name() string {
	return "city"
}

func (l *Logic) Handle(ctx *baseplugin.LogicContext, method string, params map[string]any) (*baseplugin.LogicResult, error) {
	if l.mgr == nil {
		return baseplugin.Error(500, "city manager not initialized"), nil
	}

	switch method {
	case "get_info":
		return l.handleGetInfo(ctx, params)
	case "upgrade":
		return l.handleUpgrade(ctx, params)
	case "cancel_build":
		return l.handleCancelBuild(ctx, params)
	case "build_queue":
		return l.handleBuildQueue(ctx, params)
	case "production":
		return l.handleProduction(ctx, params)
	default:
		return nil, baseplugin.ErrMethodNotFound
	}
}

func (l *Logic) handleGetInfo(ctx *baseplugin.LogicContext, params map[string]any) (*baseplugin.LogicResult, error) {
	completed, err := l.mgr.CompleteReadyBuilds(ctx.Data)
	if err != nil {
		return baseplugin.Error(500, err.Error()), nil
	}
	city, err := l.mgr.GetCity(ctx.Data)
	if err != nil {
		return baseplugin.Error(500, err.Error()), nil
	}
	if city == nil {
		return baseplugin.Error(404, "city not found"), nil
	}
	return baseplugin.Success(map[string]any{
		"city":      city,
		"completed": completed,
	}), nil
}

func (l *Logic) handleUpgrade(ctx *baseplugin.LogicContext, params map[string]any) (*baseplugin.LogicResult, error) {
	buildingType, ok := params["building_type"].(float64)
	if !ok {
		return baseplugin.Error(400, "building_type required"), nil
	}
	item, err := l.mgr.UpgradeBuilding(ctx.Data, int(buildingType))
	if err != nil {
		return baseplugin.Error(400, err.Error()), nil
	}
	return baseplugin.Success(map[string]any{"queue": item}), nil
}

func (l *Logic) handleCancelBuild(ctx *baseplugin.LogicContext, params map[string]any) (*baseplugin.LogicResult, error) {
	queueID, ok := params["queue_id"].(float64)
	if !ok {
		return baseplugin.Error(400, "queue_id required"), nil
	}
	if err := l.mgr.CancelBuild(ctx.Data, int64(queueID)); err != nil {
		return baseplugin.Error(400, err.Error()), nil
	}
	return baseplugin.Success(map[string]any{"queue_id": int64(queueID)}), nil
}

func (l *Logic) handleBuildQueue(ctx *baseplugin.LogicContext, params map[string]any) (*baseplugin.LogicResult, error) {
	completed, err := l.mgr.CompleteReadyBuilds(ctx.Data)
	if err != nil {
		return baseplugin.Error(500, err.Error()), nil
	}
	queue, err := l.mgr.GetBuildQueue(ctx.Data)
	if err != nil {
		return baseplugin.Error(500, err.Error()), nil
	}
	return baseplugin.Success(map[string]any{
		"queue":     queue,
		"completed": completed,
	}), nil
}

func (l *Logic) handleProduction(ctx *baseplugin.LogicContext, params map[string]any) (*baseplugin.LogicResult, error) {
	food, wood, stone, gold, err := l.mgr.GetProduction(ctx.Data)
	if err != nil {
		return baseplugin.Error(500, err.Error()), nil
	}
	return baseplugin.Success(map[string]any{
		"food":  food,
		"wood":  wood,
		"stone": stone,
		"gold":  gold,
	}), nil
}

func InitPlayerCity(data baseplugin.DataAccessor, rid int64) (*CityData, error) {
	if globalSceneMgr == nil {
		return nil, fmt.Errorf("scene manager not initialized")
	}
	if err := LoadBuildingConfig("config/building.yaml"); err != nil {
		_ = LoadBuildingConfig("./config/building.yaml")
	}
	mgr := NewManager(globalSceneMgr)
	x, y := DefaultCityPosition(rid)
	return mgr.InitCity(data, x, y)
}

func DefaultCityPosition(rid int64) (float64, float64) {
	baseX := float64(100 + (rid%100)*10)
	baseY := float64(100 + ((rid/100)%100)*10)
	return baseX, baseY
}

func GetCity(data baseplugin.DataAccessor) (*CityData, error) {
	return (&Manager{}).GetCity(data)
}
