package plan

import (
	"errors"
	"sync"

	"github.com/wosai/ultron/types"
)

type (
	Plan struct {
		current int
		stages  []types.StageConfig
		status  types.Status
		mu      sync.Mutex
	}

	Status string
)

var _ types.Plan = (*Plan)(nil)

func NewPlan() *Plan {
	return &Plan{
		current: -1,
		stages:  make([]types.StageConfig, 0),
		status:  types.StatusReady,
	}
}

func (p *Plan) addStage(conf types.StageConfig) {
	p.mu.Lock()
	defer p.mu.Unlock()
	if p.status == types.StatusReady {
		p.stages = append(p.stages, conf)
	}
}

func (p *Plan) AddStages(sc ...types.StageConfig) {
	for _, conf := range sc {
		p.addStage(conf)
	}
}

func (p *Plan) Check() error {
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
	return nil
}

func (p *Plan) FinishAndStartNextStage(n int) (int, types.StageConfig, error) {
	p.mu.Lock()
	defer p.mu.Unlock()

	if p.status == types.StatusInterrupted || p.status == types.StatusFinished {
		return n, types.StageConfig{}, errors.New("plan was finished or interrupted")
	}

	if p.status == types.StatusRunning {
		if p.current != n {
			return n, types.StageConfig{}, errors.New("failed to finish stage because invalid stage id was provided")
		}

		if p.current >= len(p.stages)-1 { // 最后一个阶段
			p.status = types.StatusFinished
			return n, types.StageConfig{}, errors.New("the plan is finished")
		}

		p.current++
		return p.current, p.stages[p.current], nil
	}

	if p.status == types.StatusReady && p.current == -1 {
		p.status = types.StatusRunning
		p.current++
		return p.current, p.stages[p.current], nil
	}
	return n, types.StageConfig{}, errors.New("failed to finish stage and start next stage")
}

func (p *Plan) Status() types.Status {
	p.mu.Lock()
	defer p.mu.Unlock()

	return p.status
}

func (p *Plan) Stages() []types.StageConfig {
	p.mu.Lock()
	defer p.mu.Unlock()

	ret := make([]types.StageConfig, len(p.stages))
	copy(ret, p.stages)
	return ret
}
