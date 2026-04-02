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
