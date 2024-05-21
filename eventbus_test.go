package ultron

import (
	"context"
	"sync/atomic"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/wosai/ultron/v2/pkg/statistics"
)

func testResultHandleFunc(ctx context.Context, ret statistics.AttackResult) {
	select {
	case <-ctx.Done():
	default:
	}
}

func BenchmarkIEventBus_PublishResult(b *testing.B) {
	bus := defaultEventBus
	bus.subscribeResult(testResultHandleFunc)
	go bus.start()

	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			bus.publishResult(statistics.AttackResult{Name: "benchmark", Duration: 100 * time.Millisecond})
		}
	})
}

func TestEventBus_report(t *testing.T) {
	eb := newEventBus()
	eb.start()

	var called int32
	eb.subscribeReport(func(c context.Context, sr statistics.SummaryReport) {
		atomic.AddInt32(&called, 1)
	})
	eb.publishReport(statistics.SummaryReport{})

	eb.close()
	current := atomic.LoadInt32(&called)
	assert.EqualValues(t, current, 1)
}

func TestEventBus_result(t *testing.T) {
	eb := newEventBus()
	eb.start()

	var called int32
	eb.subscribeResult(func(c context.Context, ar statistics.AttackResult) {
		atomic.AddInt32(&called, 1)
	})
	eb.publishResult(statistics.AttackResult{Name: "unittest", Duration: 10 * time.Millisecond})
	eb.close()

	current := atomic.LoadInt32(&called)
	assert.EqualValues(t, current, 1)
}
