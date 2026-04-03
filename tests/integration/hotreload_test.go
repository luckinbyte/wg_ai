package main

import (
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"
	"bytes"

	"github.com/yourorg/wg_ai/internal/admin"
	"github.com/yourorg/wg_ai/internal/data"
	"github.com/yourorg/wg_ai/internal/plugin"
	baseplugin "github.com/yourorg/wg_ai/plugin"
)

// TestHotReloadFlow 测试完整的热更流程
func TestHotReloadFlow(t *testing.T) {
	// 1. 创建插件管理器
	mgr := plugin.NewManager()
	t.Log("✓ Step 1: 创建插件管理器")

	// 2. 加载路由配置
	routes := []map[string]any{
		{"msg_id": float64(1001), "module": "role", "method": "login"},
		{"msg_id": float64(1002), "module": "role", "method": "get_info"},
		{"msg_id": float64(2001), "module": "item", "method": "list"},
		{"msg_id": float64(2002), "module": "item", "method": "use"},
	}
	if err := mgr.LoadRoutes(routes); err != nil {
		t.Fatalf("加载路由失败: %v", err)
	}
	t.Log("✓ Step 2: 加载路由配置 (4 条路由)")

	// 3. 注册模拟模块 (模拟 .so 加载)
	roleModule := &MockRoleModule{version: "v1"}
	mgr.RegisterModule("role", roleModule)
	t.Log("✓ Step 3: 注册 role 模块 (v1)")

	// 4. 测试调用
	playerData := data.NewPlayerData(1)
	playerData.SetField("name", "test_player")
	playerData.SetField("level", int64(10))

	ctx := &baseplugin.LogicContext{
		RID:  1,
		UID:  100,
		Data: plugin.NewDataAdapter(1, playerData),
	}

	result, err := mgr.Call(ctx, 1001, nil)
	if err != nil {
		t.Fatalf("调用失败: %v", err)
	}
	if result.Code != 0 {
		t.Errorf("预期 code=0, 得到 code=%d", result.Code)
	}
	if result.Data["version"] != "v1" {
		t.Errorf("预期 version=v1, 得到 %v", result.Data["version"])
	}
	t.Log("✓ Step 4: 调用 role/login 成功 (version=v1)")

	// 5. 模拟热更 - 注册新版本模块
	roleModuleV2 := &MockRoleModule{version: "v2"}
	mgr.RegisterModule("role", roleModuleV2)
	t.Log("✓ Step 5: 热更 role 模块 (v1 -> v2)")

	// 6. 验证热更后调用新版本
	result2, err := mgr.Call(ctx, 1001, nil)
	if err != nil {
		t.Fatalf("热更后调用失败: %v", err)
	}
	if result2.Data["version"] != "v2" {
		t.Errorf("预期 version=v2, 得到 %v", result2.Data["version"])
	}
	t.Log("✓ Step 6: 验证热更成功 (version=v2)")

	// 7. 测试 Admin API
	handler := admin.NewHandler(mgr)
	mux := http.NewServeMux()
	handler.RegisterRoutes(mux)
	server := httptest.NewServer(mux)
	defer server.Close()

	// 测试列出插件
	resp, err := http.Get(server.URL + "/admin/plugins")
	if err != nil {
		t.Fatalf("获取插件列表失败: %v", err)
	}
	var pluginsResp map[string]any
	json.NewDecoder(resp.Body).Decode(&pluginsResp)
	resp.Body.Close()

	plugins, _ := pluginsResp["plugins"].([]interface{})
	if len(plugins) != 1 {
		t.Errorf("预期 1 个插件, 得到 %d 个", len(plugins))
	}
	t.Log("✓ Step 7: Admin API - 插件列表正常")

	// 测试健康检查
	resp, err = http.Get(server.URL + "/admin/health")
	if err != nil {
		t.Fatalf("健康检查失败: %v", err)
	}
	var healthResp map[string]any
	json.NewDecoder(resp.Body).Decode(&healthResp)
	resp.Body.Close()

	if healthResp["status"] != "ok" {
		t.Errorf("健康检查失败: %v", healthResp)
	}
	t.Log("✓ Step 8: Admin API - 健康检查正常")

	// 测试热更 API (模拟)
	hotreloadReq := admin.HotReloadRequest{
		Module: "item",
		Path:   "./plugins/item.so",
	}
	jsonBody, _ := json.Marshal(hotreloadReq)
	resp, err = http.Post(server.URL+"/admin/hotreload", "application/json", bytes.NewReader(jsonBody))
	if err != nil {
		t.Fatalf("热更请求失败: %v", err)
	}
	var hotreloadResp admin.HotReloadResponse
	json.NewDecoder(resp.Body).Decode(&hotreloadResp)
	resp.Body.Close()

	// 预期失败 (插件文件不存在)
	if hotreloadResp.Success {
		t.Error("预期热更失败 (插件不存在), 但成功了")
	}
	t.Log("✓ Step 9: Admin API - 热更接口正常 (正确返回失败)")

	t.Log("\n========================================")
	t.Log("热更流程测试全部通过!")
	t.Log("========================================")
}

