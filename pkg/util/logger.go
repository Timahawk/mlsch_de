package util

import (
	"os"

	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

func InitLogger() *zap.Logger {
	// Stolen from
	// https://codewithmukesh.com/blog/structured-logging-in-golang-with-zap/
	config := zap.NewProductionEncoderConfig()
	configFile := zap.NewProductionEncoderConfig()
	// config := zap.NewDevelopmentEncoderConfig()
	config.EncodeTime = zapcore.RFC3339TimeEncoder
	configFile.EncodeTime = zapcore.RFC3339TimeEncoder
	config.EncodeLevel = zapcore.CapitalColorLevelEncoder
	fileEncoder := zapcore.NewJSONEncoder(configFile)
	consoleEncoder := zapcore.NewConsoleEncoder(config)
	logFile, _ := os.OpenFile("Logs.log", os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	writer := zapcore.AddSync(logFile)
	LogLevelFile := zapcore.DebugLevel
	LogLevelStd := zapcore.InfoLevel
	core := zapcore.NewTee(
		zapcore.NewCore(fileEncoder, writer, LogLevelFile),
		zapcore.NewCore(consoleEncoder, zapcore.AddSync(os.Stdout), LogLevelStd),
	)
	logger := zap.New(core, zap.AddCaller(), zap.AddStacktrace(zapcore.ErrorLevel))
	Sugar = logger.Sugar()
	return logger

}
