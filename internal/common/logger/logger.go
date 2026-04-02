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
