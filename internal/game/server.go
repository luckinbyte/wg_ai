package game

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
	"strings"
	"time"

	"github.com/luckinbyte/wg_ai/internal/admin"
	"github.com/luckinbyte/wg_ai/internal/agent"
	"github.com/luckinbyte/wg_ai/internal/common/config"
	"github.com/luckinbyte/wg_ai/internal/common/logger"
	"github.com/luckinbyte/wg_ai/internal/data"
	"github.com/luckinbyte/wg_ai/internal/gate"
	"github.com/luckinbyte/wg_ai/internal/march"
	"github.com/luckinbyte/wg_ai/internal/plugin"
	"github.com/luckinbyte/wg_ai/internal/rpc"
	"github.com/luckinbyte/wg_ai/internal/scene"
	"github.com/luckinbyte/wg_ai/internal/session"
	baseplugin "github.com/luckinbyte/wg_ai/plugin"
	cityplugin "github.com/luckinbyte/wg_ai/plugin/city"
	soldierplugin "github.com/luckinbyte/wg_ai/plugin/soldier"
)

type Server struct {
	config     *config.GameConfig
	tcpServer  *gate.TCPServer
	wsServer   *gate.WSServer // WebSocket服务器
	agentMgr   *agent.Manager
	sessionMgr *session.Manager
	rpcClient  *rpc.Client

	// 数据层
	dataStore *data.PlayerStore
	pluginMgr *plugin.Manager
	watcher   *plugin.Watcher
	adminSrv  *http.Server

	// 场景系统
	sceneMgr *scene.Manager

	// 行军系统
	marchMgr *march.Manager
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

	// 4. 初始化场景管理器
	s.sceneMgr = scene.NewManager()

	// 4.1 注册场景模块 (内置模块)
	sceneModule := scene.NewModule(s.sceneMgr)
	s.pluginMgr.RegisterModule("scene", sceneModule)

	// 4.2 注册城池模块 (内置模块)
	cityModule, err := cityplugin.NewModule(s.sceneMgr)
	if err != nil {
		return err
	}
	s.pluginMgr.RegisterModule("city", cityModule)

	// 4.3 初始化行军管理器
	s.marchMgr = march.NewManager(s.sceneMgr)

	// 4.4 注册行军模块 (内置模块)
	marchModule := march.NewModule(s.marchMgr)
	s.pluginMgr.RegisterModule("march", marchModule)

	// 4.4.1 注册角色模块 (内置模块)
	roleModule := &roleModule{}
	s.pluginMgr.RegisterModule("role", roleModule)

	// 4.4.2 注册士兵模块 (内置模块)
	soldierModule := soldierplugin.GetSoldierModule()
	s.pluginMgr.RegisterModule("soldier", soldierModule)

	// 4.4.3 设置行军模块的士兵消费者
	soldierMgr := soldierplugin.GetSoldierManager()
	if soldierMgr != nil {
		s.marchMgr.SetSoldierConsumer(soldierMgr)
		log.Printf("[Init] soldier consumer registered")
	}

	// 4.4.4 扫描并加载 plugins/ 目录下的 .so 插件 (热更用)
	if err := s.loadPlugins(s.config.Plugin.Dir); err != nil {
		log.Printf("[Init] plugin scan warning: %v", err)
	}

	// 4.5 启动行军管理器
	s.marchMgr.Start()

	// 3. 加载路由配置
	if s.config.Plugin.RouteFile != "" {
		if err := s.pluginMgr.Router().LoadFromConfig(s.config.Plugin.RouteFile); err != nil {
			return err
		}
	}

	// 5. 创建 Agent Manager
	s.agentMgr = agent.NewManager(
		s.config.Agent.Count,
		s.config.Gate.MsgQueueSize,
	)

	// 5.1 设置 fallback: 将未注册的消息转发给 PluginManager
	s.agentMgr.SetFallback(func(a *agent.Agent, msg *agent.Message) ([]byte, error) {
		log.Printf("[Fallback] processing msgID=%d from RID=%d", msg.MsgID, msg.Sess.RID)

		// 检查路由
		route, ok := s.pluginMgr.Router().Get(msg.MsgID)
		if !ok {
			log.Printf("[Fallback] route not found for msgID=%d", msg.MsgID)
			errResp, _ := json.Marshal(&baseplugin.LogicResult{Code: 404, Message: "route not found"})
			return gate.EncodePacket(gate.MsgTypeResponse, errResp), nil
		}
		log.Printf("[Fallback] route found: msgID=%d -> module=%s, method=%s", msg.MsgID, route.Module, route.Method)

		// 检查模块
		module := s.pluginMgr.GetModule(route.Module)
		if module == nil {
			log.Printf("[Fallback] module not found: %s", route.Module)
			errResp, _ := json.Marshal(&baseplugin.LogicResult{Code: 404, Message: "module not found"})
			return gate.EncodePacket(gate.MsgTypeResponse, errResp), nil
		}
		log.Printf("[Fallback] module found: %s", route.Module)

		playerData, err := s.dataStore.GetPlayer(msg.Sess.RID)
		if err != nil || playerData == nil {
			log.Printf("[Fallback] player data not found for RID=%d err=%v", msg.Sess.RID, err)
			return nil, nil
		}

		var params map[string]any
		if len(msg.Payload) > 0 {
			if err := json.Unmarshal(msg.Payload, &params); err != nil {
				params = make(map[string]any)
			}
		} else {
			params = make(map[string]any)
		}

		ctx := &baseplugin.LogicContext{
			RID:     msg.Sess.RID,
			UID:     msg.Sess.UID,
			Data:    &playerDataAdapter{PlayerData: playerData},
			Session: &sessionPushAdapter{sess: msg.Sess},
		}

		result, err := s.pluginMgr.Call(ctx, msg.MsgID, params)
		if err != nil {
			log.Printf("[Fallback] plugin call failed: msgID=%d, err=%v", msg.MsgID, err)
			errResp, _ := json.Marshal(&baseplugin.LogicResult{Code: 500, Message: err.Error()})
			return gate.EncodePacket(gate.MsgTypeResponse, errResp), nil
		}

		respData, _ := json.Marshal(result)
		return gate.EncodePacket(gate.MsgTypeResponse, respData), nil
	})

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
	s.logLifecycle("shutdown start")

	s.logLifecycle("phase 1/8 stop admin server start")
	if s.adminSrv != nil {
		if err := s.adminSrv.Close(); err != nil {
			s.logLifecyclef("phase 1/8 stop admin server error: %v", err)
		}
	}
	s.logLifecycle("phase 1/8 stop admin server done")

	s.logLifecycle("phase 2/8 stop watcher start")
	if s.watcher != nil {
		s.watcher.Stop()
	}
	s.logLifecycle("phase 2/8 stop watcher done")

	s.logLifecycle("phase 3/8 stop TCP server start")
	if s.tcpServer != nil {
		s.tcpServer.Stop()
	}
	s.logLifecycle("phase 3/8 stop TCP server done")

	s.logLifecycle("phase 4/8 stop WebSocket server start")
	if s.wsServer != nil {
		s.wsServer.Stop()
	}
	s.logLifecycle("phase 4/8 stop WebSocket server done")

	s.logLifecycle("phase 5/8 stop march manager start")
	if s.marchMgr != nil {
		s.marchMgr.Stop()
	}
	s.logLifecycle("phase 5/8 stop march manager done")

	s.logLifecycle("phase 6/8 stop agent manager start")
	if s.agentMgr != nil {
		s.agentMgr.Stop()
	}
	s.logLifecycle("phase 6/8 stop agent manager done")

	s.logLifecycle("phase 7/8 flush loaded players start")
	s.flushLoadedPlayers()
	s.logLifecycle("phase 7/8 flush loaded players done")

	s.logLifecycle("phase 8/8 close rpc client start")
	if s.rpcClient != nil {
		s.rpcClient.Close()
	}
	s.logLifecycle("phase 8/8 close rpc client done")
	s.logLifecycle("shutdown complete")
}

