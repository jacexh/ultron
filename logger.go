package ultron

import (
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

var (
	Logger *zap.Logger
)

func buildLogger() {
	cfg := zap.NewProductionConfig()
	// cfg.Encoding = "console"
	cfg.EncoderConfig.TimeKey = "@timestamp"
	cfg.EncoderConfig.EncodeTime = zapcore.ISO8601TimeEncoder

	var err error
	Logger, err = cfg.Build()
	if err != nil {
		panic(err)
	}
}

func init() {
	buildLogger()
}
