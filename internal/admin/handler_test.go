package admin

import (
    "bytes"
    "encoding/json"
    "net/http"
    "net/http/httptest"
    "testing"

    "github.com/luckinbyte/wg_ai/internal/plugin"
)

func TestHandleHotReload(t *testing.T) {
    // 创建插件管理器
    mgr := plugin.NewManager()
    handler := NewHandler(mgr)

    // 创建测试服务器
    mux := http.NewServeMux()
    handler.RegisterRoutes(mux)
    srv := httptest.NewServer(mux)
    defer srv.Close()

    // 测试热更请求 (不存在的插件，会失败)
    body := HotReloadRequest{
        Module: "test",
        Path:   "./plugins/test.so",
    }
    jsonBody, _ := json.Marshal(body)

    resp, err := http.Post(srv.URL+"/admin/hotreload", "application/json", bytes.NewReader(jsonBody))
    if err != nil {
        t.Fatal(err)
    }
    defer resp.Body.Close()

    // 预期失败 (插件不存在)
    var result HotReloadResponse
    json.NewDecoder(resp.Body).Decode(&result)

    if result.Module != "test" {
        t.Error("module mismatch")
    }
}

func TestHandleListPlugins(t *testing.T) {
    mgr := plugin.NewManager()
    handler := NewHandler(mgr)

    mux := http.NewServeMux()
    handler.RegisterRoutes(mux)
    srv := httptest.NewServer(mux)
    defer srv.Close()

    resp, err := http.Get(srv.URL + "/admin/plugins")
    if err != nil {
        t.Fatal(err)
    }
    defer resp.Body.Close()

    if resp.StatusCode != http.StatusOK {
        t.Errorf("expected 200, got %d", resp.StatusCode)
    }
}

func TestHandleHealth(t *testing.T) {
    mgr := plugin.NewManager()
    handler := NewHandler(mgr)

    mux := http.NewServeMux()
    handler.RegisterRoutes(mux)
    srv := httptest.NewServer(mux)
    defer srv.Close()

    resp, err := http.Get(srv.URL + "/admin/health")
    if err != nil {
        t.Fatal(err)
    }
    defer resp.Body.Close()

    var result map[string]any
    json.NewDecoder(resp.Body).Decode(&result)

    if result["status"] != "ok" {
        t.Error("health check failed")
    }
}

func TestHandleListRoutes(t *testing.T) {
    mgr := plugin.NewManager()
    handler := NewHandler(mgr)

    // 注册一些路由
    mgr.Router().Register(1001, "role", "login")
    mgr.Router().Register(2001, "item", "list")

    mux := http.NewServeMux()
    handler.RegisterRoutes(mux)
    srv := httptest.NewServer(mux)
    defer srv.Close()

    resp, err := http.Get(srv.URL + "/admin/routes")
    if err != nil {
        t.Fatal(err)
    }
    defer resp.Body.Close()

    var result map[string]any
    json.NewDecoder(resp.Body).Decode(&result)

    count, _ := result["count"].(float64)
    if int(count) != 2 {
        t.Errorf("expected 2 routes, got %v", count)
    }
}
