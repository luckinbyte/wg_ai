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
