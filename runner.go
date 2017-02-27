package ultron

import (
	"context"
	"fmt"
	"os"
	"sync"
	"time"
)

type runner struct {
	task           *TaskSet
	statsCollector *StatsCollector
	ctx            context.Context
	wg             *sync.WaitGroup
	lock           *sync.RWMutex
}

// CoreRunner 核心执行器
var CoreRunner *runner

func newRunner(c *StatsCollector) *runner {
	return &runner{
		statsCollector: NewStatsCollector(),
		// ctx:            context.Background(),
		wg:   &sync.WaitGroup{},
		lock: &sync.RWMutex{},
	}
}

func (r *runner) WithTaskSet(t *TaskSet) *runner {
	r.task = t
	return r
}

func (r *runner) Run() {
	go r.statsCollector.Receiving()

	if err := r.task.OnStart(); err != nil {
		panic(err)
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
		go r.attack()
	}

	r.wg.Wait()

	for _, v := range r.statsCollector.entries {
		v.Report(true)
	}

	os.Exit(0)
}

func (r *runner) attack() {
	defer r.wg.Add(-1)
	defer func() {
		if rec := recover(); rec != nil {
			Logger.Error("recoverd")
		}
	}()

	for {
		q := r.task.Choice()
		start := time.Now()
		duration, err := q.Fire()
		taskDuraton := time.Since(start)
		if duration == ZeroDuration { // 用户并未自定义请求时间
			duration = taskDuraton
		}
		r.statsCollector.receiver <- &QueryResult{Name: q.Name(), Duration: duration, Error: err}

		wait := r.task.Wait()
		if wait != ZeroDuration {
			time.Sleep(wait)
		}
	}
}

func init() {
	CoreRunner = newRunner(NewStatsCollector())
}
