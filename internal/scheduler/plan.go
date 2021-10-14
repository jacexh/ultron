package scheduler

import (
	"sync/atomic"
	"time"
)

type (
	Plan struct {
		status PlanStatus
		stages []*Stage
	}

	PlanStatus = uint32

	Stage struct {
		counts    uint64
		startedAt time.Time
	}
)

const (
	PlanReady PlanStatus = iota
	PlanRunning
	PlanFinished
	PlanInterrupted
)

// Status 获取当前计划的状态
func (p *Plan) Status() PlanStatus {
	return atomic.LoadUint32(&p.status)
}

func (p *Plan) Start() bool {
	return atomic.CompareAndSwapUint32(&p.status, PlanReady, PlanRunning)
}

func (p *Plan) Finish() bool {
	return atomic.CompareAndSwapUint32(&p.status, PlanRunning, PlanFinished)
}

func (p *Plan) Interrupt() bool {
	return atomic.CompareAndSwapUint32(&p.status, PlanRunning, PlanInterrupted)
}

func (p *Plan) CurrentStageConfig() {

}

func (p *Plan) FinishAndGetNextStage() (finished bool, stageSeq int, stage *Stage) {
	return finished, stageSeq, stage
}

func (s *Stage) IsFinished() bool {
	return false
}
