package types

import (
	"context"
	"time"
)

type (
	// Attacker 事务接口定义，需要在实现上保证其goroutine-safe
	Attacker interface {
		Name() string
		Fire(context.Context) error
	}

	Task interface {
		Add(Attacker, int)
		PickUp() Attacker
	}

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
		AddStages(...StageConfig)
		Stages() []StageConfig
		Status() Status
		Check() error
		FinishAndStartNextStage(int) (int, StageConfig, error)
	}

	MasterRunner interface {
		WithPlane(Plan)
	}

	SlaveRunner interface {
		WithTask(Task)
	}

	Runner interface {
		MasterRunner
		SlaveRunner
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
)

const (
	StatusReady Status = iota
	StatusRunning
	StatusFinished
	StatusInterrupted
)
