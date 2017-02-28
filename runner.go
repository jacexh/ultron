package ultron

import (
	"context"
	"fmt"
	"os"
	"sync"
	"time"
)

type runner struct {
	currentWorkers int
	task           *TaskSet
	statsCollector *statsCollector
	ctx            context.Context
	wg             *sync.WaitGroup
	lock           sync.RWMutex
}

// CoreRunner 核心执行器
var CoreRunner *runner

func newRunner(c *statsCollector) *runner {
	return &runner{
		statsCollector: newStatsCollector(),
		// ctx:            context.Background(),
		wg: &sync.WaitGroup{},
	}
}

func (r *runner) WithTaskSet(t *TaskSet) *runner {
	r.task = t
	return r
}

func (r *runner) Run() {
	go r.statsCollector.Receiving()

	if r.task.OnStart != nil {
		if err := r.task.OnStart(); err != nil {
			panic(err)
		}
	}

	go func() {
		for {
			time.Sleep(time.Second * 5)
			for _, v := range r.statsCollector.entries {
				fmt.Println(v.Report(false))
			}
		}
	}()

	for i := 0; i < r.task.Concurrency; i++ {
		r.wg.Add(1)
		r.currentWorkers++
		go r.attack()
	}

	r.wg.Wait()

	for _, v := range r.statsCollector.entries {
		fmt.Println(v.Report(true))
	}

	os.Exit(0)
}

func (r *runner) attack() {
	defer func() { r.currentWorkers-- }()
	defer r.wg.Done()
	defer func() {
		if rec := recover(); rec != nil {
			// Todo:
			Logger.Error("recoverd")
		}
	}()

	for {
		q := r.task.PickUp()
		start := time.Now()
		err := q.Fire()
		duration := time.Since(start)
		r.statsCollector.receiver <- &QueryResult{Name: q.Name(), Duration: duration, Error: err}

		wait := r.task.Wait()
		if wait != ZeroDuration {
			time.Sleep(wait)
		}
	}
}

// Worker return current worker counts
func (r *runner) Worker() int {
	return r.currentWorkers
}

func init() {
	CoreRunner = newRunner(newStatsCollector())
}
