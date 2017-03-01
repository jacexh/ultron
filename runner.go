package ultron

import (
	"context"
	"fmt"
	"os"
	"runtime/debug"
	"sync"
	"sync/atomic"
	"time"

	"go.uber.org/zap"
)

type runner struct {
	concurrence   int
	duration      time.Duration
	deadLine      time.Time
	requests      uint64
	totalRequests uint64
	workers       int
	hatchRate     int
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
		concurrence: DefaultConcurrence,
		collector:   c,
		duration:    ZeroDuration,
		wg:          &sync.WaitGroup{},
	}
}

func (r *runner) WithTaskSet(t *TaskSet) *runner {
	r.task = t
	return r
}

func (r *runner) Run() {
	Logger.Info("start")

	go ResultHandleChain.listening()
	go ReportHandleChain.listening()
	go r.checkExitConditions()

	feedTimer := time.NewTimer(StatsReportInterval)
	go func() {
		for {
			<-feedTimer.C
			r.feedReportHandleChain(false)
			feedTimer.Reset(StatsReportInterval)
		}
	}()

	if r.task.OnStart != nil {
		Logger.Info("call OnStart()")
		if err := r.task.OnStart(); err != nil {
			panic(err)
		}
	}

	for _, counts := range r.hatchWorkerCounts() {
		Logger.Info(fmt.Sprintf("start %d workers", counts))
		for i := 0; i < counts; i++ {
			r.wg.Add(1)
			r.workers++
			go r.attack()
		}
		time.Sleep(time.Second * 1)
	}
	Logger.Info("hatch complete")
	r.setDeadline()

	r.wg.Wait()

	feedTimer.Stop()
	r.feedReportHandleChain(true)

	ResultHandleChain.safeClose()
	ReportHandleChain.safeClose()

	Logger.Info("task done")
	time.Sleep(time.Second * 1) // wait for print total stats
	os.Exit(0)
}

func (r *runner) hatchWorkerCounts() []int {
	rounds := 1
	ret := []int{}

	if r.hatchRate > 0 && r.hatchRate < r.concurrence {
		rounds = r.concurrence / r.hatchRate
		for i := 0; i < rounds; i++ {
			ret = append(ret, r.hatchRate)
		}

		last := r.concurrence % r.hatchRate
		if last > 0 {
			ret = append(ret, last)
		}

	} else {
		ret = append(ret, r.concurrence)
	}

	return ret
}

func (r *runner) attack() {
	defer func() { r.workers-- }()
	defer r.wg.Done()
	defer func() {
		if rec := recover(); rec != nil {
			// Todo:
			debug.PrintStack()
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
		ResultHandleChain.channel() <- &QueryResult{Name: q.Name(), Duration: duration, Error: err}

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

func (r *runner) setDeadline() {
	if r.duration > ZeroDuration {
		r.deadLine = time.Now().Add(r.duration)
		Logger.Info("set deadline", zap.Time("deadline", r.deadLine))
	}
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

func (r *runner) checkExitConditions() {
	for {
		if r.duration > ZeroDuration && !r.deadLine.IsZero() && time.Now().After(r.deadLine) {
			r.shouldStop = true
			break
		}
		if r.totalRequests > 0 && atomic.LoadUint64(&r.requests) >= r.totalRequests {
			r.shouldStop = true
			break
		}
		time.Sleep(time.Millisecond * 200)
	}
	Logger.Info("should end the runner")
}

func (r *runner) feedReportHandleChain(fullHistory bool) {
	ret := r.collector.report(fullHistory)
	ReportHandleChain.channel() <- ret
}

func (r *runner) SetHatchRate(n int) *runner {
	r.hatchRate = n
	return r
}

func (r *runner) SetConcurrence(n int) *runner {
	if n > 0 {
		r.concurrence = n
	}
	return r
}

func init() {
	CoreRunner = newRunner(defaultStatsCollector)
}
