package ultron

import (
	"errors"
	"fmt"
	"sync"
	"time"

	"github.com/wosai/ultron/v2/pkg/statistics"
)

type (
	// Plan 定义测试计划接口
	Plan interface {
		Name() string
		AddStages(...Stage)
		Stages() []Stage
		Current() (int, Stage)
		Status() PlanStatus
		// Start() error
		// StopCurrentAndStartNext(int, statistics.SummaryReport) (bool, int, Stage, error)
	}

	// PlanStatus 定义测试计划状态
	PlanStatus int

	plan struct {
		locked       bool
		current      int
		name         string
		stages       []Stage
		status       PlanStatus
		actualStages []*UniversalExitConditions
		mu           sync.Mutex
	}
)

const (
	// StatusReady 测试计划尚未执行
	StatusReady PlanStatus = iota
	// StatusRunning 测试计划执行中
	StatusRunning
	// StatusFinished 测试执行执行完成
	StatusFinished
	// StatusInterrupted 测试计划执行被中断
	StatusInterrupted
)

var (
	ErrPlanClosed      = errors.New("plan was finished or interrupted")
	_             Plan = (*plan)(nil)
)

func NewPlan(name string) *plan {
	if name == "" {
		name = "unknown"
	}
	return &plan{
		name:    name,
		current: -1,
		stages:  make([]Stage, 0),
		status:  StatusReady,
	}
}

func (p *plan) Name() string {
	return p.name
}

func (p *plan) addStage(s Stage) error {
	p.mu.Lock()
	defer p.mu.Unlock()
	if p.locked {
		return errors.New("plan was locked")
	}
	if p.status == StatusReady {
		p.stages = append(p.stages, s)
	}
	return nil
}

func (p *plan) AddStages(stages ...Stage) {
	for _, stage := range stages {
		if err := p.addStage(stage); err != nil {
			panic(err)
		}
	}
}

func (p *plan) interrupt() {
	p.mu.Lock()
	defer p.mu.Unlock()
	if p.status != StatusFinished {
		p.status = StatusInterrupted
	}
}

func (p *plan) check() error {
	p.mu.Lock()
	defer p.mu.Unlock()

	if p.status != StatusReady {
		return fmt.Errorf("cannot start plan in %d status", p.status)
	}

	if len(p.stages) == 0 {
		return errors.New("empty stage")
	}

	for index, stage := range p.stages {
		strategy := stage.GetStrategy()
		switch v := strategy.(type) {
		case *FixedConcurrentUsers:
			if v.ConcurrentUsers <= 0 {
				return errors.New("concurrent users must greater than 0")
			}
		}
		// 非最后阶段
		if index < len(p.stages)-1 {
			if stage.GetExitConditions().NeverStop() {
				return errors.New("cannot break out this stage")
			}
		}
	}
	p.locked = true
	p.actualStages = make([]*UniversalExitConditions, len(p.stages))
	return nil
}

func (p *plan) stopCurrentAndStartNext(n int, report statistics.SummaryReport) (stopped bool, stageID int, s Stage, err error) {
	p.mu.Lock()
	defer p.mu.Unlock()

	if !p.locked {
		panic("did not check stage configurations yet")
	}

	if p.status == StatusInterrupted || p.status == StatusFinished {
		return false, n, nil, ErrPlanClosed
	}

	if p.status == StatusRunning {
		if p.current != n { // stage id不一致，不做任何控制
			return false, n, nil, nil
		}

		stageFinished := p.isFinishedCurrentStage(n, report)
		if !stageFinished { // 该阶段尚未结束，不做任务事情
			return false, n, nil, nil
		}

		if p.current >= len(p.stages)-1 { // 最后一个阶段
			p.status = StatusFinished
			return true, n, nil, ErrPlanClosed
		}

		p.current++
		return true, p.current, p.stages[p.current], nil
	}

	if p.status == StatusReady && p.current == -1 {
		p.status = StatusRunning
		p.current++
		return true, p.current, p.stages[p.current], nil
	}
	return false, n, nil, errors.New("failed to stop current stage and start next stage")
}

func (p *plan) isFinishedCurrentStage(n int, report statistics.SummaryReport) bool {
	totalRequests := report.TotalRequests + report.TotalFailures
	totalDuration := report.LastAttack.Sub(report.FirstAttack)
	var previousRequests, currentStageRequests uint64
	var previousDuration, currentStageDuration time.Duration

	if n > 0 {
		for i := 0; i < n; i++ {
			previousDuration += p.actualStages[i].Duration
			previousRequests += p.actualStages[i].Requests
		}
	}
	currentStageDuration = totalDuration - previousDuration
	currentStageRequests = totalRequests - previousRequests

	// todo 暂时不支持其他ExitConditions
	condition := &UniversalExitConditions{Requests: currentStageRequests, Duration: currentStageDuration}
	if p.stages[n].GetExitConditions().Check(condition) {
		p.actualStages[n] = condition
		return true
	}

	return false
}

func (p *plan) Status() PlanStatus {
	p.mu.Lock()
	defer p.mu.Unlock()

	return p.status
}

func (p *plan) Stages() []Stage {
	p.mu.Lock()
	defer p.mu.Unlock()

	ret := make([]Stage, len(p.stages))
	copy(ret, p.stages)
	return ret
}

func (p *plan) Current() (int, Stage) {
	p.mu.Lock()
	defer p.mu.Unlock()
	if p.current == -1 {
		return -1, nil
	}
	return p.current, p.stages[p.current]
}
