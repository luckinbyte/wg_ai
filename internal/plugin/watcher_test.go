package plugin

import (
    "os"
    "path/filepath"
    "testing"
    "time"
)

func TestWatcherStart(t *testing.T) {
    // 创建临时目录
    tmpDir, err := os.MkdirTemp("", "plugin_test")
    if err != nil {
        t.Fatal(err)
    }
    defer os.RemoveAll(tmpDir)

    // 创建空的 .so 文件
    soPath := filepath.Join(tmpDir, "test.so")
    if err := os.WriteFile(soPath, []byte("fake"), 0644); err != nil {
        t.Fatal(err)
    }

    // 创建 watcher
    mgr := NewManager()
    watcher, err := NewWatcher(mgr, tmpDir)
    if err != nil {
        t.Fatal(err)
    }
    defer watcher.Stop()

    if err := watcher.Start(); err != nil {
        t.Fatal(err)
    }
}

func TestWatcherExtractModule(t *testing.T) {
    tests := []struct {
        path     string
        expected string
    }{
        {"./plugins/role.so", "role"},
        {"/home/user/plugins/item.so", "item"},
        {"plugins/test_v2.so", "test_v2"},
    }

    for _, tt := range tests {
        result := extractModule(tt.path)
        if result != tt.expected {
            t.Errorf("extractModule(%s) = %s, want %s", tt.path, result, tt.expected)
        }
    }
}

func TestWatcherStop(t *testing.T) {
    tmpDir, err := os.MkdirTemp("", "plugin_test")
    if err != nil {
        t.Fatal(err)
    }
    defer os.RemoveAll(tmpDir)

    mgr := NewManager()
    watcher, err := NewWatcher(mgr, tmpDir)
    if err != nil {
        t.Fatal(err)
    }

    if err := watcher.Start(); err != nil {
        t.Fatal(err)
    }

    // 停止监听
    watcher.Stop()

    // 给一点时间让 goroutine 退出
    time.Sleep(10 * time.Millisecond)
}
