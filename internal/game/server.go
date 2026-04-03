package game

import (
    "net/http"

    "github.com/yourorg/wg_ai/internal/admin"
    "github.com/yourorg/wg_ai/internal/agent"
    "github.com/yourorg/wg_ai/internal/common/config"
    "github.com/yourorg/wg_ai/internal/data"
    "github.com/yourorg/wg_ai/internal/gate"
    "github.com/yourorg/wg_ai/internal/plugin"
    "github.com/yourorg/wg_ai/internal/rpc"
    "github.com/yourorg/wg_ai/internal/session"
)

type Server struct {
    config     *config.GameConfig
    tcpServer  *gate.TCPServer
    agentMgr   *agent.Manager
    sessionMgr *session.Manager
    rpcClient  *rpc.Client

    // 新增
    dataStore  *data.PlayerStore
    pluginMgr  *plugin.Manager
    watcher    *plugin.Watcher
    adminSrv   *http.Server
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

    // 4. 创建 Agent Manager
    s.agentMgr = agent.NewManager(
        s.config.Agent.Count,
        s.config.Gate.MsgQueueSize,
    )

    // 5. 连接 RPC
    s.rpcClient = rpc.NewClient(&rpc.ClientConfig{
        DBAddr:    s.config.Cluster.DBAddr,
        LoginAddr: s.config.Cluster.LoginAddr,
    })
    if err := s.rpcClient.ConnectDB(s.config.Cluster.DBAddr); err != nil {
        return err
    }

    // 6. 启动 TCP 服务
    addr := s.config.Server.Addr()
    s.tcpServer = gate.NewTCPServer(addr, s.sessionMgr, s.agentMgr)

    go func() {
        if err := s.tcpServer.Start(); err != nil {
            panic(err)
        }
    }()

    // 7. 启动管理接口
    s.startAdminServer()

    // 8. 启动文件监听 (可选)
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

    // 关闭 TCP 服务
    if s.tcpServer != nil {
        s.tcpServer.Stop()
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
