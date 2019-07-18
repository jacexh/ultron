package main

import (
	"time"

	"github.com/qastub/ultron"
)

type (
	benchmark struct {
		name string
	}
)

func (b *benchmark) Name() string {
	return b.name
}

func (b *benchmark) Fire() error {
	time.Sleep(10 * time.Millisecond)
	return nil
}

func main() {
	task := ultron.NewTask()
	task.Add(&benchmark{"test-1"}, 1)
	task.Add(&benchmark{"test-2"}, 2)

	ultron.LocalEventHook.Concurrency = 200
	ultron.LocalRunner.WithTask(task)
	ultron.LocalRunner.Config.AppendStages(
		&ultron.Stage{Duration: 1 * time.Minute, Concurrence: 1000, HatchRate: 200},
		&ultron.Stage{Duration: 1 * time.Minute, Concurrence: 500, HatchRate: 100},
		&ultron.Stage{Duration: 1 * time.Minute, Concurrence: 2000, HatchRate: 300},
	)
	ultron.LocalRunner.Config.MaxWait = ultron.ZeroDuration
	ultron.LocalRunner.Config.MinWait = ultron.ZeroDuration

	ultron.LocalRunner.Start()
}
