package main

import (
    "testing"

    "github.com/luckinbyte/wg_ai/internal/data"
    "github.com/luckinbyte/wg_ai/internal/plugin"
    baseplugin "github.com/luckinbyte/wg_ai/plugin"
)

func TestItemLogicName(t *testing.T) {
    logic := &ItemLogic{}
    if logic.Name() != "item" {
        t.Errorf("expected 'item', got '%s'", logic.Name())
    }
}

func TestItemLogicList(t *testing.T) {
    logic := &ItemLogic{}

    playerData := data.NewPlayerData(1)
    items := &[]ItemData{
        {ID: 1, CfgID: 100, Count: 10},
        {ID: 2, CfgID: 101, Count: 5},
    }
    playerData.Arrays["items"] = items

    ctx := &baseplugin.LogicContext{
        RID:  1,
        Data: plugin.NewDataAdapter(1, playerData),
    }

    result, err := logic.Handle(ctx, "list", nil)
    if err != nil {
        t.Fatal(err)
    }
    if result.Code != 0 {
        t.Errorf("expected code 0, got %d", result.Code)
    }

    itemsResult, ok := result.Data["items"].(*[]ItemData)
    if !ok {
        t.Fatal("items type error")
    }
    if len(*itemsResult) != 2 {
        t.Errorf("expected 2 items, got %d", len(*itemsResult))
    }
}

func TestItemLogicUse(t *testing.T) {
    logic := &ItemLogic{}

    playerData := data.NewPlayerData(1)
    items := &[]ItemData{
        {ID: 1, CfgID: 100, Count: 10},
    }
    playerData.Arrays["items"] = items

    ctx := &baseplugin.LogicContext{
        RID:  1,
        Data: plugin.NewDataAdapter(1, playerData),
    }

    // 使用 3 个
    result, err := logic.Handle(ctx, "use", map[string]any{
        "item_id": int64(1),
        "count":   int64(3),
    })
    if err != nil {
        t.Fatal(err)
    }
    if result.Code != 0 {
        t.Errorf("expected code 0, got %d", result.Code)
    }

    // 验证剩余数量
    itemsResult := *items
    if itemsResult[0].Count != 7 {
        t.Errorf("expected 7 remaining, got %d", itemsResult[0].Count)
    }
}

func TestItemLogicUseNotEnough(t *testing.T) {
    logic := &ItemLogic{}

    playerData := data.NewPlayerData(1)
    items := &[]ItemData{
        {ID: 1, CfgID: 100, Count: 5},
    }
    playerData.Arrays["items"] = items

    ctx := &baseplugin.LogicContext{
        RID:  1,
        Data: plugin.NewDataAdapter(1, playerData),
    }

    // 尝试使用 10 个 (只有 5 个)
    result, err := logic.Handle(ctx, "use", map[string]any{
        "item_id": int64(1),
        "count":   int64(10),
    })
    if err != nil {
        t.Fatal(err)
    }
    if result.Code == 0 {
        t.Error("expected error code for not enough items")
    }
}

func TestItemLogicAdd(t *testing.T) {
    logic := &ItemLogic{}

    playerData := data.NewPlayerData(1)
    items := &[]ItemData{}
    playerData.Arrays["items"] = items

    ctx := &baseplugin.LogicContext{
        RID:  1,
        Data: plugin.NewDataAdapter(1, playerData),
    }

    result, err := logic.Handle(ctx, "add", map[string]any{
        "cfg_id": int64(100),
        "count":  int64(5),
    })
    if err != nil {
        t.Fatal(err)
    }
    if result.Code != 0 {
        t.Errorf("expected code 0, got %d", result.Code)
    }

    // 验证物品已添加
    itemsResult := *items
    if len(itemsResult) != 1 {
        t.Errorf("expected 1 item, got %d", len(itemsResult))
    }
    if itemsResult[0].CfgID != 100 {
        t.Error("cfg_id mismatch")
    }
    if itemsResult[0].Count != 5 {
        t.Error("count mismatch")
    }
}
