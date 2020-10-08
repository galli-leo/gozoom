package parser

import (
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

func newLogger(name string) *zap.SugaredLogger {
	config := zap.NewDevelopmentConfig()
	config.EncoderConfig.EncodeLevel = zapcore.CapitalColorLevelEncoder
	config.InitialFields = map[string]interface{}{
		"name": name,
	}
	logger, _ := config.Build()

	return logger.Sugar()
}
