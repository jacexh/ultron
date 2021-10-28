package slave

import (
	"context"
	"fmt"
	"runtime/debug"
	"sync"
	"sync/atomic"
	"time"

	"github.com/wosai/ultron/v2"
	"github.com/wosai/ultron/v2/log"
	"github.com/wosai/ultron/v2/pkg/statistics"
	"go.uber.org/zap"
)

type (
	fixedConcurrentUsersStrategyCommander struct {
		ctx       context.Context
		cancel    context.CancelFunc
		describer ultron.AttackStrategy
		output    chan statistics.AttackResult
		timer     ultron.Timer
		task      *ultron.Task
		counter   uint32
		pool      map[uint32]*fcuExecutor
		wg        *sync.WaitGroup
		mu        sync.Mutex
	}

	// fcuExecutor FixedConcurrentUsers策略的执行者
	fcuExecutor struct {
		id     uint32
		cancel context.CancelFunc
		timer  ultron.Timer
		mu     sync.RWMutex
	}
)

var (
	_ ultron.AttackStrategyCommander = (*fixedConcurrentUsersStrategyCommander)(nil)
)

func newFixedConcurrentUsersStrategyCommander() *fixedConcurrentUsersStrategyCommander {
	return &fixedConcurrentUsersStrategyCommander{
		ctx:    context.TODO(),
		output: make(chan statistics.AttackResult, 100),
		pool:   make(map[uint32]*fcuExecutor),
		wg:     new(sync.WaitGroup),
	}
}

func (commander *fixedConcurrentUsersStrategyCommander) clearDeadExector(id uint32) {
	commander.mu.Lock()
	defer commander.mu.Unlock()
	delete(commander.pool, id)
}

func (commander *fixedConcurrentUsersStrategyCommander) Open(ctx context.Context, task *ultron.Task) <-chan statistics.AttackResult {
	commander.ctx, commander.cancel = context.WithCancel(ctx)
	commander.task = task
	return commander.output
}

func (commander *fixedConcurrentUsersStrategyCommander) Command(d ultron.AttackStrategy, t ultron.Timer) {
	var rampUpSteps []*ultron.RampUpStep

	if commander.describer == nil {
		rampUpSteps = d.Spawn()
	} else {
		rampUpSteps = commander.describer.Switch(d)
	}
	commander.describer = d

	if t == nil {
		commander.timer = ultron.NonstopTimer{}
	} else {
		commander.timer = t
	}
	for _, exe := range commander.pool {
		exe.renewTimer(commander.timer)
	}

	killed := 0
	spawned := 0
	for _, step := range rampUpSteps {
		switch {
		case step.N < 0: // 降压策略
			d := 0

			commander.mu.Lock()
			for _, e := range commander.pool {
				if d > step.N {
					delete(commander.pool, e.id) // 主动清理
					e.kill()
					d--
				} else {
					break
				}
			}
			commander.mu.Unlock()

			killed -= step.N
			log.Info(fmt.Sprintf("killed %d users in ramp-up peroid", killed))
			time.Sleep(step.Interval)

		case step.N > 0: // 增压策略
			for i := 0; i < step.N; i++ {
				id := atomic.AddUint32(&commander.counter, 1) - 1
				executor := newFCUExecutor(id, commander, t)

				commander.mu.Lock()
				commander.pool[id] = executor
				commander.mu.Unlock()

				commander.wg.Add(1)
				go func(exe *fcuExecutor, wg *sync.WaitGroup) {
					defer func() {
						commander.clearDeadExector(exe.id)
						exe.kill() // 所有清理逻辑
						wg.Done()
					}()
					exe.start(commander.ctx, commander.task, commander.output)
				}(executor, commander.wg)
			}
			spawned += step.N
			log.Info(fmt.Sprintf("spawned %d users in ramp-up period", spawned))
			time.Sleep(step.Interval)

		default:
		}
	}
}

func (commander *fixedConcurrentUsersStrategyCommander) Close() {
	commander.cancel()
	commander.wg.Wait()
	close(commander.output)
}

func newFCUExecutor(id uint32, parent *fixedConcurrentUsersStrategyCommander, t ultron.Timer) *fcuExecutor {
	return &fcuExecutor{
		id:    id,
		timer: t,
	}
}

func (e *fcuExecutor) renewTimer(t ultron.Timer) {
	e.mu.Lock()
	defer e.mu.Unlock()
	e.timer = t
}

func (e *fcuExecutor) kill() {
	e.cancel()
}

func (e *fcuExecutor) start(ctx context.Context, task *ultron.Task, output chan<- statistics.AttackResult) {
	if output == nil {
		panic("invalid output channel")
	}

	ctx, e.cancel = context.WithCancel(ctx)

	defer func() {
		if rec := recover(); rec != nil {
			debug.PrintStack()
			log.DPanic("recovered", zap.Any("panic", rec))
		}
	}()

	for {
		select {
		case <-ctx.Done():
			return
		default:
		}

		start := time.Now()
		attacker := task.PickUp()
		err := attacker.Fire(ctx)

		select {
		case output <- statistics.AttackResult{Name: attacker.Name(), Duration: time.Since(start), Error: err}:
		case <-ctx.Done():
		}

		e.mu.RLock()
		t := e.timer
		e.mu.RUnlock()
		t.Sleep()
	}
}
