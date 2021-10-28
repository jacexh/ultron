package main

import (
	"time"

	"github.com/wosai/ultron/v2"
	"github.com/wosai/ultron/v2/log"
	"go.uber.org/zap"
)

func main() {
	runner := ultron.NewMasterRunner()
	go func() {
		<-time.After(2 * time.Second)
		plan := ultron.NewPlan("foobar")
		plan.AddStages(&ultron.V1StageConfig{ConcurrentUsers: 100})
		if err := runner.StartPlan(plan); err != nil {
			log.Error("failed to start a new plan", zap.Error(err))
		}
	}()
	runner.Launch(ultron.RunnerConfig{})
}
