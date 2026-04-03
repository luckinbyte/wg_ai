package plugin_test

import (
    "testing"

    "github.com/yourorg/wg_ai/plugin"
)

func TestRouterRegister(t *testing.T) {
    r := plugin.NewRouter()
    r.Register(1001, "role", "login")

    cfg, ok := r.Get(1001)
    if !ok {
        t.Fatal("route not found")
    }
    if cfg.Module != "role" {
        t.Errorf("expected module 'role', got '%s'", cfg.Module)
    }
    if cfg.Method != "login" {
        t.Errorf("expected method 'login', got '%s'", cfg.Method)
    }
}

func TestRouterNotFound(t *testing.T) {
    r := plugin.NewRouter()
    _, ok := r.Get(9999)
    if ok {
        t.Error("expected not found")
    }
}

func TestRouterLoadFromMap(t *testing.T) {
    r := plugin.NewRouter()

    routes := []map[string]any{
        {"msg_id": float64(1001), "module": "role", "method": "login"},
        {"msg_id": float64(2001), "module": "item", "method": "use"},
    }

    if err := r.LoadFromMap(routes); err != nil {
        t.Fatal(err)
    }

    cfg, ok := r.Get(1001)
    if !ok || cfg.Module != "role" {
        t.Error("route 1001 mismatch")
    }

    cfg, ok = r.Get(2001)
    if !ok || cfg.Module != "item" {
        t.Error("route 2001 mismatch")
    }
}
