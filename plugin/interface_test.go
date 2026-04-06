package plugin_test

import (
    "testing"

    "github.com/luckinbyte/wg_ai/plugin"
)

func TestLogicResult(t *testing.T) {
    // Test Success
    result := plugin.Success(map[string]any{"key": "value"})
    if result.Code != 0 {
        t.Errorf("expected code 0, got %d", result.Code)
    }
    if result.Data["key"] != "value" {
        t.Error("data mismatch")
    }

    // Test Error
    errResult := plugin.Error(100, "test error")
    if errResult.Code != 100 {
        t.Errorf("expected code 100, got %d", errResult.Code)
    }
    if errResult.Message != "test error" {
        t.Error("message mismatch")
    }
}

func TestLogicResultWithPush(t *testing.T) {
    result := plugin.Success(nil).
        WithPush(1001, []byte("data1")).
        WithPush(1002, []byte("data2"))

    if len(result.Push) != 2 {
        t.Errorf("expected 2 pushes, got %d", len(result.Push))
    }
    if result.Push[0].MsgID != 1001 {
        t.Error("push msg_id mismatch")
    }
}

func TestErrors(t *testing.T) {
    if plugin.ErrModuleNotFound.Error() != "module not found" {
        t.Error("error message mismatch")
    }
    if plugin.ErrMethodNotFound.Error() != "method not found" {
        t.Error("error message mismatch")
    }
}
