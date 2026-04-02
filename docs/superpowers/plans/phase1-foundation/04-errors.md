# Task 4: Error System - 错误系统

## 背景与目标

建立统一的错误码体系，用于服务内部和跨服务的错误传递。

**为什么需要这个任务：**
- 游戏服务器需要向客户端返回明确的错误码
- 统一的错误定义方便前后端对接
- 错误码按模块划分，易于扩展

**输出：**
- `internal/common/errors/errors.go` - 错误码定义和 GameError 类型

## 依赖

- Task 1: 项目结构已创建

## 步骤

### Step 1: Write the failing test

Create `internal/common/errors/errors_test.go`:

```go
package errors

import (
	"testing"
)

func TestGameError(t *testing.T) {
	err := NewGameError(CodeInvalidToken, "token is invalid")
	if err.Code != CodeInvalidToken {
		t.Errorf("expected code %d, got %d", CodeInvalidToken, err.Code)
	}
	if err.Error() != "[100] token is invalid" {
		t.Errorf("unexpected error string: %s", err.Error())
	}
}

func TestErrorCodeRanges(t *testing.T) {
	// 测试错误码范围
	if CodeSuccess != 0 {
		t.Error("CodeSuccess should be 0")
	}
	if CodeInvalidToken < 100 || CodeInvalidToken >= 200 {
		t.Error("Login errors should be in 100-199 range")
	}
	if CodeNotEnoughResource < 200 || CodeNotEnoughResource >= 300 {
		t.Error("Game errors should be in 200-299 range")
	}
}
```

### Step 2: Run test to verify it fails

```bash
cd /root/ai_project/wg_ai
go test ./internal/common/errors/...
```

Expected: FAIL - undefined: NewGameError, CodeInvalidToken

### Step 3: Write minimal implementation

Create `internal/common/errors/errors.go`:

```go
package errors

import "fmt"

// 通用错误 0-99
const (
	CodeSuccess      = 0
	CodeUnknown      = 1
	CodeInvalidParam = 2
	CodeTimeout      = 3
)

// 登录相关错误 100-199
const (
	CodeInvalidToken    = 100
	CodeTokenExpired    = 101
	CodeAccountNotFound = 102
	CodePasswordWrong   = 103
)

// 游戏逻辑错误 200-299
const (
	CodeNotEnoughResource = 200
	CodeInvalidOperation  = 201
	CodePlayerNotFound    = 202
)

type GameError struct {
	Code    int
	Message string
}

func (e *GameError) Error() string {
	return fmt.Sprintf("[%d] %s", e.Code, e.Message)
}

func NewGameError(code int, msg string) *GameError {
	return &GameError{Code: code, Message: msg}
}

func IsGameError(err error) bool {
	_, ok := err.(*GameError)
	return ok
}

func GetCode(err error) int {
	if ge, ok := err.(*GameError); ok {
		return ge.Code
	}
	return CodeUnknown
}
```

### Step 4: Run test to verify it passes

```bash
cd /root/ai_project/wg_ai
go test ./internal/common/errors/...
```

Expected: PASS

### Step 5: Commit

```bash
git add .
git commit -m "feat: add game error system with error codes"
```

## 验证

```bash
cd /root/ai_project/wg_ai
go test ./internal/common/errors/... -v
```

Expected: PASS

## 完成标志

- [ ] 测试通过
- [ ] errors.go 包含所有错误码常量
- [ ] GameError 类型实现 error 接口
- [ ] Commit 完成
