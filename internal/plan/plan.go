package plan

import (
	"errors"
	"sync"
	"time"

	"github.com/wosai/ultron/v2"
	"github.com/wosai/ultron/v2/pkg/statistics"
)

type (
	Plan struct {
		locked     bool
		current    int
		stages     []ultron.StageConfig
		status     ultron.Status
		stageDatas []stageData
		mu         sync.Mutex
	}

	stageData struct {
		requests uint64
		duration time.Duration
	}

	Status string
)

var _ ultron.Plan = (*Plan)(nil)

func NewPlan() *Plan {
	return &Plan{
		current: -1,
		stages:  make([]ultron.StageConfig, 0),
		status:  ultron.StatusReady,
	}
}

func (p *Plan) addStage(conf ultron.StageConfig) error {
	p.mu.Lock()
	defer p.mu.Unlock()
	if p.locked {
		return errors.New("plan was locked")
	}
	if p.status == ultron.StatusReady {
		p.stages = append(p.stages, conf)
	}
	return nil
}

func (p *Plan) AddStages(sc ...ultron.StageConfig) error {
	for _, conf := range sc {
		if err := p.addStage(conf); err != nil {
			return err
		}
	}
	return nil
}

func (p *Plan) Check() error {
	p.mu.Lock()
	defer p.mu.Unlock()

	if len(p.stages) == 0 {
		return errors.New("empty stage")
	}

	for index, stage := range p.stages {
		if stage.Concurrence == 0 {
			return errors.New("bad concurrence")
		}

		if stage.MinWait < 0 || stage.MaxWait < 0 || stage.MinWait > stage.MaxWait {
			return errors.New("bad min_wait or max_wait")
		}

		// 非最后阶段
		if index < len(p.stages)-1 {
			if stage.Duration == 0 && stage.Requests == 0 {
				return errors.New("cannot break out this stage")
			}
		}
	}
	p.locked = true
	p.stageDatas = make([]stageData, len(p.stages))
	return nil
}

func (p *Plan) StopCurrentAndStartNext(n int, report *statistics.SummaryReport) (stopped bool, stageID int, conf ultron.StageConfig, err error) {
	p.mu.Lock()
	defer p.mu.Unlock()

	if !p.locked {
		return false, n, ultron.StageConfig{}, errors.New("did not check stage configurations yet")
	}

	if p.status == ultron.StatusInterrupted || p.status == ultron.StatusFinished {
		return false, n, ultron.StageConfig{}, ultron.ErrPlanClosed
	}

	if p.status == ultron.StatusRunning {
		if p.current != n {
			return false, n, ultron.StageConfig{}, nil
		}

		stageFinished := p.isFinishedCurrentStage(n, report)
		if !stageFinished { // 该阶段尚未结束，不做任务事情
			return false, n, ultron.StageConfig{}, nil
		}

		if p.current >= len(p.stages)-1 { // 最后一个阶段
			p.status = ultron.StatusFinished
			return true, n, ultron.StageConfig{}, ultron.ErrPlanClosed
		}

		p.current++
		return true, p.current, p.stages[p.current], nil
	}

	if p.status == ultron.StatusReady && p.current == -1 {
		p.status = ultron.StatusRunning
		p.current++
		return true, p.current, p.stages[p.current], nil
	}
	return false, n, ultron.StageConfig{}, errors.New("failed to stop current stage and start next stage")
}

func (p *Plan) isFinishedCurrentStage(n int, report *statistics.SummaryReport) bool {
	totalRequests := report.TotalRequests + report.TotalFailures
	totalDuration := report.LastAttack.Sub(report.FirstAttack)
	var previousRequests, currentStageRequests uint64
	var previousDuration, currentStageDuration time.Duration

	if n > 0 {
		for i := 0; i < n; i++ {
			previousDuration += p.stageDatas[i].duration
			previousRequests += p.stageDatas[i].requests
		}
	}
	currentStageDuration = totalDuration - previousDuration
	currentStageRequests = totalRequests - previousRequests

	if p.stages[n].Duration > 0 && currentStageDuration >= p.stages[n].Duration {
		p.stageDatas[n] = stageData{requests: currentStageRequests, duration: currentStageDuration}
		return true
	}
	if p.stages[n].Requests > 0 && currentStageRequests >= p.stages[n].Requests {
		p.stageDatas[n] = stageData{requests: currentStageRequests, duration: currentStageDuration}
		return true
	}
	return false
}

func (p *Plan) Status() ultron.Status {
	p.mu.Lock()
	defer p.mu.Unlock()

	return p.status
}

func (p *Plan) Stages() []ultron.StageConfig {
	p.mu.Lock()
	defer p.mu.Unlock()

	ret := make([]ultron.StageConfig, len(p.stages))
	copy(ret, p.stages)
	return ret
}
