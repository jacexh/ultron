package ultron

import (
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

// Logger global logger
var Logger *zap.Logger

func init() {
	cfg := zap.NewDevelopmentConfig()
	cfg.DisableCaller = true
	cfg.Level.SetLevel(zapcore.InfoLevel)
	Logger, _ = cfg.Build()
}
