package game

import (
    "fmt"
    "net/http"

    "github.com/luckinbyte/wg_ai/internal/admin"
    "github.com/luckinbyte/wg_ai/internal/agent"
    "github.com/luckinbyte/wg_ai/internal/common/config"
    "github.com/luckinbyte/wg_ai/internal/data"
    "github.com/luckinbyte/wg_ai/internal/gate"
    "github.com/luckinbyte/wg_ai/internal/march"
    "github.com/luckinbyte/wg_ai/internal/plugin"
    "github.com/luckinbyte/wg_ai/internal/rpc"
    "github.com/luckinbyte/wg_ai/internal/scene"
    "github.com/luckinbyte/wg_ai/internal/session"
)

type Server struct {
    config     *config.GameConfig
    tcpServer  *gate.TCPServer
    wsServer   *gate.WSServer    // WebSocket服务器
    agentMgr   *agent.Manager
    sessionMgr *session.Manager
    rpcClient  *rpc.Client

    // 数据层
    dataStore  *data.PlayerStore
    pluginMgr  *plugin.Manager
    watcher    *plugin.Watcher
    adminSrv   *http.Server

    // 场景系统
    sceneMgr   *scene.Manager

    // 行军系统
    marchMgr   *march.Manager
}

func NewServer(cfg *config.GameConfig) *Server {
    return &Server{
        config:     cfg,
        sessionMgr: session.NewManager(),
    }
}

func (s *Server) Start() error {
    // 1. 初始化数据层
    s.dataStore = data.NewPlayerStore()

    // 2. 初始化插件管理器
    s.pluginMgr = plugin.NewManager()

    // 3. 加载路由配置
    if s.config.Plugin.RouteFile != "" {
        if err := s.pluginMgr.Router().LoadFromConfig(s.config.Plugin.RouteFile); err != nil {
            return err
        }
    }

    // 4. 初始化场景管理器
    s.sceneMgr = scene.NewManager()

    // 4.1 注册场景模块 (内置模块)
    sceneModule := scene.NewModule(s.sceneMgr)
    s.pluginMgr.RegisterModule("scene", sceneModule)

    // 4.2 初始化行军管理器
    s.marchMgr = march.NewManager(s.sceneMgr)

    // 4.3 注册行军模块 (内置模块)
    marchModule := march.NewModule(s.marchMgr)
    s.pluginMgr.RegisterModule("march", marchModule)

    // 4.4 启动行军管理器
    s.marchMgr.Start()

    // 5. 创建 Agent Manager
    s.agentMgr = agent.NewManager(
        s.config.Agent.Count,
        s.config.Gate.MsgQueueSize,
    )

    // 6. 连接 RPC
    s.rpcClient = rpc.NewClient(&rpc.ClientConfig{
        DBAddr:    s.config.Cluster.DBAddr,
        LoginAddr: s.config.Cluster.LoginAddr,
    })
    if err := s.rpcClient.ConnectDB(s.config.Cluster.DBAddr); err != nil {
        return err
    }

    // 7. 启动 TCP 服务
    addr := s.config.Server.Addr()
    s.tcpServer = gate.NewTCPServer(addr, s.sessionMgr, s.agentMgr)

    go func() {
        if err := s.tcpServer.Start(); err != nil {
            panic(err)
        }
    }()

    // 7.1 启动 WebSocket 服务
    if s.config.Gate.WSPort > 0 {
        wsAddr := fmt.Sprintf(":%d", s.config.Gate.WSPort)
        s.wsServer = gate.NewWSServer(wsAddr, s.sessionMgr, s.agentMgr)
        if err := s.wsServer.Start(); err != nil {
            return err
        }
    }

    // 8. 启动管理接口
    s.startAdminServer()

    // 9. 启动文件监听 (可选)
    if s.config.Plugin.Watch && s.config.Plugin.Dir != "" {
        s.startWatcher()
    }

    return nil
}

func (s *Server) startWatcher() {
    watcher, err := plugin.NewWatcher(s.pluginMgr, s.config.Plugin.Dir)
    if err != nil {
        return
    }
    s.watcher = watcher
    watcher.Start()
}

func (s *Server) startAdminServer() {
    mux := http.NewServeMux()
    handler := admin.NewHandler(s.pluginMgr)
    handler.RegisterRoutes(mux)

    s.adminSrv = &http.Server{
        Addr:    s.config.Admin.Addr,
        Handler: mux,
    }

    go s.adminSrv.ListenAndServe()
}

func (s *Server) Stop() {
    // 关闭管理接口
    if s.adminSrv != nil {
        s.adminSrv.Close()
    }

    // 停止文件监听
    if s.watcher != nil {
        s.watcher.Stop()
    }

    // 停止行军管理器
    if s.marchMgr != nil {
        s.marchMgr.Stop()
    }

    // 关闭 TCP 服务
    if s.tcpServer != nil {
        s.tcpServer.Stop()
    }

    // 关闭 WebSocket 服务
    if s.wsServer != nil {
        s.wsServer.Stop()
    }

    // 关闭 Agent
    if s.agentMgr != nil {
        s.agentMgr.Stop()
    }

    // 关闭 RPC
    if s.rpcClient != nil {
        s.rpcClient.Close()
    }
}

