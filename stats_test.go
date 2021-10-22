package ultron

import (
	"context"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/wosai/ultron/v2/pkg/statistics"
	"go.uber.org/zap"
)

type (
	mockStatsProvider struct {
		id string
	}
)

var (
	_ StatsProvider = (*mockStatsProvider)(nil)
)

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

	// r := rand.Float64()
	// if r <= 0.1 {
	// 	return nil, errors.New("unknown error")
	// }
	sg := statistics.NewStatisticianGroup()
	sg.Record(statistics.AttackResult{Name: "mock test", Duration: 3 * time.Millisecond})
	Logger.Info("provider report", zap.Any("report", sg.Report(true)))
	return sg, nil
}

func TestStatsProvider_Aggregate(t *testing.T) {
	agg := newStatsAggregator()
	for i := 0; i < 10; i++ {
		agg.Add(&mockStatsProvider{id: uuid.New().String()})
	}

	report, err := agg.Aggregate(context.Background(), true)
	assert.Nil(t, err)
	Logger.Info("report", zap.Any("report", report))
}
