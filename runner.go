package ultron

import (
	"context"
	"fmt"
	"os"
	"sync"
	"sync/atomic"
	"time"
)

type runner struct {
	duration      time.Duration
	deadLine      time.Time
	requests      uint64
	totalRequests uint64
	workers       int
	task          *TaskSet
	collector     *statsCollector
	ctx           context.Context
	cancel        context.CancelFunc
	wg            *sync.WaitGroup
	shouldStop    bool
	lock          sync.RWMutex
}

// CoreRunner 核心执行器
var CoreRunner *runner

func newRunner(c *statsCollector) *runner {
	return &runner{
		collector: newStatsCollector(),
		duration:  ZeroDuration,
		wg:        &sync.WaitGroup{},
	}
}

func (r *runner) WithTaskSet(t *TaskSet) *runner {
	r.task = t
	return r
}

func (r *runner) Run() {
	Logger.Info("start")
	go r.checkStatus()
	go r.collector.Receiving()

	if r.task.OnStart != nil {
		Logger.Info("call OnStart()")
		if err := r.task.OnStart(); err != nil {
			panic(err)
		}
	}

	if r.duration > ZeroDuration {
		r.deadLine = time.Now().Add(r.duration)
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
		go r.attack()
	}
	Logger.Info("all workers are ready")

	r.wg.Wait()

	close(r.collector.Receiver())
	Logger.Info("all workers finished the task")
	for _, v := range r.collector.entries {
		fmt.Println(v.Report(true))
	}

	os.Exit(0)
}

func (r *runner) attack() {
	defer func() { r.workers-- }()
	defer r.wg.Done()
	defer func() {
		if rec := recover(); rec != nil {
			// Todo:
			Logger.Error("recoverd")
		}
	}()

	for {
		if r.shouldStop {
			return
		}
		q := r.task.PickUp()
		start := time.Now()

		if r.shouldStop {
			return
		}
		err := q.Fire()
		duration := time.Since(start)
		r.collector.receiver <- &QueryResult{Name: q.Name(), Duration: duration, Error: err}

		if r.shouldStop {
			return
		}
		wait := r.task.Wait()
		if wait != ZeroDuration {
			time.Sleep(wait)
		}
		atomic.AddUint64(&r.requests, 1)
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

func (r *runner) SetTotalRequests(n uint64) *runner {
	r.totalRequests = n
	return r
}

func (r *runner) checkStatus() {
	for {
		if r.duration > ZeroDuration && time.Now().After(r.deadLine) {
			r.shouldStop = true
			break
		}
		if r.totalRequests > 0 && atomic.LoadUint64(&r.requests) >= r.totalRequests {
			r.shouldStop = true
			break
		}
		time.Sleep(time.Millisecond * 200)
	}
}

func init() {
	CoreRunner = newRunner(newStatsCollector())
}
