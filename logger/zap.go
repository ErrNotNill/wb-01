package logger

import (
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
	"log"
)

func InitLogger() *zap.Logger {
	var logger *zap.Logger
	var err error

	zapConfig := zap.NewProductionConfig()
	zapConfig.EncoderConfig.TimeKey = "timestamp"
	zapConfig.EncoderConfig.EncodeTime = zapcore.ISO8601TimeEncoder

	if logger, err = zapConfig.Build(); err != nil {
		log.Fatal("Error building zap logger:", err.Error())
	}
	zap.ReplaceGlobals(logger)
	return logger
}
