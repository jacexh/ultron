package types

import (
	"context"
	"errors"
	"time"

	"github.com/wosai/ultron/pkg/statistics"
)

type (
	// StageConfig 测试阶段的描述
	StageConfig struct {
		Duration    time.Duration `json:"duration,omitempty"`   // 阶段持续时间
		Requests    uint64        `json:"requests,omitempty"`   // 阶段请求总数
		Concurrence int           `json:"concurrence"`          // 阶段目标并发数
		HatchRate   int           `json:"hatch_rate,omitempty"` // 阶段孵化期每秒加压、减压数目
		MinWait     time.Duration `json:"min_wait,omitempty"`   // Attacker之间最小等待时间
		MaxWait     time.Duration `json:"max_wait,omitempty"`   // Attacker之间最长等待时间
	}

	Status int

	Plan interface {
		AddStages(...StageConfig) error
		Stages() []StageConfig
		Status() Status
		Check() error
		StopCurrentAndStartNext(int, *statistics.SummaryReport) (bool, int, StageConfig, error)
	}

	MasterRunner interface {
		WithPlane(Plan)
	}

	Runner interface {
		MasterRunner
	}

	// EventType 事件类型
	EventType string
	// Event 事件接口定义
	Event interface {
		Type() EventType
	}
	// EventHandler 事件处理器
	EventHandler interface {
		Subscribe(EventType)           // 订阅
		Handle(context.Context, Event) // 处理事件
	}

	Schedulable interface {
		Interrupt() error
		Finish() error
	}

	ReportHandleFunc func(context.Context, *statistics.SummaryReport, statistics.Tags)
)

const (
	StatusReady Status = iota
	StatusRunning
	StatusFinished
	StatusInterrupted
)

var (
	ErrPlanClosed = errors.New("plan was finished or interrupted")
)
