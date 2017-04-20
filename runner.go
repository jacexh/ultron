package ultron

import (
	"errors"
	"fmt"
	"os"
	"os/signal"
	"runtime/debug"
	"sync"
	"sync/atomic"
	"time"

	"go.uber.org/zap"
)

type (
	runner struct {
		Config    *RunnerConfig
		deadLine  time.Time
		requests  uint64
		workers   int64
		task      *TaskSet
		collector *statsCollector
		wg        sync.WaitGroup
		stop      bool
		lock      sync.RWMutex
	}

	// RunnerConfig runner配置参数
	RunnerConfig struct {
		Duration    time.Duration
		Requests    uint64
		Concurrence int
		HatchRate   int
	}

	cleanupFunc func(map[string]*StatsReport)
)

// Runner 执行器
var Runner *runner

func newRunner(c *statsCollector) *runner {
	return &runner{
		Config: &RunnerConfig{
			Duration:    ZeroDuration,
			Requests:    0,
			Concurrence: DefaultConcurrence,
			HatchRate:   0, // start all workers in same time
		},
		collector: c,
	}
}

func (r *runner) Run(t *TaskSet) {
	if t == nil {
		Logger.Panic("invalid TaskSet")
		panic("invalid TaskSet")
	}

	r.task = t

	if err := r.checkConfig(); err != nil {
		panic(err)
	}

	Logger.Info("start")

	go ResultHandleChain.listening()
	go ReportHandleChain.listening()
	go r.checkExitConditions()
	go r.handleInterrupt(printReportToConsole)

	feedTimer := time.NewTicker(StatsReportInterval)
	go func() {
		for range feedTimer.C {
			r.feedReportHandleChain(false)
		}
	}()

	if r.task.OnStart != nil {
		Logger.Info("call OnStart()")
		if err := r.task.OnStart(); err != nil {
			panic(err)
		}
	}

	entries := []string{}
	for e := range r.task.attackers {
		entries = append(entries, e.Name())
	}
	r.collector.createEntries(entries...)

	for _, counts := range r.hatchWorkerCounts() {
		Logger.Info(fmt.Sprintf("hatched %d workers", counts))
		for i := 0; i < counts; i++ {
			r.wg.Add(1)
			atomic.AddInt64(&r.workers, 1)
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

	if r.Config.HatchRate > 0 && r.Config.HatchRate < r.Config.Concurrence {
		rounds = r.Config.Concurrence / r.Config.HatchRate
		for i := 0; i < rounds; i++ {
			ret = append(ret, r.Config.HatchRate)
		}

		last := r.Config.Concurrence % r.Config.HatchRate
		if last > 0 {
			ret = append(ret, last)
		}

	} else {
		ret = append(ret, r.Config.Concurrence)
	}

	return ret
}

func (r *runner) attack() {
	defer func() { atomic.AddInt64(&r.workers, -1) }()
	defer r.wg.Done()
	defer func() {
		if rec := recover(); rec != nil {
			// Todo:
			debug.PrintStack()
			Logger.Error("recoverd")
		}
	}()

	for {
		q := r.task.pickUp()
		start := time.Now()

		if r.shouldStop() {
			return
		}

		size, err := q.Fire()
		duration := time.Since(start)
		ResultHandleChain.channel() <- &AttackResult{Name: q.Name(), Duration: duration, Error: err, Received: size}

		if r.shouldStop() {
			return
		}

		wait := r.task.wait()
		if wait != ZeroDuration {
			time.Sleep(wait)
		}
		atomic.AddUint64(&r.requests, 1)
	}
}

// Worker return current worker counts
func (r *runner) Worker() int64 {
	return r.workers
}

func (r *runner) setDeadline() {
	if r.Config.Duration > ZeroDuration {
		r.lock.Lock()
		r.deadLine = time.Now().Add(r.Config.Duration)
		r.lock.Unlock()
		Logger.Info("set deadline", zap.Time("deadline", r.deadLine))
	}
}

func (r *runner) handleInterrupt(c cleanupFunc) {
	signalCh := make(chan os.Signal, 1)
	signal.Notify(signalCh, os.Interrupt)
	go func() {
		<-signalCh
		c(r.collector.report(true))
		os.Exit(1)
	}()
}

func (r *runner) shouldStop() bool {
	r.lock.RLock()
	defer r.lock.RUnlock()
	return r.stop
}

func (r *runner) checkExitConditions() {
	for {
		if r.Config.Duration > ZeroDuration {
			r.lock.Lock()
			if !r.deadLine.IsZero() && time.Now().After(r.deadLine) {
				r.stop = true
				r.lock.Unlock()
				break
			}
			r.lock.Unlock()
		}
		if r.Config.Requests > 0 {
			r.lock.Lock()
			if atomic.LoadUint64(&r.requests) >= r.Config.Requests {
				r.stop = true
				r.lock.Unlock()
				break
			}
			r.lock.Unlock()
		}
		time.Sleep(time.Millisecond * 200)
	}
	Logger.Info("should end the runner")
}

func (r *runner) feedReportHandleChain(fullHistory bool) {
	ret := r.collector.report(fullHistory)
	ReportHandleChain.channel() <- ret
}

func (r *runner) checkConfig() error {
	if r.Config.Concurrence == 0 {
		Logger.Error("invalid Concurrence value", zap.Int("Concurrenct", r.Config.Concurrence))
		return errors.New("invalid Concurrence value")
	}
	return nil
}

func init() {
	Runner = newRunner(defaultStatsCollector)
}
