package master

import (
	"context"
	"time"

	"github.com/wosai/ultron/v2/pkg/statistics"
)

type (
	mockStatsProvider struct {
		id string
	}
)

// var (
// 	_ ultron.StatsProvider = (*mockStatsProvider)(nil)
// )

func (mock *mockStatsProvider) ID() string {
	return mock.id
}

func (mock *mockStatsProvider) Submit(ctx context.Context, batch uint32) (*statistics.StatisticianGroup, error) {
	time.Sleep(10 * time.Millisecond)
	select {
	case <-ctx.Done():
		return nil, ctx.Err()
	default:
	}

	sg := statistics.NewStatisticianGroup()
	sg.Record(statistics.AttackResult{Name: "mock test", Duration: 3 * time.Millisecond})
	return sg, nil
}

// func TestStatsProvider_Aggregate(t *testing.T) {
// 	agg := newSlaveSupervisor()
// 	for i := 0; i < 10; i++ {
// 		agg.Add(&mockStatsProvider{id: uuid.New().String()})
// 	}

// 	report, err := agg.Aggregate(true)
// 	assert.Nil(t, err)
// 	ultron.Logger.Info("report", zap.Any("report", report))
// }