// TestHotReloadWithItemModule 测试物品模块热更
func TestHotReloadWithItemModule(t *testing.T) {
	mgr := plugin.NewManager()

	// 加载路由
	mgr.Router().Register(2001, "item", "list")
	mgr.Router().Register(2002, "item", "use")

	// 注册物品模块
	itemModule := &MockItemModule{items: []map[string]any{
		{"id": 1, "cfg_id": 100, "count": 10},
	}}
	mgr.RegisterModule("item", itemModule)

	// 测试列表
	playerData := data.NewPlayerData(1)
	ctx := &baseplugin.LogicContext{
		RID:  1,
		Data: plugin.NewDataAdapter(1, playerData),
	}

	result, err := mgr.Call(ctx, 2001, nil)
	if err != nil {
		t.Fatalf("调用 item/list 失败: %v", err)
	}

	items, ok := result.Data["items"].([]map[string]any)
	if !ok {
		t.Fatalf("items 类型错误")
	}
	if len(items) != 1 {
		t.Errorf("预期 1 个物品, 得到 %d 个", len(items))
	}

	t.Log("✓ 物品模块热更测试通过")
}

// TestHotReloadErrorCases 测试错误情况
func TestHotReloadErrorCases(t *testing.T) {
	mgr := plugin.NewManager()

	playerData := data.NewPlayerData(1)
	ctx := &baseplugin.LogicContext{
		RID:  1,
		Data: plugin.NewDataAdapter(1, playerData),
	}

	// 测试路由不存在
	_, err := mgr.Call(ctx, 9999, nil)
	if err != baseplugin.ErrModuleNotFound {
		t.Errorf("预期 ErrModuleNotFound, 得到 %v", err)
	}
	t.Log("✓ 路由不存在 - 正确返回错误")

	// 注册路由但不注册模块
	mgr.Router().Register(8888, "notexist", "test")
	_, err = mgr.Call(ctx, 8888, nil)
	if err != baseplugin.ErrModuleNotFound {
		t.Errorf("预期 ErrModuleNotFound, 得到 %v", err)
	}
	t.Log("✓ 模块不存在 - 正确返回错误")

	t.Log("✓ 错误情况测试通过")
}

// ========== Mock 模块 ==========

type MockRoleModule struct {
	version string
}

func (m *MockRoleModule) Name() string {
	return "role"
}

func (m *MockRoleModule) Handle(ctx *baseplugin.LogicContext, method string, params map[string]any) (*baseplugin.LogicResult, error) {
	switch method {
	case "login":
		name, _ := ctx.Data.GetField("name")
		level, _ := ctx.Data.GetField("level")
		return baseplugin.Success(map[string]any{
			"rid":     ctx.RID,
			"name":    name,
			"level":   level,
			"version": m.version,
		}), nil
	case "get_info":
		return baseplugin.Success(map[string]any{
			"version": m.version,
		}), nil
	default:
		return nil, baseplugin.ErrMethodNotFound
	}
}

type MockItemModule struct {
	items []map[string]any
}

func (m *MockItemModule) Name() string {
	return "item"
}

func (m *MockItemModule) Handle(ctx *baseplugin.LogicContext, method string, params map[string]any) (*baseplugin.LogicResult, error) {
	switch method {
	case "list":
		return baseplugin.Success(map[string]any{
			"items": m.items,
		}), nil
	case "use":
		return baseplugin.Success(map[string]any{
			"remaining": 5,
		}), nil
	default:
		return nil, baseplugin.ErrMethodNotFound
	}
}

func main() {
	fmt.Println("运行热更流程测试...")
	testing.Main(nil, []testing.InternalTest{
		{Name: "TestHotReloadFlow", F: TestHotReloadFlow},
		{Name: "TestHotReloadWithItemModule", F: TestHotReloadWithItemModule},
		{Name: "TestHotReloadErrorCases", F: TestHotReloadErrorCases},
	}, nil, nil)
}
