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
