package main

import (
    "time"

    baseplugin "github.com/luckinbyte/wg_ai/plugin"
)

// ItemData 物品数据结构
type ItemData struct {
    ID    int64 `json:"id"`     // 唯一ID
    CfgID int64 `json:"cfg_id"` // 配置ID
    Count int64 `json:"count"`  // 数量
}

// ItemLogic 物品逻辑模块
type ItemLogic struct{}

// Name 实现LogicModule接口
func (l *ItemLogic) Name() string {
    return "item"
}

// Handle 处理请求
func (l *ItemLogic) Handle(ctx *baseplugin.LogicContext, method string, params map[string]any) (*baseplugin.LogicResult, error) {
    switch method {
    case "list":
        return l.handleList(ctx, params)
    case "use":
        return l.handleUse(ctx, params)
    case "add":
        return l.handleAdd(ctx, params)
    default:
        return nil, baseplugin.ErrMethodNotFound
    }
}

// handleList 获取背包列表
func (l *ItemLogic) handleList(ctx *baseplugin.LogicContext, params map[string]any) (*baseplugin.LogicResult, error) {
    items, err := ctx.Data.GetArray("items")
    if err != nil {
        return nil, err
    }

    // 如果背包为空，返回空列表
    if items == nil {
        return baseplugin.Success(map[string]any{
            "items": &[]ItemData{},
        }), nil
    }

    return baseplugin.Success(map[string]any{
        "items": items,
    }), nil
}

// handleUse 使用物品
func (l *ItemLogic) handleUse(ctx *baseplugin.LogicContext, params map[string]any) (*baseplugin.LogicResult, error) {
    itemID, _ := params["item_id"].(int64)
    count, _ := params["count"].(int64)
    if count <= 0 {
        count = 1
    }

    // 获取背包数据指针
    itemsAny, err := ctx.Data.GetArray("items")
    if err != nil {
        return nil, err
    }

    if itemsAny == nil {
        return baseplugin.Error(202, "item not found"), nil
    }

    items, ok := itemsAny.(*[]ItemData)
    if !ok {
        return baseplugin.Error(201, "invalid items data"), nil
    }

    // 查找并扣除
    for i, item := range *items {
        if item.ID == itemID {
            if item.Count < count {
                return baseplugin.Error(200, "not enough items"), nil
            }

            (*items)[i].Count -= count

            // 删除空物品
            if (*items)[i].Count <= 0 {
                *items = append((*items)[:i], (*items)[i+1:]...)
            }

            ctx.Data.MarkDirty()

            remaining := int64(0)
            if i < len(*items) {
                remaining = (*items)[i].Count
            }

            return baseplugin.Success(map[string]any{
                "remaining": remaining,
            }), nil
        }
    }

    return baseplugin.Error(202, "item not found"), nil
}

// handleAdd 添加物品
func (l *ItemLogic) handleAdd(ctx *baseplugin.LogicContext, params map[string]any) (*baseplugin.LogicResult, error) {
    cfgID, _ := params["cfg_id"].(int64)
    count, _ := params["count"].(int64)
    if count <= 0 {
        count = 1
    }

    // 获取背包数据指针
    itemsAny, err := ctx.Data.GetArray("items")
    if err != nil {
        return nil, err
    }

    var items *[]ItemData
    if itemsAny == nil {
        // 初始化背包
        newItems := make([]ItemData, 0)
        items = &newItems
        // 设置新数组
        ctx.Data.MarkDirty()
    } else {
        items, _ = itemsAny.(*[]ItemData)
    }

    // 添加物品
    newItem := ItemData{
        ID:    generateItemID(),
        CfgID: cfgID,
        Count: count,
    }
    *items = append(*items, newItem)

    ctx.Data.MarkDirty()

    return baseplugin.Success(map[string]any{
        "item": newItem,
    }), nil
}

// generateItemID 生成物品唯一ID
func generateItemID() int64 {
    return time.Now().UnixNano()
}

// 导出符号
var ItemModule baseplugin.LogicModule = &ItemLogic{}
