package admin

import (
    "encoding/json"
    "net/http"

    "github.com/yourorg/wg_ai/internal/plugin"
)

// Handler 管理接口处理器
type Handler struct {
    pluginMgr *plugin.Manager
}

// NewHandler 创建管理接口处理器
func NewHandler(pluginMgr *plugin.Manager) *Handler {
    return &Handler{pluginMgr: pluginMgr}
}

// RegisterRoutes 注册路由
func (h *Handler) RegisterRoutes(mux *http.ServeMux) {
    mux.HandleFunc("/admin/hotreload", h.handleHotReload)
    mux.HandleFunc("/admin/plugins", h.handleListPlugins)
    mux.HandleFunc("/admin/health", h.handleHealth)
    mux.HandleFunc("/admin/routes", h.handleListRoutes)
}

// HotReloadRequest 热更请求
type HotReloadRequest struct {
    Module string `json:"module"` // 模块名: role, item
    Path   string `json:"path"`   // 插件路径: ./plugins/role.so
}

// HotReloadResponse 热更响应
type HotReloadResponse struct {
    Success bool   `json:"success"`
    Message string `json:"message"`
    Module  string `json:"module"`
}

// handleHotReload 处理热更请求
// POST /admin/hotreload
func (h *Handler) handleHotReload(w http.ResponseWriter, r *http.Request) {
    if r.Method != http.MethodPost {
        http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
        return
    }

    var req HotReloadRequest
    if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
        http.Error(w, err.Error(), http.StatusBadRequest)
        return
    }

    // 参数校验
    if req.Module == "" || req.Path == "" {
        w.Header().Set("Content-Type", "application/json")
        json.NewEncoder(w).Encode(HotReloadResponse{
            Success: false,
            Message: "module and path are required",
            Module:  req.Module,
        })
        return
    }

    // 执行热更
    err := h.pluginMgr.HotReload(req.Module, req.Path)
    if err != nil {
        w.Header().Set("Content-Type", "application/json")
        json.NewEncoder(w).Encode(HotReloadResponse{
            Success: false,
            Message: err.Error(),
            Module:  req.Module,
        })
        return
    }

    w.Header().Set("Content-Type", "application/json")
    json.NewEncoder(w).Encode(HotReloadResponse{
        Success: true,
        Message: "hot reload success",
        Module:  req.Module,
    })
}

// handleListPlugins 列出已加载插件
// GET /admin/plugins
func (h *Handler) handleListPlugins(w http.ResponseWriter, r *http.Request) {
    modules := h.pluginMgr.ListModules()

    w.Header().Set("Content-Type", "application/json")
    json.NewEncoder(w).Encode(map[string]any{
        "plugins": modules,
        "count":   len(modules),
    })
}

// handleHealth 健康检查
// GET /admin/health
func (h *Handler) handleHealth(w http.ResponseWriter, r *http.Request) {
    w.Header().Set("Content-Type", "application/json")
    json.NewEncoder(w).Encode(map[string]any{
        "status": "ok",
    })
}

// handleListRoutes 列出所有路由
// GET /admin/routes
func (h *Handler) handleListRoutes(w http.ResponseWriter, r *http.Request) {
    routes := h.pluginMgr.Router().All()

    w.Header().Set("Content-Type", "application/json")
    json.NewEncoder(w).Encode(map[string]any{
        "routes": routes,
        "count":  len(routes),
    })
}
