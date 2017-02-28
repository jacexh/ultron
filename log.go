package ultron

import (
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

// Logger global logger
var Logger *zap.Logger

func init() {
	cfg := zap.NewProductionConfig()
	cfg.Encoding = "console"
	cfg.EncoderConfig.EncodeTime = zapcore.ISO8601TimeEncoder
	// cfg.DisableCaller = true
	cfg.Level.SetLevel(zapcore.InfoLevel)
	Logger, _ = cfg.Build()
}
