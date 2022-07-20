package logger

import (
	"os"

	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

func getLoggerTee(logFile *os.File, config zapcore.EncoderConfig, defaultLogLevel zapcore.Level) (core zapcore.Core) {
	consoleEncoder := zapcore.NewConsoleEncoder(config)

	if logFile == nil {
		core = zapcore.NewTee(
			zapcore.NewCore(consoleEncoder, zapcore.AddSync(os.Stdout), defaultLogLevel),
		)
		return
	}

	fileEncoderJSON := zapcore.NewJSONEncoder(config)
	writer := zapcore.AddSync(logFile)
	core = zapcore.NewTee(
		zapcore.NewCore(fileEncoderJSON, writer, defaultLogLevel),
		zapcore.NewCore(consoleEncoder, zapcore.AddSync(os.Stdout), defaultLogLevel),
	)
	return
}

func InitializeLogger(logFile *os.File, defaultLogLevel zapcore.Level) (logger *zap.Logger) {
	config := zap.NewProductionEncoderConfig()
	config.EncodeTime = zapcore.ISO8601TimeEncoder

	core := getLoggerTee(logFile, config, defaultLogLevel)
	logger = zap.New(core, zap.AddCaller(), zap.AddStacktrace(zapcore.ErrorLevel))

	return
}
