# Task 3: Logger System - 日志系统

## 背景与目标

建立统一的日志系统，使用 Zap 库提供高性能结构化日志。

**为什么需要这个任务：**
- 所有服务都需要日志记录
- Zap 是 Go 中性能最高的日志库之一
- 统一的日志格式方便问题排查

**输出：**
- `internal/common/logger/logger.go` - 日志封装

## 依赖

- Task 1: 项目结构已创建

## 步骤

### Step 1: Write the failing test

Create `internal/common/logger/logger_test.go`:

```go
package logger

import (
	"bytes"
	"strings"
	"testing"
)

func TestLogger(t *testing.T) {
	var buf bytes.Buffer
	l := New(&buf, "debug")
	l.Info("test message")
	if !strings.Contains(buf.String(), "test message") {
		t.Errorf("expected log to contain 'test message', got %s", buf.String())
	}
}

func TestLogLevel(t *testing.T) {
	var buf bytes.Buffer
	l := New(&buf, "error")
	l.Info("should not appear")
	l.Error("should appear")
	
	output := buf.String()
	if strings.Contains(output, "should not appear") {
		t.Error("info should be filtered at error level")
	}
	if !strings.Contains(output, "should appear") {
		t.Error("error should be logged")
	}
}
```

### Step 2: Run test to verify it fails

```bash
cd /root/ai_project/wg_ai
go test ./internal/common/logger/...
```

Expected: FAIL - undefined: New

### Step 3: Write minimal implementation

Create `internal/common/logger/logger.go`:

```go
package logger

import (
	"io"

	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

var Log *zap.SugaredLogger

func New(w io.Writer, level string) *zap.SugaredLogger {
	var lvl zapcore.Level
	if err := lvl.UnmarshalText([]byte(level)); err != nil {
		lvl = zapcore.InfoLevel
	}

	encoder := zapcore.NewConsoleEncoder(zap.NewDevelopmentEncoderConfig())
	core := zapcore.NewCore(encoder, zapcore.AddSync(w), lvl)
	logger := zap.New(core, zap.AddCaller())

	Log = logger.Sugar()
	return Log
}

func Sync() {
	if Log != nil {
		_ = Log.Sync()
	}
}
```

### Step 4: Install dependency and run test

```bash
cd /root/ai_project/wg_ai
go get go.uber.org/zap
go test ./internal/common/logger/...
```

Expected: PASS

### Step 5: Commit

```bash
git add .
git commit -m "feat: add zap logger system"
```

## 验证

```bash
cd /root/ai_project/wg_ai
go test ./internal/common/logger/... -v
```

Expected: PASS

## 完成标志

- [ ] 测试通过
- [ ] logger.go 包含 New 函数和全局 Log 变量
- [ ] Commit 完成
