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

const (
	// StatusIdle 空闲状态
	StatusIdle Status = iota
	// StatusBusy 执行中状态
	StatusBusy
	// StatusStopped 已经停止状态
	StatusStopped
)

var (
	// LocalRunner 单机执行入口
	LocalRunner *localRunner

	// SlaveRunner 分布式执行，节点执行入口
	SlaveRunner *slaveRunner
)

type (
	// Runner 定义执行器接口
	Runner interface {
		WithConfig(*RunnerConfig)
		WithTask(*Task)
		GetConfig() *RunnerConfig
		GetStatus() Status
		Start()
		Done()
	}

	// Status Runner状态
	Status int

	baseRunner struct {
		Config   *RunnerConfig
		task     *Task
		status   Status
		counts   uint64
		deadline time.Time
		mu       sync.RWMutex
		wg       sync.WaitGroup
	}

	localRunner struct {
		stats *summaryStats
		once  sync.Once
		*baseRunner
	}

	slaveRunner struct {
		*baseRunner
	}
)

func newBaseRunner() *baseRunner {
	return &baseRunner{status: StatusIdle, Config: DefaultRunnerConfig}
}

func (br *baseRunner) WithConfig(rc *RunnerConfig) {
	br.Config = rc
}

func (br *baseRunner) WithTask(t *Task) {
	br.task = t
}

func (br *baseRunner) GetConfig() *RunnerConfig {
	return br.Config
}

func (br *baseRunner) Done() {
	br.mu.Lock()
	defer br.mu.Unlock()
	br.status = StatusStopped
}

func isFinished(br *baseRunner) bool {
	if br.GetStatus() == StatusStopped {
		return true
	}

	if br.Config.Requests > 0 && atomic.LoadUint64(&br.counts) >= br.Config.Requests {
		br.Done()
		return true
	}

	br.mu.RLock()
	if br.Config.Duration > ZeroDuration && !br.deadline.IsZero() && time.Now().After(br.deadline) {
		br.mu.RUnlock()
		br.Done()
		return true
	}
	br.mu.RUnlock()
	return false
}

func (br *baseRunner) GetStatus() Status {
	br.mu.RLock()
	defer br.mu.RUnlock()
	return br.status
}

func newLocalRunner(ss *summaryStats) *localRunner {
	return &localRunner{stats: ss, baseRunner: newBaseRunner()}
}

func checkRunner(br *baseRunner) error {
	if br.task == nil {
		return errors.New("no Task provided")
	}

	if br.Config == nil {
		return errors.New("no RunnerConfig provided")
	}

	if err := br.Config.check(); err != nil {
		return err
	}
	return nil
}

func (lr *localRunner) log(r *Result) {
	lr.stats.log(r)
}

func (lr *localRunner) Start() {
	if err := checkRunner(lr.baseRunner); err != nil {
		Logger.Panic("occur error", zap.Error(err))
		panic(err)
	}

	Logger.Info("start to attack")
	lr.status = StatusBusy

	lr.once.Do(func() {
		localReportPipeline = newReportPipeline(LocalReportPipelineBufferSize)
		localResultPipeline = newResultPipeline(LocalResultPipelineBufferSize)
		LocalEventHook.AddResultHandleFunc(lr.log)
		go LocalEventHook.listen(localResultPipeline, localReportPipeline)

		signalCh := make(chan os.Signal, 1)
		signal.Notify(signalCh, os.Interrupt)
		go func() {
			<-signalCh
			Logger.Error("capatured interrupt signal")
			printReportToConsole(lr.stats.report(true))
			os.Exit(1)
		}()

	})

	lr.stats.reset()

	feedTicker := time.NewTicker(StatsReportInterval)
	go func() {
		for range feedTicker.C {
			localReportPipeline <- lr.stats.report(false)
		}
	}()

	go func() {
		t := time.NewTicker(time.Millisecond * 200)
		for range t.C {
			if isFinished(lr.baseRunner) {
				t.Stop()
				break
			}
		}
	}()

	hatchWorkers(lr.baseRunner, localResultPipeline)

	if lr.Config.Duration > ZeroDuration {
		lr.mu.Lock()
		lr.deadline = time.Now().Add(lr.Config.Duration)
		lr.mu.Unlock()
		Logger.Info("set deadline", zap.Time("deadline", lr.deadline))
	}
	Logger.Info("hatched complete")

	lr.wg.Wait()

	feedTicker.Stop()
	localReportPipeline <- lr.stats.report(true)

	Logger.Info("task done")
	time.Sleep(time.Second * 1)
	os.Exit(0)
}

func hatchWorkers(br *baseRunner, ch resultPipeline) {
	var hatched int
	for _, counts := range br.Config.hatchWorkerCounts() {
		for i := 0; i < counts; i++ {
			br.wg.Add(1)
			go attack(br, ch)
		}
		hatched += counts
		time.Sleep(time.Second)
		Logger.Info(fmt.Sprintf("hatched %d workers", hatched))
	}
}

func attack(br *baseRunner, ch resultPipeline) {
	defer br.wg.Done()
	defer func() {
		if rec := recover(); rec != nil {
			// Todo:
			debug.PrintStack()
			Logger.Error("recovered")
		}
	}()

	if ch == nil {
		panic("invalid resultPipeline")
	}

	for {
		q := br.task.pickUp()
		start := time.Now()
		if br.GetStatus() == StatusStopped {
			return
		}
		err := q.Fire()
		duration := time.Since(start)
		atomic.AddUint64(&br.counts, 1)
		ret := newResult(q.Name(), duration, err)
		ch <- ret

		if br.GetStatus() == StatusStopped {
			return
		}
		br.Config.block()
	}
}
