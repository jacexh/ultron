package scheduler

import (
	"context"
	"errors"
	"sync"

	"github.com/wosai/ultron/pkg/proto"
	"github.com/wosai/ultron/pkg/statistics"
	"github.com/wosai/ultron/types"
)

type (
	// Scheduler 全局调度对象，负责计划、节点(Slave)的生命周期
	Scheduler struct {
		batch uint32
		plan  types.Plan
		agg   StatsAggregator
		// slaves map[string]types.SlaveRunner
		server proto.UltronServiceServer
		mu     sync.Mutex
	}

	StatsAggregator interface {
		Aggregate(ctx context.Context, c chan<- *statistics.SummaryReport)
		Start(...StatsProvider)
		Stop(ctx context.Context, c chan<- *statistics.SummaryReport)
	}

	StatsProvider interface {
		ID() string
		Provide(stage int, batch int) *statistics.StatisticianGroup
	}

	StatsReporter interface {
		Report(bool) *statistics.SummaryReport
	}

	StatsRecorder interface {
		Record(*statistics.AttackResult)
	}
)

func SplitStageConfiguration(sc types.StageConfig, n int) []types.StageConfig {
	if n == 0 {
		panic(errors.New("bad slices number"))
	}
	ret := make([]types.StageConfig, n)
	// 先处理不切分的配置
	for i := 0; i < n; i++ {
		ret[i] = types.StageConfig{
			Duration:    sc.Duration,
			Requests:    sc.Requests / uint64(n),
			Concurrence: sc.Concurrence / n,
			HatchRate:   sc.HatchRate / n,
			MinWait:     sc.MinWait,
			MaxWait:     sc.MaxWait,
		}
	}

	if remainder := sc.Requests % uint64(n); remainder > 0 {
		for i := 0; i < int(remainder); i++ {
			ret[i].Requests++
		}
	}

	if remainder := sc.Concurrence % n; remainder > 0 {
		for i := 0; i < int(remainder); i++ {
			ret[i].Concurrence++
		}
	}

	if remainder := sc.HatchRate % n; remainder > 0 {
		for i := 0; i < int(remainder); i++ {
			ret[i].HatchRate++
		}
	}
	return ret
}
