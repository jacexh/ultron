package ultron

import (
	"context"
	"errors"
	"fmt"
	"math/rand"
	"os"
	"os/signal"
	"runtime/debug"
	"sync"
	"sync/atomic"
	"time"
	"go.uber.org/zap"
	"github.com/qastub/ultron/utils"
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

	// context的父节点，主控入口持有
	_parentCtx, _parentCancel = context.WithCancel(context.Background())
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
		deadline time.Time          //总体停止时间
		cancels  []context.CancelFunc
		mu       sync.RWMutex
		wg       sync.WaitGroup
	}

	localRunner struct {
		stats *summaryStats
		once  sync.Once
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

func (br *baseRunner) WithDeadLine(deadline time.Time) {
	br.deadline = deadline
}

func (br *baseRunner) GetConfig() *RunnerConfig {
	return br.Config
}

// TODO
func (br *baseRunner) Done() {
	br.mu.Lock()
	defer br.mu.Unlock()
	br.status = StatusStopped
}

func (br *baseRunner) GetStageRunningTime() []time.Duration{
	br.mu.RLock()
	defer br.mu.RUnlock()
	var stageRunningTime []time.Duration

	for _, StageConfig := range br.Config.Stages {
		stageRunningTime = append(stageRunningTime, StageConfig.Duration)
	}
	return stageRunningTime
}

// TODO
func isFinished(br *baseRunner) bool {
	//if br.GetStatus() == StatusStopped {
	//	return true
	//}
	//
	//if br.Config.Requests > 0 && atomic.LoadUint64(&br.counts) >= br.Config.Requests {
	//	br.Done()
	//	return true
	//}
	//
	//br.mu.RLock()
	//if br.Config.Duration > ZeroDuration && !br.deadline.IsZero() && time.Now().After(br.deadline) {
	//	br.mu.RUnlock()
	//	br.Done()
	//	return true
	//}
	//br.mu.RUnlock()
	br.mu.RLock()
	defer br.mu.RUnlock()

	if br.GetStatus() == StatusStopped {
		return true
	}
	return false
}

func (br *baseRunner) GetStatus() Status {
	br.mu.RLock()
	defer br.mu.RUnlock()
	return br.status
}


func (br *baseRunner) getStageConfig() []*StageConfig{
	br.mu.RLock()
	defer br.mu.RUnlock()
	return br.Config.Stages
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

	br.Config.UpdateStageConfig()
	return nil
}

func (lr *localRunner) log(r *Result) {
	lr.stats.log(r)
}


func CountNumbers2Stop(countPipeline countPipeline, number2Stop *uint64) {
	defer Logger.Info("func CountNumbers2Stop have finished")
	Logger.Info("set requests", zap.Uint64("requests", *number2Stop))
	var count uint64 = 0
	if *number2Stop <= 0 {
		for {
			select {
			case <-countPipeline:
				// do nothing
			}
		}
	} else {
		for {
			if atomic.LoadUint64(&count) < *number2Stop {
				select {
				case <-countPipeline:
					atomic.AddUint64(&count, 1)
				}
			} else {
				Logger.Info(fmt.Sprintf("have finished %d requests.STOP!", atomic.LoadUint64(&count)))
				StageRunnerStatusPipeline <- StatusStopped
			}
		}
	}
}

// 主控入口
func statusControl(ch chan Status) {
	for {
		select {
		case status := <- ch:
			if status == StatusStopped {
				_parentCancel()
				Logger.Info("stageRunner status is stoped.STOP!")
				time.Sleep(2 * time.Second)
				os.Exit(0)
			}
		}
	}
}

// 随机取消运行中的协程
func (br *baseRunner) CancelWorkers(num int) error{
	br.mu.Lock()
	defer br.mu.Unlock()
	Logger.Info(fmt.Sprintf("cancel %d workers", num))

	if num < 0 && len(br.cancels) < num {
		return errors.New("CancelWorkers num wrong")
	}

	for i := 0; i < num; i++ {
		//Logger.Info(fmt.Sprintf("cancel a attacks"))
		index := rand.Intn(len(br.cancels))
		br.cancels[index]()
		br.cancels = append(br.cancels[:index], br.cancels[index+1:]...)
	}
	return nil

}


func (br *baseRunner) AddCancelFunc(cancel *context.CancelFunc) {
	br.mu.Lock()
	defer br.mu.Unlock()

	br.cancels = append(br.cancels, *cancel)
}


func createCancelFunc(br *baseRunner, parentctx context.Context) (context.Context, context.CancelFunc) {
	if br.deadline.IsZero() {
		Logger.Info("create WithCancel context")
		return context.WithCancel(parentctx)
	} else {
		Logger.Info("create WithDeadline context. dead at ", zap.Time("deadline", br.deadline))
		return context.WithDeadline(parentctx, br.deadline)
	}
}


//func (lr *localRunner) Start() {
//	if err := checkRunner(lr.baseRunner); err != nil {
//		Logger.Panic("occur error", zap.Error(err))
//		panic(err)
//	}
//
//	Logger.Info("start to attack")
//	lr.status = StatusBusy
//
//	lr.once.Do(func() {
//		localReportPipeline = newReportPipeline(LocalReportPipelineBufferSize)
//		localResultPipeline = newResultPipeline(LocalResultPipelineBufferSize)
//		LocalEventHook.AddResultHandleFunc(lr.log)
//		LocalEventHook.listen(localResultPipeline, localReportPipeline)
//
//		signalCh := make(chan os.Signal, 1)
//		signal.Notify(signalCh, os.Interrupt)
//		go func() {
//			<-signalCh
//			Logger.Error("capatured interrupt signal")
//			printReportToConsole(lr.stats.report(true))
//			os.Exit(1)
//		}()
//
//	})
//
//	lr.stats.reset()
//
//	feedTicker := time.NewTicker(StatsReportInterval)
//	go func() {
//		for range feedTicker.C {
//			localReportPipeline <- lr.stats.report(false)
//		}
//	}()
//
//	go func() {
//		t := time.NewTicker(200 * time.Millisecond)
//		for range t.C {
//			if isFinished(lr.baseRunner) {
//				t.Stop()
//				break
//			}
//		}
//	}()
//
//	hatchWorkers(lr.baseRunner, localResultPipeline)
//
//	if lr.Config.Duration > ZeroDuration {
//		lr.mu.Lock()
//		lr.deadline = time.Now().Add(lr.Config.Duration)
//		lr.mu.Unlock()
//		Logger.Info("set deadline", zap.Time("deadline", lr.deadline))
//	}
//	Logger.Info("hatched complete")
//
//	lr.wg.Wait()
//
//	feedTicker.Stop()
//	localReportPipeline <- lr.stats.report(true)
//
//	Logger.Info("task done")
//	time.Sleep(1 * time.Second)
//	os.Exit(0)
//}


func (lr *localRunner) Start() {
	fmt.Println(lr.baseRunner.Config.Stages)
	if err := checkRunner(lr.baseRunner); err != nil {
		panic(err)
	}

	lr.once.Do(func() {
		localReportPipeline = newReportPipeline(LocalReportPipelineBufferSize)
		localResultPipeline = newResultPipeline(LocalResultPipelineBufferSize)
		LocalEventHook.AddResultHandleFunc(lr.stats.log)
		LocalEventHook.listen(localResultPipeline, localReportPipeline)

		// ctrl+c退出,输出信号
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

	Logger.Info("stages start")
	defer _parentCancel()

	//counter := newCounter(1000)

	feedTicker := time.NewTicker(StatsReportInterval)
	go func() {
		for range feedTicker.C {
			localReportPipeline <- lr.stats.report(false)  // 管道传输统计结果
		}
	}()
	//defer func() {
	//	fmt.Println("feedTicker.Stop()")
	//	feedTicker.Stop()
	//	localReportPipeline <- sr.stats.report(true)
	//}()

	go CountNumbers2Stop(CounterPipeline, &lr.Config.Requests)

	go statusControl(StageRunnerStatusPipeline)

	//Logger.Info(fmt.Sprintf("deadline at ", sr.deadline))
	//pctx, pcancel := context.WithDeadline(parentCtx, sr.deadline)
	pctx, pcancel := createCancelFunc(lr.baseRunner, _parentCtx)

	//[]time.Duration{1 * time.Second, 2 * time.Second, 4 * time.Second}
	timers := utils.NewTimers(lr.GetStageRunningTime())
	Logger.Info("start to attack")
	lr.status = StatusBusy

	for {
		select {
		case <- pctx.Done():
			//fmt.Println("feedTicker.Stop()")
			lr.baseRunner.Done()
			feedTicker.Stop()
			//localReportPipeline <- lr.stats.report(true)
			printReportToConsole(lr.stats.report(true))
			Logger.Info("stages have be done. STOP!")
			StageRunnerStatusPipeline <- StatusStopped
			return
		case cc := <-timers.C:
			if cc >= 0 && cc <= len(lr.baseRunner.Config.Stages) - 1 {
				scc := lr.baseRunner.Config.Stages[cc]
				Logger.Info("start ", zap.Int("task：", cc))

				func() {

					//if scc.Concurrence > 0 {
					//	hatchWorkersCancelable(pctx, lr.baseRunner, scc, localResultPipeline, CounterPipeline)
					//}
					if scc.Concurrence == 0 {
						// do nothing
						Logger.Info("keep Concurrence")
					} else {
						hatchWorkersCancelable(pctx, lr.baseRunner, scc, localResultPipeline, CounterPipeline)
					}
					//if scc.Concurrence < 0 {
					//	hatchCancelWorkers(lr.baseRunner, scc)
					//}
					//Logger.Info(fmt.Sprintf("time.sleep %s", scc.Duration))
					//time.Sleep(scc.Duration)

					//Logger.Info("task done", zap.Int("task：", cc))
				}()
			} else {
				Logger.Info("pcancel()")
				pcancel()
				StageRunnerStatusPipeline <- StatusStopped
			}
		}
	}
}



//func hatchCancelWorkers(br *baseRunner, sc *StageConfig) {
//
//	var hatched int
//	for _, counts := range sc.hatchWorkerCounts() {
//		br.CancelWorkers(utils.Abs(counts))
//		hatched += counts
//		time.Sleep(time.Second)
//		Logger.Info(fmt.Sprintf("cancel %d workers", hatched))
//		time.Sleep(1 *time.Second)
//	}
//
//}


// countPipeline需要有消费者，否则超过buffer会阻塞
func hatchWorkersCancelable(parentctx context.Context, br *baseRunner, sc *StageConfig, ch resultPipeline, countPipe countPipeline) {

	var hatched int
	if sc.Concurrence > 0 {
		for _, counts := range sc.hatchWorkerCounts() {
			for i := 0; i < counts; i++ {
				ctx, cancel := context.WithCancel(parentctx)
				//sr.cancels = append(sr.cancels, cancel)
				br.AddCancelFunc(&cancel) // append cancel
				//sr.wg.Add(1)
				go attackCancelAble(ctx, br, ch, countPipe)
			}
			hatched += counts
			Logger.Info(fmt.Sprintf("hatched %d workers", hatched))
			time.Sleep(time.Second)
		}
	} else {
		for _, counts := range sc.hatchWorkerCounts() {
			//取消协程
			br.CancelWorkers(utils.Abs(counts))
			hatched += counts
			Logger.Info(fmt.Sprintf("hatched %d workers", hatched))
			time.Sleep(time.Second)
		}
	}

}


func attackCancelAble(ctx context.Context, br *baseRunner, ch resultPipeline, countPipe countPipeline) {
	//defer sr.wg.Done()
	defer ctx.Done()
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
	//fmt.Println("hahhahahaha")

	for {
		select {
		case <-ctx.Done():
			return
		default:
			q := br.task.pickUp()
			start := time.Now()
			if br.GetStatus() == StatusStopped {
				return
			}
			err := q.Fire()
			duration := time.Since(start)

			countPipe <- 1 // 往计数channel发送信号
			//atomic.AddUint64(&br.counts, 1)
			ret := newResult(q.Name(), duration, err)
			ch <- ret

			if err != nil {
				Logger.Warn("occur error: " + err.Error())
			}

			if br.GetStatus() == StatusStopped {
				return
			}
			br.Config.block()
		}
	}

}

////TODO
//func hatchWorkers(br *baseRunner, ch resultPipeline) {
//	var hatched int
//	for _, counts := range br.Config.hatchWorkerCounts() {
//		for i := 0; i < counts; i++ {
//			br.wg.Add(1)
//			go attack(br, ch)
//		}
//		hatched += counts
//		time.Sleep(time.Second)
//		Logger.Info(fmt.Sprintf("hatched %d workers", hatched))
//	}
//}
//
////TODO
//func attack(br *baseRunner, ch resultPipeline) {
//	defer br.wg.Done()
//	defer func() {
//		if rec := recover(); rec != nil {
//			// Todo:
//			debug.PrintStack()
//			Logger.Error("recovered")
//		}
//	}()
//
//	if ch == nil {
//		panic("invalid resultPipeline")
//	}
//
//	for {
//		q := br.task.pickUp()
//		start := time.Now()
//		if br.GetStatus() == StatusStopped {
//			return
//		}
//		err := q.Fire()
//		duration := time.Since(start)
//		atomic.AddUint64(&br.counts, 1)
//		ret := newResult(q.Name(), duration, err)
//		ch <- ret
//
//		if err != nil {
//			Logger.Warn("occur error: " + err.Error())
//		}
//
//		if br.GetStatus() == StatusStopped {
//			return
//		}
//		br.Config.block()
//	}
//}



