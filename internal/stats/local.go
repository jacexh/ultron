package stats

import (
	"context"

	"github.com/wosai/ultron/v2/internal/scheduler"
	"github.com/wosai/ultron/v2/pkg/statistics"
)

type (
	LocalStatsAggregator struct {
		*statistics.StatisticianGroup
	}

	RemoteStatsAggregator struct {
		batch  uint32
		slaves []scheduler.StatsProvider
	}
)

var _ scheduler.StatsAggregator = (*LocalStatsAggregator)(nil)

func (lsa LocalStatsAggregator) Start(slave ...scheduler.StatsProvider) {
	lsa.Reset()
}

func (lsa LocalStatsAggregator) Aggregate(ctx context.Context, input chan<- *statistics.SummaryReport) {
	input <- lsa.Report(false)
}

func (lsa LocalStatsAggregator) Stop(ctx context.Context, input chan<- *statistics.SummaryReport) {
	input <- lsa.Report(true)
}

func (rsa *RemoteStatsAggregator) Aggregate(ctx context.Context, input chan<- *statistics.SummaryReport) {
}
