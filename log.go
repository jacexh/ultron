package ultron

import (
	"sync"

	"github.com/jacexh/gopkg/zaprotate"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

var (
	// Logger 全局日志
	Logger *zap.Logger

	levelMapper = map[string]zapcore.Level{
		"info":  zapcore.InfoLevel,
		"debug": zapcore.DebugLevel,
		"warn":  zapcore.WarnLevel,
		"error": zapcore.ErrorLevel,
	}

	once sync.Once
)

func buildLogger(opt LoggerOption) {
	once.Do(func() {
		cfg := zap.NewProductionConfig()
		cfg.EncoderConfig.TimeKey = "@timestamp"
		cfg.EncoderConfig.EncodeTime = zapcore.ISO8601TimeEncoder
		cfg.Sampling = nil
		cfg.Level = zap.NewAtomicLevelAt(levelMapper[opt.Level])

		Logger = zaprotate.BuildRotateLogger(cfg, zaprotate.RotatingFileConfig{
			LoggerName: "",
			Filename:   opt.FileName,
			MaxSize:    opt.MaxSize,
			MaxBackups: opt.MaxBackups,
		})
	})
}
