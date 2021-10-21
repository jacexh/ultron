package ultron

import (
	"context"
	"errors"

	"github.com/wosai/ultron/v2/pkg/statistics"
)

type (
	// StageConfig 测试阶段的描述

	Status int

	// MasterRunner interface {
	// 	WithPlane(Plan)
	// }

	// Runner interface {
	// 	MasterRunner
	// }

	// // EventType 事件类型
	// EventType string
	// // Event 事件接口定义
	// Event interface {
	// 	Type() EventType
	// }
	// // EventHandler 事件处理器
	// EventHandler interface {
	// 	Subscribe(EventType)           // 订阅
	// 	Handle(context.Context, Event) // 处理事件
	// }

	// Schedulable interface {
	// 	Interrupt() error
	// 	Finish() error
	// }

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
