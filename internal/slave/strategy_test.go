package slave

import (
	"context"
	"net/http"
	"sync"
	"testing"
	"time"

	"github.com/wosai/ultron/v2"
	"github.com/wosai/ultron/v2/log"
	"github.com/wosai/ultron/v2/pkg/statistics"
	"go.uber.org/zap"
)

type (
	benchmarkAttacker struct {
		name string
		wait time.Duration
	}
)

type fakeAttacker struct{}

func (fs *fakeAttacker) Name() string {
	return "fake"
}

func (fs *fakeAttacker) Fire(ctx context.Context) error {
	req, _ := http.NewRequest(http.MethodGet, "https://www.google.com", nil)
	_ = req.WithContext(ctx)
	select {
	case <-ctx.Done():
		return ctx.Err()
	default:
	}
	return nil
}

func newBenchmarkAttacker(n string, wait time.Duration) ultron.Attacker {
	return &benchmarkAttacker{name: n, wait: wait}
}

func (b *benchmarkAttacker) Name() string {
	return b.name
}

func (b *benchmarkAttacker) Fire(ctx context.Context) error {
	select {
	case <-ctx.Done():
		return ctx.Err()
	default:
	}
	time.Sleep(b.wait)
	return nil
}

func TestFCUExecutor(t *testing.T) {
	commander := newFixedConcurrentUsersStrategyCommander()
	task := ultron.NewTask()
	task.Add(&fakeAttacker{}, 10)

	output := commander.Open(context.Background(), task)
	go func() {
		for range output {
		}
	}()

	commander.Command(&ultron.FixedConcurrentUsers{ConcurrentUsers: 50, RampUpPeriod: 3}, ultron.NonstopTimer{})
	<-time.After(2 * time.Second)
	commander.Command(&ultron.FixedConcurrentUsers{ConcurrentUsers: 80, RampUpPeriod: 5}, ultron.NonstopTimer{})
	<-time.After(2 * time.Second)
	commander.Command(&ultron.FixedConcurrentUsers{ConcurrentUsers: 30, RampUpPeriod: 7}, ultron.NonstopTimer{})
	commander.Close()
}

func TestFCUSBenchmark(t *testing.T) {
	commander := newFixedConcurrentUsersStrategyCommander()
	task := ultron.NewTask()
	task.Add(newBenchmarkAttacker("benchmark", 1*time.Millisecond), 10)
	sg := statistics.NewStatisticianGroup()
	output := commander.Open(context.Background(), task)

	var wg sync.WaitGroup
	wg.Add(1)
	go func() {
		defer wg.Done()
		for result := range output {
			sg.Record(result)
		}
	}()

	commander.Command(&ultron.FixedConcurrentUsers{ConcurrentUsers: 100, RampUpPeriod: 0}, ultron.NonstopTimer{})
	<-time.After(5 * time.Second)
	commander.Close()
	wg.Wait()

	report := sg.Report(true) // tps理论最大值10000, 1.6.0该配置均值在8000左右
	log.Info("report", zap.Float64("tps", report.TotalTPS), zap.Time("first_attack", report.FirstAttack), zap.Time("last_attack", report.LastAttack))
}
