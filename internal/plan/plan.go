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
		mu      sync.Mutex
	}

	Status string
)

const (
	StatusReady       = "ready"
	StatueRunning     = "running"
	StatueFinished    = "finished"
	StatusInterrupted = "interrupted"
)

var _ types.Plan = (*Plan)(nil)

func NewPlan() *Plan {
	return &Plan{
		current: -1,
		stages:  make([]types.StageConfig, 0),
	}
}

func (p *Plan) addStage(conf types.StageConfig) {
	p.mu.Lock()
	defer p.mu.Unlock()

	p.stages = append(p.stages, conf)
}

func (p *Plan) AddStages(sc ...types.StageConfig) {
	for _, conf := range sc {
		p.addStage(conf)
	}
}

func (p *Plan) check() error {
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
				return errors.New("cannot break currest stage")
			}
		}
	}
	return nil
}

func (p *Plan) startNextStage() (bool, int, types.StageConfig) {
	p.mu.Lock()
	defer p.mu.Unlock()

	if p.current < len(p.stages)-1 {
		p.current++
		return true, p.current, p.stages[p.current]
	}
	return false, p.current, p.stages[p.current]
}

func (p *Plan) finishStage(n int) error {
	p.mu.Lock()
	defer p.mu.Unlock()

	switch {
	case n < p.current:
		return errors.New("stage was already finished")

	case n > p.current:
		return errors.New("unknown stage")

	default:
		return nil
	}
}
