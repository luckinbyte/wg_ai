package main

import (
    "testing"

    "github.com/luckinbyte/wg_ai/internal/data"
    "github.com/luckinbyte/wg_ai/internal/plugin"
    "github.com/luckinbyte/wg_ai/internal/scene"
    cityplugin "github.com/luckinbyte/wg_ai/plugin/city"
    baseplugin "github.com/luckinbyte/wg_ai/plugin"
)

func TestRoleLogicName(t *testing.T) {
    logic := &RoleLogic{}
    if logic.Name() != "role" {
        t.Errorf("expected 'role', got '%s'", logic.Name())
    }
}

func TestRoleLogicLogin(t *testing.T) {
    logic := &RoleLogic{}

    playerData := data.NewPlayerData(1)
    playerData.SetField("name", "player1")
    playerData.SetField("level", int64(10))
    playerData.SetField("exp", int64(5000))

    ctx := &baseplugin.LogicContext{
        RID:  1,
        UID:  100,
        Data: plugin.NewDataAdapter(1, playerData),
    }

    result, err := logic.Handle(ctx, "login", nil)
    if err != nil {
        t.Fatal(err)
    }
    if result.Code != 0 {
        t.Errorf("expected code 0, got %d", result.Code)
    }
    if result.Data["name"] != "player1" {
        t.Error("name mismatch")
    }
    if result.Data["level"] != int64(10) {
        t.Error("level mismatch")
    }
}

func TestRoleLogicGetInfo(t *testing.T) {
    logic := &RoleLogic{}

    playerData := data.NewPlayerData(1)
    playerData.SetField("name", "test")
    playerData.SetField("level", int64(5))
    playerData.SetField("exp", int64(100))
    playerData.SetField("gold", int64(1000))
    playerData.SetField("vip", int64(0))

    ctx := &baseplugin.LogicContext{
        RID:  1,
        Data: plugin.NewDataAdapter(1, playerData),
    }

    result, err := logic.Handle(ctx, "get_info", nil)
    if err != nil {
        t.Fatal(err)
    }
    if result.Code != 0 {
        t.Errorf("expected code 0, got %d", result.Code)
    }
    if result.Data["rid"] != int64(1) {
        t.Error("rid mismatch")
    }
}

func TestRoleLogicUpdateName(t *testing.T) {
    logic := &RoleLogic{}

    playerData := data.NewPlayerData(1)
    playerData.SetField("name", "oldname")

    ctx := &baseplugin.LogicContext{
        RID:  1,
        Data: plugin.NewDataAdapter(1, playerData),
    }

    result, err := logic.Handle(ctx, "update_name", map[string]any{
        "name": "newname",
    })
    if err != nil {
        t.Fatal(err)
    }
    if result.Code != 0 {
        t.Errorf("expected code 0, got %d", result.Code)
    }

    // 验证数据已更新
    name, _ := ctx.Data.GetField("name")
    if name != "newname" {
        t.Errorf("expected 'newname', got '%v'", name)
    }
}

func TestRoleLogicMethodNotFound(t *testing.T) {
    logic := &RoleLogic{}

    playerData := data.NewPlayerData(1)
    ctx := &baseplugin.LogicContext{
        RID:  1,
        Data: plugin.NewDataAdapter(1, playerData),
    }

    _, err := logic.Handle(ctx, "unknown_method", nil)
    if err != baseplugin.ErrMethodNotFound {
        t.Errorf("expected ErrMethodNotFound, got %v", err)
    }
}

func TestRoleLogicLoginInitCity(t *testing.T) {
    if err := cityplugin.LoadBuildingConfig("../../config/building.yaml"); err != nil {
        t.Fatal(err)
    }

    sceneMgr := scene.NewManager()
    sceneMgr.CreateScene(scene.SceneConfig{ID: 1, Width: 1000, Height: 1000, GridSize: 50})

    cityplugin.SetSceneManager(sceneMgr)
    logic := &RoleLogic{}
    playerData := data.NewPlayerData(1)
    playerData.SetField("name", "player1")
    playerData.SetField("level", int64(10))
    playerData.SetField("exp", int64(5000))

    ctx := &baseplugin.LogicContext{
        RID:  1,
        UID:  100,
        Data: plugin.NewDataAdapter(1, playerData),
    }

    result, err := logic.Handle(ctx, "login", nil)
    if err != nil {
        t.Fatal(err)
    }
    if result.Code != 0 {
        t.Fatalf("expected code 0, got %d", result.Code)
    }

    cityData, err := cityplugin.GetCity(ctx.Data)
    if err != nil {
        t.Fatal(err)
    }
    if cityData == nil {
        t.Fatal("expected city data to be initialized")
    }
    if cityData.CityID == 0 {
        t.Fatal("expected city entity id")
    }
}
