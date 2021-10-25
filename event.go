package ultron

import (
	"context"

	"github.com/wosai/ultron/v2/pkg/statistics"
)

type (
	// EventType 事件类型
	EventType string

	// Event 定义Event接口
	Event interface {
		Type() EventType
	}
	// EventHandleFunc 通用消息处理函数
	EventHandleFunc func(context.Context, Event)
	// EventBus 事件总线
	EventBus interface {
		Subscribe(EventType, EventHandleFunc)
		Publish(Event)
	}
)

type (
	// ReportHandleFunc 聚合报告处理函数
	ReportHandleFunc func(context.Context, statistics.SummaryReport)
	// ReportBus 聚合报告事件总线
	ReportBus interface {
		SubscribeReport(ReportHandleFunc)
		PublishReport(statistics.SummaryReport)
	}
)

type (
	// ResultHandleFunc 请求结果处理函数
	ResultHandleFunc func(context.Context, statistics.AttackResult)

	// ResultBus 压测结果事件总线
	ResultBus interface {
		SubscribeResult(ResultHandleFunc)
		PublishResult(statistics.AttackResult)
	}
)
