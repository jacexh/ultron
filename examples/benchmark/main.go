package main

import (
	"context"
	"time"

	"github.com/wosai/ultron/v2"
)

type benchmarkAttacker struct{}

func (a benchmarkAttacker) Name() string {
	return "benchmark"
}

func (a benchmarkAttacker) Fire(ctx context.Context) error {
	time.Sleep(10 * time.Millisecond)
	return nil
}

func main() {
	runner := ultron.BuildLocalRunner()
	go runner.Launch(ultron.RunnerConfig{})

	task := ultron.NewTask()
	task.Add(benchmarkAttacker{}, 1)
	runner.Assign(task)

	plan := ultron.NewPlan("benchmark")
	plan.AddStages(&ultron.V1StageConfig{ConcurrentUsers: 100})
	runner.StartPlan(plan)
}
