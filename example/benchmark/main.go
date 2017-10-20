package main

import (
	"time"

	"github.com/jacexh/ultron"
)

type (
	benchmark struct{}
)

func (b benchmark) Name() string {
	return "benchmark"
}

func (b benchmark) Fire() error {
	time.Sleep(time.Millisecond * 10)
	return nil
}

func main() {
	task := ultron.NewTask()
	t := benchmark{}
	task.Add(t, 1)

	ultron.LocalEventHook.Concurrency = 0
	ultron.LocalRunner.WithTask(task)
	ultron.LocalRunner.Config.Concurrence = 10000
	ultron.LocalRunner.Config.Requests = 7471348
	ultron.LocalRunner.Config.Duration = time.Minute
	ultron.LocalRunner.Config.MaxWait = ultron.ZeroDuration
	ultron.LocalRunner.Config.MinWait = ultron.ZeroDuration

	ultron.LocalRunner.Start()
}
