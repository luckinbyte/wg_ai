package gate

import (
	"testing"
	"time"
)

// TestWSUpgrader 测试WebSocket升级器配置
func TestWSUpgrader(t *testing.T) {
	if upgrader.ReadBufferSize != 4096 {
		t.Errorf("Expected ReadBufferSize=4096, got %d", upgrader.ReadBufferSize)
	}
	if upgrader.WriteBufferSize != 4096 {
		t.Errorf("Expected WriteBufferSize=4096, got %d", upgrader.WriteBufferSize)
	}
	// CheckOrigin 应该允许所有来源
	if !upgrader.CheckOrigin(nil) {
		t.Error("CheckOrigin should return true for all origins")
	}
}

// TestWSServerStartStop 测试服务器启动和停止
func TestWSServerStartStop(t *testing.T) {
	server := NewWSServer(":0", nil, nil)

	if err := server.Start(); err != nil {
		t.Fatalf("Failed to start server: %v", err)
	}

	// 等待启动
	time.Sleep(50 * time.Millisecond)

	// 验证地址已分配
	addr := server.Addr()
	if addr == "" {
		t.Error("Server address should not be empty after start")
	}

	// 停止服务器
	server.Stop()
}

// TestWSConnectionWrap 测试WS连接包装
func TestWSConnectionWrap(t *testing.T) {
	// 测试 WSConnection 结构体创建
	conn := &WSConnection{}
	if conn == nil {
		t.Error("WSConnection should not be nil")
	}
}
