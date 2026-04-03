package plugin

import (
    "encoding/json"
    "os"
)

// RouteConfig 路由配置
type RouteConfig struct {
    MsgID  uint16 `json:"msg_id"`
    Module string `json:"module"`
    Method string `json:"method"`
}

// Router 消息路由器
type Router struct {
    routes map[uint16]*RouteConfig
}

// NewRouter 创建路由器
func NewRouter() *Router {
    return &Router{
        routes: make(map[uint16]*RouteConfig),
    }
}

// Register 注册路由
func (r *Router) Register(msgID uint16, module, method string) {
    r.routes[msgID] = &RouteConfig{
        MsgID:  msgID,
        Module: module,
        Method: method,
    }
}

// Get 获取路由配置
func (r *Router) Get(msgID uint16) (*RouteConfig, bool) {
    cfg, ok := r.routes[msgID]
    return cfg, ok
}

// LoadFromMap 从 map 列表加载路由
func (r *Router) LoadFromMap(routes []map[string]any) error {
    for _, route := range routes {
        msgID, ok := route["msg_id"].(float64)
        if !ok {
            continue
        }

        module, _ := route["module"].(string)
        method, _ := route["method"].(string)

        r.Register(uint16(msgID), module, method)
    }
    return nil
}

// LoadFromConfig 从配置文件加载路由
func (r *Router) LoadFromConfig(path string) error {
    data, err := os.ReadFile(path)
    if err != nil {
        return err
    }

    var routes []map[string]any
    if err := json.Unmarshal(data, &routes); err != nil {
        return err
    }

    return r.LoadFromMap(routes)
}

// All 获取所有路由
func (r *Router) All() map[uint16]*RouteConfig {
    return r.routes
}
