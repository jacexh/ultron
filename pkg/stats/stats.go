package stats

import (
	"sync"
	"time"
)

var (
	timeDistributions = [10]float64{0.5, 0.6, 0.7, 0.8, 0.9, 0.95, 0.97, 0.98, 0.99, 1.0}
)

type (
	Statistician struct {
	}

	ResultAggregator struct {
		name              string
		requests          uint64
		failures          uint64
		totalResponseTime time.Duration
		minResponseTime   time.Duration
		maxResponseTime   time.Duration
		trendSuccess      interface{}
		trendFailures     interface{}
		responseBucket    map[time.Duration]uint64
		failureBucket     map[string]uint64
		since             time.Time
		lastAttack        time.Time
		interval          time.Duration
		mu                sync.RWMutex
	}

	// AggregatedReport 聚合报告
	AggregatedReport struct {
		Name           string                   // 事务名称
		Requests       uint64                   // 成功请求总数
		Failures       uint64                   // 失败请求总数
		Min            time.Duration            // 最小延迟
		Max            time.Duration            // 最大延迟
		Median         time.Duration            // 中位数
		Average        time.Duration            // 平均数
		TPS            uint64                   // 每秒事务数
		Distributions  map[string]time.Duration // 百分位分布
		FailRation     float64                  // 错误:率
		FailureDetails map[string]int32         // 错误详情分布
		FullHistory    bool                     // 是否是该阶段完整的报告
	}

	SummaryReport map[string]AggregatedReport
)
