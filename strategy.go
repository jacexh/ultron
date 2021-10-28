package ultron

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"runtime/debug"
	"sync"
	"sync/atomic"
	"time"

	"github.com/wosai/ultron/v2/log"
	"github.com/wosai/ultron/v2/pkg/genproto"
	"github.com/wosai/ultron/v2/pkg/statistics"
	"go.uber.org/zap"
)

type (
	// AttackStrategyCommander 压测策略
	AttackStrategyCommander interface {
		Open(context.Context, Task) <-chan statistics.AttackResult
		Command(AttackStrategy, Timer)
		Close()
	}

	// AttackStrategy 压测策略描述
	AttackStrategy interface {
		Spawn() []*RampUpStep
		Switch(next AttackStrategy) []*RampUpStep
		Split(int) []AttackStrategy
	}

	// RampUpStep 增/降压描述
	RampUpStep struct {
		N        int           // 增、降的数量，>0 为加压， <0为降压
		Interval time.Duration // 间隔时间
	}

	// FixedConcurrentUsers 固定goroutine/线程/用户的并发策略
	FixedConcurrentUsers struct {
		ConcurrentUsers int `json:"conncurrent_users"`        // 并发用户数
		RampUpPeriod    int `json:"ramp_up_period,omitempty"` // 增压周期时长
	}

	AttackStrategyConverter struct {
		convertDTOFunc map[string]AttackStrategyConvertDTOFunc
	}

	AttackStrategyConvertDTOFunc func([]byte) (AttackStrategy, error)

	NamedAttackStrategy interface {
		AttackStrategy
		Name() string
	}
)

var (
	_ AttackStrategy = (*FixedConcurrentUsers)(nil)

	DefaultAttackStrategyConverter *AttackStrategyConverter
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
func (fc *FixedConcurrentUsers) Switch(next AttackStrategy) []*RampUpStep {
	n, ok := next.(*FixedConcurrentUsers)
	if !ok {
		panic("cannot switch to different type of AttackStrategyDescriber")
	}
	return fc.spawn(fc.ConcurrentUsers, n.ConcurrentUsers, n.RampUpPeriod, 1)
}

// Split 切分配置
func (fx *FixedConcurrentUsers) Split(n int) []AttackStrategy {
	if n <= 0 {
		panic("bad slices number")
	}
	ret := make([]AttackStrategy, n)
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

func (fx *FixedConcurrentUsers) Name() string {
	return "fixed-concurrent-users"
}

func newAttackStrategyConverter() *AttackStrategyConverter {
	return &AttackStrategyConverter{
		convertDTOFunc: map[string]AttackStrategyConvertDTOFunc{
			"fixed-concurrent-users": func(data []byte) (AttackStrategy, error) {
				as := new(FixedConcurrentUsers)
				err := json.Unmarshal(data, as)
				return as, err
			},
		},
	}
}

func (c *AttackStrategyConverter) ConvertDTO(dto *genproto.AttackStrategyDTO) (AttackStrategy, error) {
	fn, ok := c.convertDTOFunc[dto.Type]
	if !ok {
		return nil, errors.New("cannot found convertion function")
	}
	return fn(dto.AttackStrategy)
}

func (c *AttackStrategyConverter) ConvertAttackStrategy(as AttackStrategy) (*genproto.AttackStrategyDTO, error) {
	na, ok := as.(NamedAttackStrategy)
	if !ok {
		return nil, errors.New("cannot convert attack strategy")
	}
	data, err := json.Marshal(na)
	if err != nil {
		return nil, err
	}
	return &genproto.AttackStrategyDTO{Type: na.Name(), AttackStrategy: data}, nil
}

type (
	fixedConcurrentUsersStrategyCommander struct {
		ctx       context.Context
		cancel    context.CancelFunc
		describer AttackStrategy
		output    chan statistics.AttackResult
		timer     Timer
		task      Task
		counter   uint32
		pool      map[uint32]*fcuExecutor
		wg        *sync.WaitGroup
		mu        sync.Mutex
	}

	// fcuExecutor FixedConcurrentUsers策略的执行者
	fcuExecutor struct {
		id     uint32
		cancel context.CancelFunc
		timer  Timer
		mu     sync.RWMutex
	}
)

var (
	_ AttackStrategyCommander = (*fixedConcurrentUsersStrategyCommander)(nil)
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

func (commander *fixedConcurrentUsersStrategyCommander) Open(ctx context.Context, task Task) <-chan statistics.AttackResult {
	commander.ctx, commander.cancel = context.WithCancel(ctx)
	commander.task = task
	return commander.output
}

func (commander *fixedConcurrentUsersStrategyCommander) Command(d AttackStrategy, t Timer) {
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

func newFCUExecutor(id uint32, parent *fixedConcurrentUsersStrategyCommander, t Timer) *fcuExecutor {
	return &fcuExecutor{
		id:    id,
		timer: t,
	}
}

func (e *fcuExecutor) renewTimer(t Timer) {
	e.mu.Lock()
	defer e.mu.Unlock()
	e.timer = t
}

func (e *fcuExecutor) kill() {
	e.cancel()
}

func (e *fcuExecutor) start(ctx context.Context, task Task, output chan<- statistics.AttackResult) {
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
