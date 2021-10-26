package master

import (
	"errors"
	"fmt"
	"sync"
	"time"

	"github.com/wosai/ultron/v2"
	"github.com/wosai/ultron/v2/pkg/statistics"
)

type (
	plan struct {
		locked       bool
		current      int
		stages       []ultron.Stage
		status       ultron.PlanStatus
		actualStages []ultron.UniversalExitConditions
		mu           sync.Mutex
	}
)

var (
	_ ultron.Plan = (*plan)(nil)
)

func NewPlan() ultron.Plan {
	return &plan{
		current: -1,
		stages:  make([]ultron.Stage, 0),
		status:  ultron.StatusReady,
	}
}

func (p *plan) addStage(s ultron.Stage) error {
	p.mu.Lock()
	defer p.mu.Unlock()
	if p.locked {
		return errors.New("plan was locked")
	}
	if p.status == ultron.StatusReady {
		p.stages = append(p.stages, s)
	}
	return nil
}

func (p *plan) AddStages(stages ...ultron.Stage) {
	for _, stage := range stages {
		if err := p.addStage(stage); err != nil {
			panic(err)
		}
	}
}

func (p *plan) Start() error {
	p.mu.Lock()
	defer p.mu.Unlock()

	if p.status != ultron.StatusReady {
		return fmt.Errorf("cannot start plan in %d status", p.status)
	}

	if len(p.stages) == 0 {
		return errors.New("empty stage")
	}

	for index, stage := range p.stages {
		// 非最后阶段
		if index < len(p.stages)-1 {
			if stage.GetExitConditions().NeverStop() {
				return errors.New("cannot break out this stage")
			}
		}
	}
	p.locked = true
	p.actualStages = make([]ultron.UniversalExitConditions, len(p.stages))
	return nil
}

func (p *plan) StopCurrentAndStartNext(n int, report statistics.SummaryReport) (stopped bool, stageID int, s ultron.Stage, err error) {
	p.mu.Lock()
	defer p.mu.Unlock()

	if !p.locked {
		panic("did not check stage configurations yet")
	}

	if p.status == ultron.StatusInterrupted || p.status == ultron.StatusFinished {
		return false, n, nil, ultron.ErrPlanClosed
	}

	if p.status == ultron.StatusRunning {
		if p.current != n { // stage id不一致，不做任何控制
			return false, n, nil, nil
		}

		stageFinished := p.isFinishedCurrentStage(n, report)
		if !stageFinished { // 该阶段尚未结束，不做任务事情
			return false, n, nil, nil
		}

		if p.current >= len(p.stages)-1 { // 最后一个阶段
			p.status = ultron.StatusFinished
			return true, n, nil, ultron.ErrPlanClosed
		}

		p.current++
		return true, p.current, p.stages[p.current], nil
	}

	if p.status == ultron.StatusReady && p.current == -1 {
		p.status = ultron.StatusRunning
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
	condition := ultron.UniversalExitConditions{Requests: currentStageRequests, Duration: currentStageDuration}
	if p.stages[n].GetExitConditions().Check(condition) {
		p.actualStages[n] = condition
		return true
	}

	return false
}

func (p *plan) Status() ultron.PlanStatus {
	p.mu.Lock()
	defer p.mu.Unlock()

	return p.status
}

func (p *plan) Stages() []ultron.Stage {
	p.mu.Lock()
	defer p.mu.Unlock()

	ret := make([]ultron.Stage, len(p.stages))
	copy(ret, p.stages)
	return ret
}

func (p *plan) Current() (int, ultron.Stage) {
	p.mu.Lock()
	defer p.mu.Unlock()

	return p.current, p.stages[p.current]
}
