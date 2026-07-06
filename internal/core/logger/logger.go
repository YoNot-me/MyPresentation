package logger

import (
	"os"

	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

func InitLogger(logFile *os.File) *zap.Logger {
	cfg := zap.NewDevelopmentEncoderConfig()
	cfg.EncodeTime = zapcore.ISO8601TimeEncoder
	cfg.EncodeCaller = zapcore.ShortCallerEncoder
	cfg.EncodeLevel = zapcore.CapitalColorLevelEncoder

	encoder := zapcore.NewConsoleEncoder(cfg)
	consoleWriter := zapcore.AddSync(os.Stdout)
	fileWriter := zapcore.AddSync(logFile)
	multiWriter := zapcore.NewMultiWriteSyncer(
		consoleWriter,
		fileWriter,
	)

	core := zapcore.NewCore(
		encoder,
		multiWriter,
		zap.DebugLevel,
	)

	return zap.New(
		core,
		zap.AddCaller(),
		zap.AddStacktrace(zapcore.PanicLevel),
	)
}
