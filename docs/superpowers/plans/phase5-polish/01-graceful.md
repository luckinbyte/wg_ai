# Task 25: Graceful Shutdown - 优雅关闭

## 背景

实现优雅关闭，确保服务停止时完成正在处理的请求。

## 步骤

### Step 1: Update game main

修改 `cmd/game/main.go`，添加超时关闭：

```go
import (
	"context"
	"time"
	// ... existing imports
)

func main() {
	// ... existing code ...

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	logger.Log.Info("Shutting down...")

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	done := make(chan struct{})
	go func() {
		srv.Stop()
		close(done)
	}()

	select {
	case <-done:
		logger.Log.Info("Server stopped gracefully")
	case <-ctx.Done():
		logger.Log.Warn("Shutdown timeout, forcing exit")
	}
}
```

### Step 2: Apply to login and db servers

同样更新 `cmd/login/main.go` 和 `cmd/db/main.go`。

### Step 3: Commit

```bash
git add .
git commit -m "feat: add graceful shutdown with timeout"
```

## 完成标志

- [ ] 三个服务都支持优雅关闭
- [ ] 测试手动验证
