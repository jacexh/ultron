package main

import (
	"time"

	"github.com/wosai/ultron/v2"
	"github.com/wosai/ultron/v2/log"
	"go.uber.org/zap"
)

func main() {
	runner := ultron.NewMasterRunner()
	go runner.Launch(ultron.RunnerConfig{})

	<-time.After(3 * time.Second)
	plan := ultron.NewPlan("foobar")
	plan.AddStages(&ultron.V1StageConfig{})
	if err := runner.StartPlan(plan); err != nil {
		log.Error("failed", zap.Error(err))
	}
}
