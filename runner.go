package ultron

import (
	"context"
	"errors"
	"fmt"
	"github.com/qastub/ultron/utils"
	"go.uber.org/zap"
	"os"
	"os/signal"
	"runtime/debug"
	"sync"
	"sync/atomic"
	"time"
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

	BaseRunner baseRunner

	// Status Runner状态
	Status int

	baseRunner struct {
		Config   *RunnerConfig			`json:"Config"`
		task     *Task
		status   Status
		counts   uint64
		deadline time.Time               //`json:"deadline,omitempy"`        //总体停止时间
		cancels  cancelMap
		mu       sync.RWMutex
		wg       sync.WaitGroup
	}

	localRunner struct {
		stats *summaryStats
		once  sync.Once
		*baseRunner
	}

	cancelMap struct {
		index   uint8
		cancels map[uint8]context.CancelFunc
	}
)

func newCancelMap() cancelMap {
	return cancelMap{cancels: make(map[uint8]context.CancelFunc)}
}

func newBaseRunner() *baseRunner {
	return &baseRunner{status: StatusIdle, Config: DefaultRunnerConfig, cancels: newCancelMap()}
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

// 把阶段运行时间组成一个列表
func (br *baseRunner) GetStageRunningTime() []time.Duration{
	br.mu.RLock()
	defer br.mu.RUnlock()
	var stageRunningTime []time.Duration

	for _, StageConfig := range br.Config.Stages {
		stageRunningTime = append(stageRunningTime, StageConfig.Duration)
	}
	return stageRunningTime
}

//TODO
// master通过主动查询来确保结束
func isFinished(br *baseRunner) bool {
	if br.GetStatus() == StatusStopped {
		Logger.Debug("StatusStopped RUNNER IS FINISHED")
		return true
	}

	if br.Config.Requests > 0 && atomic.LoadUint64(&br.counts) >= br.Config.Requests {
		br.Done()
		Logger.Debug("counts RUNNER IS FINISHED")
		return true
	}

	br.mu.RLock()
	if !br.deadline.IsZero() && time.Now().After(br.deadline) {
		br.mu.RUnlock()
		br.Done()
		Logger.Debug("Deadline RUNNER IS FINISHED")
		return true
	}
	br.mu.RUnlock()

	return false
}


//for slave
func isOverAmount(br *baseRunner) bool {

	if br.GetStatus() == StatusStopped {
		Logger.Info("StatusStopped RUNNER IS FINISHED")
		return true
	}

	if br.Config.Requests > 0 && atomic.LoadUint64(&br.counts) >= br.Config.Requests {
		br.Done()
		Logger.Info("requests is over. RUNNER FINISHED")
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

	br.activeBaseRunner()
	return nil
}


// 将stage配置及deadline 更新 为 Machine friendly
// 需要在压测开始前运行
func (br *baseRunner) activeBaseRunner() {

	br.Config.updateStageConfig()
	br.updateDeadline()
}


func (lr *localRunner) log(r *Result) {
	lr.stats.log(r)
}


//func CountNumbers2Stop(countPipeline countPipeline, number2Stop *uint64) {
//	defer Logger.Info("func CountNumbers2Stop have finished")
//	Logger.Info("set requests", zap.Uint64("requests", *number2Stop))
//	var count uint64 = 0
//	if *number2Stop <= 0 {
//		for {
//			select {
//			case <-countPipeline:
//				// do nothing
//			}
//		}
//	} else {
//		for {
//			if atomic.LoadUint64(&count) < *number2Stop {
//				select {
//				case <-countPipeline:
//					atomic.AddUint64(&count, 1)
//				}
//			} else {
//				Logger.Info(fmt.Sprintf("have finished %d requests.STOP!", atomic.LoadUint64(&count)))
//				StageRunnerStatusPipeline <- StatusStopped
//			}
//		}
//	}
//}

//// 主控入口  for localrunner
//func statusControlEndExit(ch chan Status, pcancel context.CancelFunc) {
//	for {
//		select {
//		case status := <- ch:
//			if status == StatusStopped {
//				pcancel()
//				Logger.Info("stageRunner status is stoped.STOP!")
//				time.Sleep(2 * time.Second)
//				os.Exit(0)
//			}
//		}
//	}
//}

// 主控入口
func statusControl(ch chan Status, pcancel context.CancelFunc, isExit bool) {
	for {
		select {
		case status := <- ch:
			if status == StatusStopped {
				Logger.Info("pcancel()")
				pcancel()
				Logger.Info("stageRunner status is stoped.STOP!")
				//是否在最后 退出整个程序
				if isExit {
					time.Sleep(2 * time.Second)
					os.Exit(0)
				}
			}
		}
	}
}

// 随机取消运行中的协程
func (br *baseRunner) CancelWorkers(num int) error{
	br.mu.Lock()
	defer br.mu.Unlock()
	Logger.Info(fmt.Sprintf("cancel %d workers", num))

	if num < 0 || len(br.cancels.cancels) < num {
		return errors.New("CancelWorkers num wrong")
	}

	//for i := 0; i < num; i++ {
	//	//Logger.Info(fmt.Sprintf("cancel a attacks"))
	//	index := rand.Intn(len(br.cancels))
	//	br.cancels[index]()
	//	br.cancels = append(br.cancels[:index], br.cancels[index+1:]...)
	//}
	//return nil

	for k, v := range br.cancels.cancels {
		if _, ok := br.cancels.cancels[k]; ok {
			v()
			delete(br.cancels.cancels, k)
		} else {
			Logger.Warn("wrong cancelFunc")
		}
	}
	return nil
}


func (br *baseRunner) AddCancelFunc(cancel *context.CancelFunc) {
	br.mu.Lock()

	br.cancels.cancels[br.cancels.index] = *cancel
	br.cancels.index ++

	br.mu.Unlock()
}


//更新deadline
func (br *baseRunner) updateDeadline() {
	br.mu.Lock()
	defer br.mu.Unlock()
	var d = time.Now()

	//配置过deadline的，不更新deadline
	if !br.deadline.Equal(time.Time{}) {
		return
	}

	for _, sc := range br.Config.stagesChanged {
		if sc.Duration <= ZeroDuration {
			br.WithDeadLine(time.Time{})
			return
		} else {
			d = d.Add(sc.Duration)
		}
	}

	br.WithDeadLine(d)
	return
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


func (lr *localRunner) Start() {
	Logger.Info("deadline info", zap.Time("deadline", lr.baseRunner.deadline))
	Logger.Info("baseRunner info", zap.Any("baseRunner", lr.baseRunner))
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


	feedTicker := time.NewTicker(StatsReportInterval)
	go func() {
		for range feedTicker.C {
			localReportPipeline <- lr.stats.report(false)  // 管道传输统计结果
		}
	}()


	//go CountNumbers2Stop(CounterPipeline, &lr.Config.Requests)

	pctx, pcancel := createCancelFunc(lr.baseRunner, _parentCtx)

	go statusControl(StageRunnerStatusPipeline, pcancel, true)

	go func() {
		t := time.NewTicker(200 * time.Millisecond)
		for range t.C {
			if isOverAmount(lr.baseRunner) {
				t.Stop()
				StageRunnerStatusPipeline <- StatusStopped
			}
		}
	}()

	timers := NewTimers(lr.GetStageRunningTime())
	Logger.Info("start to attack")
	lr.status = StatusBusy

	for {
		select {
		case <- pctx.Done():

			lr.baseRunner.Done()
			feedTicker.Stop()

			printReportToConsole(lr.stats.report(true))
			Logger.Info("stages have be done. STOP!")
			StageRunnerStatusPipeline <- StatusStopped
			return
		case cc := <-timers.C:
			if cc >= 0 && cc <= len(lr.baseRunner.Config.stagesChanged) - 1 {
				scc := lr.baseRunner.Config.stagesChanged[cc]
				Logger.Info("start ", zap.Int("task：", cc))

				func() {

					if scc.Concurrence == 0 {
						// do nothing
						Logger.Info("keep Concurrence")
					} else {
						hatchWorkersCancelable(pctx, lr.baseRunner, scc, localResultPipeline)
					}
				}()
			} else {
				Logger.Info("pcancel()")
				pcancel()
				StageRunnerStatusPipeline <- StatusStopped
			}
		}
	}
}


// countPipeline需要有消费者，否则超过buffer会阻塞
func hatchWorkersCancelable(parentctx context.Context, br *baseRunner, sc *StageConfigChanged, ch resultPipeline) {

	var hatched int
	if sc.Concurrence > 0 {
		for _, counts := range sc.hatchWorkerCounts() {
			for i := 0; i < counts; i++ {
				ctx, cancel := context.WithCancel(parentctx)
				br.AddCancelFunc(&cancel) // append cancel
				go attackCancelAble(ctx, br, ch)
			}
			hatched += counts
			Logger.Info(fmt.Sprintf("hatched %d workers", hatched))
			time.Sleep(time.Second)
		}
	} else {
		for _, counts := range sc.hatchWorkerCounts() {
			//取消协程
			br.CancelWorkers(Abs(counts))
			hatched += counts
			Logger.Info(fmt.Sprintf("hatched %d workers", hatched))
			time.Sleep(time.Second)
		}
	}

}


func attackCancelAble(ctx context.Context, br *baseRunner, ch resultPipeline) {
	//defer sr.wg.Done()
	//defer ctx.Done()
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
			//Logger.Info("fire")
			duration := time.Since(start)

			atomic.AddUint64(&br.counts, 1)
			ret := newResult(q.Name(), duration, err)
			ch <- ret

			if err != nil {
				Logger.Warn("occur error: " + err.Error())
			}

			//if br.GetStatus() == StatusStopped {
			//	return
			//}
			br.Config.block()
		}
	}
}