package plugin

import (
    goplugin "plugin"
    "sync"

    baseplugin "github.com/yourorg/wg_ai/plugin"
)

// Manager 插件管理器
type Manager struct {
    plugins map[string]*goplugin.Plugin // module -> go plugin (.so)
    modules map[string]baseplugin.LogicModule // module -> LogicModule
    router  *baseplugin.Router // 消息路由
    mutex   sync.RWMutex
}

// NewManager 创建插件管理器
func NewManager() *Manager {
    return &Manager{
        plugins: make(map[string]*goplugin.Plugin),
        modules: make(map[string]baseplugin.LogicModule),
        router:  baseplugin.NewRouter(),
    }
}

// RegisterModule 直接注册模块 (用于测试或不使用 .so 的场景)
func (m *Manager) RegisterModule(name string, module baseplugin.LogicModule) {
    m.mutex.Lock()
    defer m.mutex.Unlock()
    m.modules[name] = module
}

// LoadPlugin 加载 .so 插件
func (m *Manager) LoadPlugin(moduleName, path string) error {
    // 1. 打开 .so 文件
    p, err := goplugin.Open(path)
    if err != nil {
        return err
    }

    // 2. 查找导出符号 (如 RoleModule, ItemModule)
    symName := moduleName + "Module"
    sym, err := p.Lookup(symName)
    if err != nil {
        return err
    }

    // 3. 类型断言为 LogicModule
    logicModule, ok := sym.(baseplugin.LogicModule)
    if !ok {
        return baseplugin.ErrInvalidModule
    }

    // 4. 注册模块
    m.mutex.Lock()
    defer m.mutex.Unlock()
    m.plugins[moduleName] = p
    m.modules[moduleName] = logicModule

    return nil
}

// HotReload 热更新插件 (Go plugin 无法卸载，直接加载新版本覆盖)
func (m *Manager) HotReload(moduleName, newPath string) error {
    return m.LoadPlugin(moduleName, newPath)
}

// GetModule 获取模块
func (m *Manager) GetModule(name string) baseplugin.LogicModule {
    m.mutex.RLock()
    defer m.mutex.RUnlock()
    return m.modules[name]
}

// ListModules 列出所有已加载模块
func (m *Manager) ListModules() []string {
    m.mutex.RLock()
    defer m.mutex.RUnlock()

    list := make([]string, 0, len(m.modules))
    for name := range m.modules {
        list = append(list, name)
    }
    return list
}

// Router 获取路由器
func (m *Manager) Router() *baseplugin.Router {
    return m.router
}

// LoadRoutes 加载路由配置
func (m *Manager) LoadRoutes(routes []map[string]any) error {
    return m.router.LoadFromMap(routes)
}

// Call 调用逻辑处理
func (m *Manager) Call(ctx *baseplugin.LogicContext, msgID uint16, params map[string]any) (*baseplugin.LogicResult, error) {
    // 1. 根据消息ID获取路由
    route, ok := m.router.Get(msgID)
    if !ok {
        return nil, baseplugin.ErrModuleNotFound
    }

    // 2. 获取模块
    m.mutex.RLock()
    module := m.modules[route.Module]
    m.mutex.RUnlock()
    if module == nil {
        return nil, baseplugin.ErrModuleNotFound
    }

    // 3. 调用处理方法
    return module.Handle(ctx, route.Method, params)
}
