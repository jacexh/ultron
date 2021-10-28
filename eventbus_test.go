package ultron

import (
	"context"
	"testing"
	"time"

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
