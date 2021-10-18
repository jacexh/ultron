package scheduler

import (
	"context"
	"errors"

	"github.com/wosai/ultron/pkg/statistics"
	"github.com/wosai/ultron/types"
)

type (
	// Scheduler 全局调度对象，负责计划、节点(Slave)的生命周期
	Scheduler struct{}

	Stats interface {
		Record(*statistics.AttackResut)
		Report(bool) *statistics.SummaryReport
		Start(context.Context, string)
		NexStage(string, int)
		Upload(string, int, int, *statistics.AttackStatistician)
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
