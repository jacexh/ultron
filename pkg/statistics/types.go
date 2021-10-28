package statistics

import "context"

type (
	// ReportHandleFunc 聚合报告处理函数
	ReportHandleFunc func(context.Context, SummaryReport)

	// ReportBus 聚合报告事件总线
	ReportBus interface {
		SubscribeReport(ReportHandleFunc)
		PublishReport(SummaryReport)
	}
)

type (

	// ResultHandleFunc 请求结果处理函数
	ResultHandleFunc func(context.Context, AttackResult)

	// ResultBus 压测结果事件总线
	ResultBus interface {
		SubscribeResult(ResultHandleFunc)
		PublishResult(AttackResult)
	}
)
