package ultron

import (
	"errors"

	"github.com/wosai/ultron/v2/pkg/statistics"
)

type (
	// Plan 定义测试计划接口
	Plan interface {
		AddStages(...Stage)
		Stages() []Stage
		Current() (int, Stage)
		Status() PlanStatus
		Start() error
		StopCurrentAndStartNext(int, statistics.SummaryReport) (bool, int, Stage, error)
	}

	// PlanStatus 定义测试计划状态
	PlanStatus int
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
	ErrPlanClosed = errors.New("plan was finished or interrupted")
)
