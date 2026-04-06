package main

import (
    baseplugin "github.com/luckinbyte/wg_ai/plugin"
)

// RoleLogic 角色逻辑模块
type RoleLogic struct{}

// Name 实现LogicModule接口
func (l *RoleLogic) Name() string {
    return "role"
}

// Handle 处理请求
func (l *RoleLogic) Handle(ctx *baseplugin.LogicContext, method string, params map[string]any) (*baseplugin.LogicResult, error) {
    switch method {
    case "login":
        return l.handleLogin(ctx, params)
    case "heartbeat":
        return l.handleHeartbeat(ctx, params)
    case "get_info":
        return l.handleGetInfo(ctx, params)
    case "update_name":
        return l.handleUpdateName(ctx, params)
    default:
        return nil, baseplugin.ErrMethodNotFound
    }
}

// handleLogin 处理登录
func (l *RoleLogic) handleLogin(ctx *baseplugin.LogicContext, params map[string]any) (*baseplugin.LogicResult, error) {
    name, _ := ctx.Data.GetField("name")
    level, _ := ctx.Data.GetField("level")
    exp, _ := ctx.Data.GetField("exp")

    return baseplugin.Success(map[string]any{
        "rid":   ctx.RID,
        "name":  name,
        "level": level,
        "exp":   exp,
    }), nil
}

// handleHeartbeat 处理心跳
func (l *RoleLogic) handleHeartbeat(ctx *baseplugin.LogicContext, params map[string]any) (*baseplugin.LogicResult, error) {
    return baseplugin.Success(nil), nil
}

// handleGetInfo 获取玩家信息
func (l *RoleLogic) handleGetInfo(ctx *baseplugin.LogicContext, params map[string]any) (*baseplugin.LogicResult, error) {
    name, _ := ctx.Data.GetField("name")
    level, _ := ctx.Data.GetField("level")
    exp, _ := ctx.Data.GetField("exp")
    gold, _ := ctx.Data.GetField("gold")
    vip, _ := ctx.Data.GetField("vip")

    return baseplugin.Success(map[string]any{
        "rid":   ctx.RID,
        "name":  name,
        "level": level,
        "exp":   exp,
        "gold":  gold,
        "vip":   vip,
    }), nil
}

// handleUpdateName 更新名字
func (l *RoleLogic) handleUpdateName(ctx *baseplugin.LogicContext, params map[string]any) (*baseplugin.LogicResult, error) {
    name, ok := params["name"].(string)
    if !ok || name == "" {
        return baseplugin.Error(2, "invalid name"), nil
    }

    // 更新数据
    if err := ctx.Data.SetField("name", name); err != nil {
        return nil, err
    }

    return baseplugin.Success(map[string]any{"name": name}), nil
}

// 导出符号 - 必须命名为 "Role" + "Module"
var RoleModule baseplugin.LogicModule = &RoleLogic{}
