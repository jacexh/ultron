package main

import (
	"context"
	"time"

	"github.com/wosai/ultron/v2"
)

type benchmarkAttacker struct {
	name string
}

func (b *benchmarkAttacker) Name() string {
	return b.name
}

func (b *benchmarkAttacker) Fire(_ context.Context) error {
	time.Sleep(10 * time.Millisecond)
	return nil
}

func main() {
	runner := ultron.NewLocalRunner()
	task := ultron.NewTask()
	task.Add(&benchmarkAttacker{name: "benchmark"}, 1)
	runner.Assign(task)

	go func() {
		plan := ultron.NewPlan("benchmark test")
		plan.AddStages(&ultron.V1StageConfig{ConcurrentUsers: 200, Duration: 3 * time.Minute})
		<-time.After(4 * time.Second)
		runner.StartPlan(plan)
	}()

	runner.Launch()
}
