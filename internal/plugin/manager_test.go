package plugin

import (
    "testing"

    "github.com/yourorg/wg_ai/internal/data"
    baseplugin "github.com/yourorg/wg_ai/plugin"
)

// MockLogicModule 模拟逻辑模块
type MockLogicModule struct {
    name string
}

func (m *MockLogicModule) Name() string {
    return m.name
}

func (m *MockLogicModule) Handle(ctx *baseplugin.LogicContext, method string, params map[string]any) (*baseplugin.LogicResult, error) {
    return baseplugin.Success(map[string]any{"method": method}), nil
}

func TestNewManager(t *testing.T) {
    mgr := NewManager()
    if mgr == nil {
        t.Fatal("expected manager")
    }
}

func TestManagerRegisterModule(t *testing.T) {
    mgr := NewManager()

    // 直接注册模块 (不通过 .so 文件)
    mgr.RegisterModule("mock", &MockLogicModule{name: "mock"})

    module := mgr.GetModule("mock")
    if module == nil {
        t.Fatal("module not found")
    }
    if module.Name() != "mock" {
        t.Error("name mismatch")
    }
}

func TestManagerGetModuleNotFound(t *testing.T) {
    mgr := NewManager()

    module := mgr.GetModule("notexist")
    if module != nil {
        t.Error("expected nil for non-existent module")
    }
}

func TestManagerListModules(t *testing.T) {
    mgr := NewManager()
    mgr.RegisterModule("role", &MockLogicModule{name: "role"})
    mgr.RegisterModule("item", &MockLogicModule{name: "item"})

    list := mgr.ListModules()
    if len(list) != 2 {
        t.Errorf("expected 2 modules, got %d", len(list))
    }
}

 func TestManagerCall(t *testing.T) {
    mgr := NewManager()
    mgr.RegisterModule("mock", &MockLogicModule{name: "mock"})

    // 注册路由
    mgr.Router().Register(1001, "mock", "test")

    // 创建上下文
    playerData := data.NewPlayerData(1)
    ctx := &baseplugin.LogicContext{
        RID:  1,
        UID:  1,
        Data: NewDataAdapter(1, playerData),
    }

    // 调用
    result, err := mgr.Call(ctx, 1001, map[string]any{})
    if err != nil {
        t.Fatal(err)
    }
    if result.Code != 0 {
        t.Errorf("expected code 0, got %d", result.Code)
    }
}

func TestManagerCallModuleNotFound(t *testing.T) {
    mgr := NewManager()
    mgr.Router().Register(9999, "notexist", "test")

    playerData := data.NewPlayerData(1)
    ctx := &baseplugin.LogicContext{
        RID:  1,
        Data: NewDataAdapter(1, playerData),
    }

    _, err := mgr.Call(ctx, 9999, nil)
    if err != baseplugin.ErrModuleNotFound {
        t.Errorf("expected ErrModuleNotFound, got %v", err)
    }
}

func TestManagerCallRouteNotFound(t *testing.T) {
    mgr := NewManager()

    playerData := data.NewPlayerData(1)
    ctx := &baseplugin.LogicContext{
        RID:  1,
        Data: NewDataAdapter(1, playerData),
    }

    _, err := mgr.Call(ctx, 8888, nil)
    if err != baseplugin.ErrModuleNotFound {
        t.Errorf("expected ErrModuleNotFound, got %v", err)
    }
}
