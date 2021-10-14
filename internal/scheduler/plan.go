package scheduler

import (
	"errors"
	"sync/atomic"
	"time"
)

type (
	Plan struct {
		stageCount   int
		currentStage int
		status       Status
		stages       []*stage
	}

	Status = uint32

	stage struct {
		index    int
		since    time.Time
		status   Status
		config   StageConfiguration
		previous *stage
	}

	StageConfiguration struct {
		Duration    time.Duration // 阶段持续时间，不严格控制
		Requests    uint64        // 阶段请求总数，不严格控制
		Concurrence uint32        // 阶段目标并发数
		HatchRate   int32         // 进入该阶段时，每秒增压、降压数目。
		MinWait     time.Duration // 最小等待时间
		MaxWait     time.Duration // 最大等待时间
	}
)

const (
	StatusReady Status = iota
	StatusRunning
	StatusFinished
	StatusInterrupted
)

func NewPlan() *Plan {
	return &Plan{
		stageCount:   0,
		currentStage: -1,
		status:       StatusReady,
		stages:       make([]*stage, 0),
	}
}

// Status 获取当前计划的状态
func (p *Plan) Status() Status {
	return atomic.LoadUint32(&p.status)
}

func (p *Plan) Start() bool {
	return atomic.CompareAndSwapUint32(&p.status, StatusReady, StatusRunning)
}

func (p *Plan) Finish() bool {
	return atomic.CompareAndSwapUint32(&p.status, StatusRunning, StatusFinished)
}

func (p *Plan) Interrupt() bool {
	return atomic.CompareAndSwapUint32(&p.status, StatusRunning, StatusInterrupted)
}

// checkStageConfigurations
func (p *Plan) checkStageConfigurations() error {
	for index, stage := range p.stages {
		if stage.config.Concurrence == 0 {
			return errors.New("bad concurrence")
		}

		if stage.config.MinWait < 0 || stage.config.MaxWait < 0 || stage.config.MinWait > stage.config.MaxWait {
			return errors.New("bad min_wait or max_wait")
		}

		// 非最后阶段
		if index < len(p.stages)-1 {
			if stage.config.Duration == 0 && stage.config.Requests == 0 {
				return errors.New("cannot break currest stage")
			}
		}
	}
	return nil
}

func (p *Plan) addStage(conf StageConfiguration) {
	previousStageIndex := p.stageCount - 1
	stage := &stage{
		index:  previousStageIndex + 1,
		config: conf,
	}
	if previousStageIndex >= 0 {
		stage.previous = p.stages[previousStageIndex]
	}
	p.stages = append(p.stages, stage)
	p.stageCount++
}

func (p *Plan) AddStages(cs ...StageConfiguration) {
	for _, conf := range cs {
		p.addStage(conf)
	}
}

// Stages 获取本执行计划的所有阶段
func (p *Plan) Stages() []StageConfiguration {
	sc := make([]StageConfiguration, len(p.stages))
	for index, stage := range p.stages {
		sc[index] = stage.config
	}
	return sc
}

// func (p *Plan) RunNextStage(i int, count uint64) (int, StageConfiguration, bool) {

// }

func (sc StageConfiguration) Split(n int) []StageConfiguration {
	if n == 0 {
		panic(errors.New("bad slices number"))
	}
	ret := make([]StageConfiguration, n)
	// 先处理不切分的配置
	for i := 0; i < n; i++ {
		ret[i] = StageConfiguration{
			Duration:    sc.Duration,
			Requests:    sc.Requests / uint64(n),
			Concurrence: sc.Concurrence / uint32(n),
			HatchRate:   sc.HatchRate / int32(n),
			MinWait:     sc.MinWait,
			MaxWait:     sc.MaxWait,
		}
	}

	if remainder := sc.Requests % uint64(n); remainder > 0 {
		for i := 0; i < int(remainder); i++ {
			ret[i].Requests++
		}
	}

	if remainder := sc.Concurrence % uint32(n); remainder > 0 {
		for i := 0; i < int(remainder); i++ {
			ret[i].Concurrence++
		}
	}

	if remainder := sc.HatchRate % int32(n); remainder > 0 {
		for i := 0; i < int(remainder); i++ {
			ret[i].HatchRate++
		}
	}
	return ret
}
