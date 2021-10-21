package ultron

import (
	"context"
	"log"
	"runtime/debug"
	"sync"
	"sync/atomic"
	"time"

	"github.com/wosai/ultron/v2/pkg/statistics"
)

type (
	// AttackStrategyCommander 压测策略
	AttackStrategyCommander interface {
		Open(context.Context, *Task) <-chan *statistics.AttackResult
		Command(AttackStrategyDescriber, Timer)
		Close()
	}

	// AttackStrategyDescriber 压测策略描述
	AttackStrategyDescriber interface {
		Spawn() []*RampUpStep
		Switch(next AttackStrategyDescriber) []*RampUpStep
		Split(int) []AttackStrategyDescriber
	}

	// RampUpStep 增/降压描述
	RampUpStep struct {
		N        int           // 增、降的数量，>0 为加压， <0为降压
		Interval time.Duration // 间隔时间
	}

	// FixedConcurrentUsers 固定goroutine/线程/用户的并发策略
	FixedConcurrentUsers struct {
		ConcurrentUsers int // 并发用户数
		RampUpPeriod    int // 增压周期时长
	}

	fixedConcurrentUsersStrategyCommander struct {
		ctx       context.Context
		cancel    context.CancelFunc
		describer AttackStrategyDescriber
		output    chan *statistics.AttackResult
		timer     Timer
		task      *Task
		counter   uint32
		pool      map[uint32]*fcuExecutor
		wg        sync.WaitGroup
	}

	// fcuExecutor FixedConcurrentUsers策略的执行者
	fcuExecutor struct {
		id        uint32
		cancel    context.CancelFunc
		commander *fixedConcurrentUsersStrategyCommander
	}
)

var (
	_ AttackStrategyDescriber = (*FixedConcurrentUsers)(nil)
	_ AttackStrategyCommander = (*fixedConcurrentUsersStrategyCommander)(nil)
)

func (fc *FixedConcurrentUsers) spawn(current, expected, period, interval int) []*RampUpStep {
	var ret []*RampUpStep

	if current == expected {
		return ret
	}

	if period < interval {
		period = interval
	}

	steps := period / interval
	nPerStep := (expected - current) / steps

	if current < expected {
		for current <= expected-nPerStep {
			current += nPerStep
			ret = append(ret, &RampUpStep{N: nPerStep, Interval: time.Duration(interval) * time.Second})
		}
	} else {
		for current >= expected-nPerStep {
			current += nPerStep
			ret = append(ret, &RampUpStep{N: nPerStep, Interval: time.Duration(interval) * time.Second})
		}
	}

	if current != expected {
		ret[len(ret)-1].N += (expected - current)
	}
	return ret
}

// Spawn 增压、降压
func (fc *FixedConcurrentUsers) Spawn() []*RampUpStep {
	if fc.ConcurrentUsers <= 0 {
		panic("the number of concurrent users must be greater than 0")
	}
	return fc.spawn(0, fc.ConcurrentUsers, fc.RampUpPeriod, 1)
}

// Switch 转入下一个阶段
func (fc *FixedConcurrentUsers) Switch(next AttackStrategyDescriber) []*RampUpStep {
	n, ok := next.(*FixedConcurrentUsers)
	if !ok {
		panic("cannot switch to different type of AttackStrategyDescriber")
	}
	return fc.spawn(fc.ConcurrentUsers, n.ConcurrentUsers, n.RampUpPeriod, 1)
}

// Split 切分配置
func (fx *FixedConcurrentUsers) Split(n int) []AttackStrategyDescriber {
	if n <= 0 {
		panic("bad slices number")
	}
	ret := make([]AttackStrategyDescriber, n)
	for i := 0; i < n; i++ {
		ret[i] = &FixedConcurrentUsers{
			ConcurrentUsers: fx.ConcurrentUsers / n,
			RampUpPeriod:    fx.RampUpPeriod,
		}
	}
	if remainder := fx.ConcurrentUsers % n; remainder > 0 {
		for i := 0; i < remainder; i++ {
			ret[i].(*FixedConcurrentUsers).ConcurrentUsers++
		}
	}
	return ret
}

func newFCUExecutor(id uint32, parent *fixedConcurrentUsersStrategyCommander) *fcuExecutor {
	return &fcuExecutor{
		id:        id,
		commander: parent,
	}
}

func (e *fcuExecutor) kill() {
	log.Printf("executor-%d is quit\n", e.id)
	e.cancel()
	e.commander.clearDeadExector(e.id)
}

func (e *fcuExecutor) start(ctx context.Context, task *Task, output chan<- *statistics.AttackResult) {
	if output == nil {
		panic("invalid output channel")
	}
	e.commander.wg.Add(1)
	defer e.commander.wg.Done()

	ctx, e.cancel = context.WithCancel(ctx)

	defer func() {
		if rec := recover(); rec != nil {
			debug.PrintStack()
			// todo
		}
		e.kill()
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
		case output <- &statistics.AttackResult{Name: attacker.Name(), Duration: time.Since(start), Error: err}:
		case <-ctx.Done():
		case <-time.After(3 * time.Second): // 这里是channel非常不好的使用方式，通常情况下，channel只允许一个对象写入，这样就可以保证安全的关闭；
			//在当前场景下，channel存在多个写入方，可能会造成channel不安全的关闭从而阻塞进程
			panic("output channel may closed")
		}

		e.commander.timer.Sleep()
	}
}

func newFixedConcurrentUsersStrategyCommander() *fixedConcurrentUsersStrategyCommander {
	return &fixedConcurrentUsersStrategyCommander{
		ctx:    context.TODO(),
		output: make(chan *statistics.AttackResult, 100),
		pool:   make(map[uint32]*fcuExecutor),
	}
}

func (commander *fixedConcurrentUsersStrategyCommander) clearDeadExector(id uint32) {
	delete(commander.pool, id)
}

func (commander *fixedConcurrentUsersStrategyCommander) Open(ctx context.Context, task *Task) <-chan *statistics.AttackResult {
	commander.ctx, commander.cancel = context.WithCancel(ctx)
	commander.task = task
	return commander.output
}

func (commander *fixedConcurrentUsersStrategyCommander) Command(d AttackStrategyDescriber, t Timer) {
	var rampUpSteps []*RampUpStep
	if commander.describer == nil {
		rampUpSteps = d.Spawn()
	} else {
		rampUpSteps = commander.describer.Switch(d)
	}
	commander.describer = d

	if t == nil {
		commander.timer = NonstopTimer{}
	} else {
		commander.timer = t
	}

	killed := 0
	spawned := 0
	for _, step := range rampUpSteps {
		switch {
		case step.N < 0: // 降压策略
			d := 0
			for _, e := range commander.pool {
				if d > step.N {
					e.kill()
					d--
				} else {
					break
				}
			}
			killed -= step.N
			log.Printf("killed %d users\n", killed)
			time.Sleep(step.Interval)

		case step.N > 0: // 增压策略
			for i := 0; i < step.N; i++ {
				id := atomic.AddUint32(&commander.counter, 1) - 1
				executor := newFCUExecutor(id, commander)
				commander.pool[id] = executor

				go executor.start(commander.ctx, commander.task, commander.output)
			}
			spawned += step.N
			log.Printf("spawn %d users\n", spawned)
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
