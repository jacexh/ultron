package ultron

import (
	"fmt"
	"math/rand"
	"time"

	"github.com/json-iterator/go"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

var (
	json = jsoniter.ConfigCompatibleWithStandardLibrary
	// J .
	J = json
	// Logger 全局日志
	Logger *zap.Logger
)

func init() {
	cfg := zap.NewProductionConfig()
	cfg.Encoding = "console"
	cfg.EncoderConfig.TimeKey = "time"
	cfg.EncoderConfig.EncodeTime = zapcore.ISO8601TimeEncoder

	var err error
	Logger, err = cfg.Build()
	if err != nil {
		fmt.Printf("init Logger failed: %v\n", err)
		panic(err)
	}

	rand.Seed(time.Now().UnixNano())

	LocalEventHook = newEventHook(LocalEventHookConcurrency)
	LocalEventHook.AddReportHandleFunc(printReportToConsole)

	SlaveEventHook = newEventHook(SlaveEventHookConcurrency)

	MasterEventHook = newEventHook(MasterEventHookConcurrency)
	MasterEventHook.AddReportHandleFunc(printReportToConsole)

	LocalRunner = newLocalRunner(newSummaryStats())
	MasterRunner = newMasterRunner(MasterListenAddr, newSummaryStats())
	SlaveRunner = newSlaveRunner()
}