func (s *Server) flushLoadedPlayers() {
	if s.dataStore == nil || s.rpcClient == nil || !s.rpcClient.HasDBConnection() {
		return
	}

	err := s.dataStore.ForEachLoadedPlayer(func(rid int64, p *data.PlayerData) error {
		dirty, version := p.SnapshotDirtyVersion()
		if !dirty {
			return nil
		}

		payload, err := data.SerializePlayer(p)
		if err != nil {
			s.logLifecyclef("failed to serialize dirty player rid=%d err=%v", rid, err)
			return nil
		}

		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()

		if err := s.rpcClient.SaveRole(ctx, rid, payload); err != nil {
			s.logLifecyclef("failed to save dirty player rid=%d err=%v", rid, err)
			return nil
		}

		if !p.ClearDirtyIfVersion(version) {
			s.logLifecyclef("player changed during flush rid=%d; leaving dirty", rid)
		}
		return nil
	})
	if err != nil {
		s.logLifecyclef("failed to iterate loaded players err=%v", err)
	}
}

func (s *Server) logLifecycle(msg string) {
	s.logLifecyclef("%s", msg)
}

func (s *Server) logLifecyclef(format string, args ...any) {
	if logger.Log != nil {
		logger.Log.Infof(format, args...)
		return
	}
	log.Printf(format, args...)
}

