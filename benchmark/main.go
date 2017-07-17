package main

import (
	"time"

	"github.com/jacexh/ultron"
)

type BenchmarkAttacher struct{}

func (b *BenchmarkAttacher) Name() string {
	return "benchmark"
}

func (b *BenchmarkAttacher) Fire() (int, error) {
	time.Sleep(time.Millisecond * 10)
	return 0, nil
}

func main() {
	attacker := &BenchmarkAttacher{}

	tasks := ultron.NewTaskSet()
	tasks.Add(attacker, 1)
	tasks.MinWait = ultron.ZeroDuration
	tasks.MaxWait = ultron.ZeroDuration

	ultron.Runner.Config.Concurrence = 10000

	ultron.Runner.Run(tasks)
}
