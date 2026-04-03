package plugin

import (
    "path/filepath"
    "strings"

    "github.com/fsnotify/fsnotify"
)

// Watcher 文件监听器
type Watcher struct {
    pluginMgr *Manager
    watcher   *fsnotify.Watcher
    pluginDir string
    stopCh    chan struct{}
}

// NewWatcher 创建文件监听器
func NewWatcher(pluginMgr *Manager, pluginDir string) (*Watcher, error) {
    w, err := fsnotify.NewWatcher()
    if err != nil {
        return nil, err
    }

    return &Watcher{
        pluginMgr: pluginMgr,
        watcher:   w,
        pluginDir: pluginDir,
        stopCh:    make(chan struct{}),
    }, nil
}

// Start 开始监听
func (w *Watcher) Start() error {
    // 添加监听目录
    if err := w.watcher.Add(w.pluginDir); err != nil {
        return err
    }

    go w.run()
    return nil
}

// Stop 停止监听
func (w *Watcher) Stop() {
    close(w.stopCh)
    w.watcher.Close()
}

func (w *Watcher) run() {
    for {
        select {
        case event, ok := <-w.watcher.Events:
            if !ok {
                return
            }

            // 检测 .so 文件写入/创建事件
            if (event.Has(fsnotify.Write) || event.Has(fsnotify.Create)) &&
                strings.HasSuffix(event.Name, ".so") {
                w.handleSoChange(event.Name)
            }

        case err, ok := <-w.watcher.Errors:
            if !ok {
                return
            }
            // 记录错误日志
            _ = err

        case <-w.stopCh:
            return
        }
    }
}

func (w *Watcher) handleSoChange(path string) {
    // 从文件名提取模块名
    module := extractModule(path)

    // 执行热更
    if err := w.pluginMgr.HotReload(module, path); err != nil {
        // 记录错误日志
        return
    }

    // 记录成功日志
}

// extractModule 从路径提取模块名
// ./plugins/role.so -> role
func extractModule(path string) string {
    base := filepath.Base(path)
    return strings.TrimSuffix(base, ".so")
}