type playerDataAdapter struct {
	*data.PlayerData
}

func (a *playerDataAdapter) GetField(key string) (any, error) {
	return a.PlayerData.GetField(key), nil
}

func (a *playerDataAdapter) SetField(key string, value any) error {
	a.PlayerData.SetField(key, value)
	return nil
}

func (a *playerDataAdapter) GetArray(key string) (any, error) {
	return a.PlayerData.GetArray(key), nil
}

func (a *playerDataAdapter) SetArray(key string, value any) error {
	a.PlayerData.SetArray(key, value)
	return nil
}

type sessionPushAdapter struct {
	sess *session.Session
}

func (s *sessionPushAdapter) Push(msgID uint16, data []byte) error {
	packet := gate.EncodePacket(gate.MsgTypePush, data)
	return s.sess.Send(packet)
}

type roleModule struct{}

func (l *roleModule) Name() string {
	return "role"
}

func (l *roleModule) Handle(ctx *baseplugin.LogicContext, method string, params map[string]any) (*baseplugin.LogicResult, error) {
	switch method {
	case "login":
		return l.handleLogin(ctx, params)
	case "heartbeat":
		return l.handleHeartbeat(ctx, params)
	case "get_info":
		return l.handleGetInfo(ctx, params)
	case "update_name":
		return l.handleUpdateName(ctx, params)
	default:
		return nil, baseplugin.ErrMethodNotFound
	}
}

func (l *roleModule) handleLogin(ctx *baseplugin.LogicContext, params map[string]any) (*baseplugin.LogicResult, error) {
	if cityplugin.HasSceneManager() {
		if _, err := cityplugin.GetCity(ctx.Data); err != nil {
			return baseplugin.Error(500, err.Error()), nil
		}
		if _, err := cityplugin.InitPlayerCity(ctx.Data, ctx.RID); err != nil {
			return baseplugin.Error(500, err.Error()), nil
		}
	}

	name, _ := ctx.Data.GetField("name")
	level, _ := ctx.Data.GetField("level")
	exp, _ := ctx.Data.GetField("exp")

	return baseplugin.Success(map[string]any{
		"rid":   ctx.RID,
		"name":  name,
		"level": level,
		"exp":   exp,
	}), nil
}

func (l *roleModule) handleHeartbeat(ctx *baseplugin.LogicContext, params map[string]any) (*baseplugin.LogicResult, error) {
	return baseplugin.Success(nil), nil
}

func (l *roleModule) handleGetInfo(ctx *baseplugin.LogicContext, params map[string]any) (*baseplugin.LogicResult, error) {
	name, _ := ctx.Data.GetField("name")
	level, _ := ctx.Data.GetField("level")
	exp, _ := ctx.Data.GetField("exp")
	gold, _ := ctx.Data.GetField("gold")
	vip, _ := ctx.Data.GetField("vip")

	return baseplugin.Success(map[string]any{
		"rid":   ctx.RID,
		"name":  name,
		"level": level,
		"exp":   exp,
		"gold":  gold,
		"vip":   vip,
	}), nil
}

func (l *roleModule) handleUpdateName(ctx *baseplugin.LogicContext, params map[string]any) (*baseplugin.LogicResult, error) {
	name, ok := params["name"].(string)
	if !ok || name == "" {
		return baseplugin.Error(2, "invalid name"), nil
	}

	// 更新数据
	if err := ctx.Data.SetField("name", name); err != nil {
		return nil, err
	}

	return baseplugin.Success(map[string]any{"name": name}), nil
}

// loadPlugins 扫描插件目录，加载所有已有的 .so 文件
func (s *Server) loadPlugins(pluginDir string) error {
	entries, err := os.ReadDir(pluginDir)
	if err != nil {
		return fmt.Errorf("read plugin dir %s: %w", pluginDir, err)
	}
	for _, entry := range entries {
		if entry.IsDir() || !strings.HasSuffix(entry.Name(), ".so") {
			continue
		}
		name := strings.TrimSuffix(entry.Name(), ".so")
		soPath := pluginDir + "/" + entry.Name()
		if err := s.pluginMgr.LoadPlugin(name, soPath); err != nil {
			log.Printf("[Init] plugin %s load failed: %v", name, err)
		} else {
			log.Printf("[Init] plugin %s loaded from %s", name, soPath)
		}
	}
	return nil
}
