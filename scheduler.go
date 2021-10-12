package ultron

import "sync/atomic"

type (
	// Scheduler 调度器对象，用于master进程中的计划、节点管理
	Scheduler struct {
		plane  *Plan
		slaves map[string]Slave
	}

	// Slave 节点对象
	// todo: 需要分别实现LocalSlave、RemoteSlave
	Slave interface{}

	// Plan 执行计划
	Plan struct {
		ID     string
		stags  interface{}
		status uint32
	}

	PlanStatus = uint32
)

const (
	PlanReady        PlanStatus = iota // 计划可运行
	PlanRunning                        // 计划运行中
	PlanFinished                       // 计划正常结束
	PlaneInterrupted                   // 技术被中断
)

func (p *Plan) CurrentStatus() PlanStatus {
	return atomic.LoadUint32(&p.status)
}

func (p *Plan) Start() bool {
	return atomic.CompareAndSwapUint32(&p.status, PlanReady, PlanRunning)
}

func (p *Plan) Finish() bool {
	return atomic.CompareAndSwapUint32(&p.status, PlanRunning, PlanFinished)
}

func (p *Plan) Interrupt() {
	atomic.StoreUint32(&p.status, PlaneInterrupted)
}
