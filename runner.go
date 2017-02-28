package ultron

import (
	"context"
	"fmt"
	"os"
	"sync"
	"time"
)

type runner struct {
	duration  time.Duration
	workers   int
	task      *TaskSet
	collector *statsCollector
	ctx       context.Context
	wg        *sync.WaitGroup
	lock      sync.RWMutex
}

// CoreRunner 核心执行器
var CoreRunner *runner

func newRunner(c *statsCollector) *runner {
	return &runner{
		collector: newStatsCollector(),
		duration:  ZeroDuration,
		ctx:       context.Background(),
		wg:        &sync.WaitGroup{},
	}
}

func (r *runner) WithTaskSet(t *TaskSet) *runner {
	r.task = t
	return r
}

func (r *runner) Run() {
	Logger.Info("start")
	go r.collector.Receiving()

	if r.task.OnStart != nil {
		Logger.Info("call OnStart()")
		if err := r.task.OnStart(); err != nil {
			panic(err)
		}
	}

	ctx := r.ctx
	cancel := func() {}

	if r.duration > ZeroDuration {
		end := time.Now().Add(r.duration)
		ctx, cancel = context.WithDeadline(r.ctx, end)
	}

	go func() {
		for {
			time.Sleep(time.Second * 5)
			for _, v := range r.collector.entries {
				fmt.Println(v.Report(false))
			}
		}
	}()

	for i := 0; i < r.task.Concurrency; i++ {
		r.wg.Add(1)
		r.workers++
		go r.attack(ctx, cancel)
	}
	Logger.Info("all workers are ready")

	r.wg.Wait()

	Logger.Info("all workers finished the task")
	for _, v := range r.collector.entries {
		fmt.Println(v.Report(true))
	}

	os.Exit(0)
}

func (r *runner) attack(ctx context.Context, cancel context.CancelFunc) {
	defer cancel()
	defer func() { r.workers-- }()
	defer r.wg.Done()
	defer func() {
		if rec := recover(); rec != nil {
			// Todo:
			Logger.Error("recoverd")
		}
	}()

	for {
		select {
		case <-ctx.Done():
			return
		default:
		}

		q := r.task.PickUp()
		start := time.Now()

		select {
		case <-ctx.Done():
			return
		default:
		}

		err := q.Fire()
		duration := time.Since(start)
		r.collector.receiver <- &QueryResult{Name: q.Name(), Duration: duration, Error: err}

		select {
		case <-ctx.Done():
			return
		default:
		}

		wait := r.task.Wait()
		if wait != ZeroDuration {
			time.Sleep(wait)
		}
	}
}

// Worker return current worker counts
func (r *runner) Worker() int {
	return r.workers
}

// SetDuration .
func (r *runner) SetDuration(d time.Duration) *runner {
	r.duration = d
	return r
}

func init() {
	CoreRunner = newRunner(newStatsCollector())
}
