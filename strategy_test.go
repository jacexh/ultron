package ultron

import (
	"context"
	"net/http"
	"sync"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/wosai/ultron/pkg/statistics"
	"go.uber.org/zap"
)

func TestFixedConcurrentUsers_Spawn(t *testing.T) {
	s := &FixedConcurrentUsers{
		ConcurrentUsers: 1000,
		RampUpPeriod:    3,
	}
	waves := s.Spawn()
	assert.EqualValues(t, waves, []*RampUpStep{
		{N: 333, Interval: 1 * time.Second},
		{N: 333, Interval: 1 * time.Second},
		{N: 334, Interval: 1 * time.Second},
	})
}

func TestFixedConcurrentUsers_Switch(t *testing.T) {
	s1 := &FixedConcurrentUsers{
		ConcurrentUsers: 1000,
		RampUpPeriod:    3,
	}
	s2 := &FixedConcurrentUsers{
		ConcurrentUsers: 600,
		RampUpPeriod:    6,
	}
	waves := s1.Switch(s2)
	assert.EqualValues(t, waves, []*RampUpStep{
		{N: -66, Interval: 1 * time.Second},
		{N: -66, Interval: 1 * time.Second},
		{N: -66, Interval: 1 * time.Second},
		{N: -66, Interval: 1 * time.Second},
		{N: -66, Interval: 1 * time.Second},
		{N: -70, Interval: 1 * time.Second},
	})
}

func TestFixedConcurrentUsers_Spilt(t *testing.T) {
	fx := &FixedConcurrentUsers{
		ConcurrentUsers: 1000,
		RampUpPeriod:    3,
	}
	subs := fx.Split(3)
	assert.EqualValues(t, subs, []AttackStrategy{
		&FixedConcurrentUsers{ConcurrentUsers: 334, RampUpPeriod: 3},
		&FixedConcurrentUsers{ConcurrentUsers: 333, RampUpPeriod: 3},
		&FixedConcurrentUsers{ConcurrentUsers: 333, RampUpPeriod: 3},
	})
}

func TestAttackStrategyConverter(t *testing.T) {
	converter := newAttackStrategyConverter()
	as := &FixedConcurrentUsers{ConcurrentUsers: 200, RampUpPeriod: 4}
	dto, err := converter.convertAttackStrategy(as)
	assert.Nil(t, err)
	as2, err := converter.convertDTO(dto)
	assert.Nil(t, err)
	assert.EqualValues(t, as, as2)
}

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

func newBenchmarkAttacker(n string, wait time.Duration) Attacker {
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
	task := NewTask()
	task.Add(&fakeAttacker{}, 10)

	output := commander.Open(context.Background(), task)
	go func() {
		for range output {
		}
	}()

	commander.Command(&FixedConcurrentUsers{ConcurrentUsers: 50, RampUpPeriod: 3}, NonstopTimer{})
	<-time.After(2 * time.Second)
	commander.Command(&FixedConcurrentUsers{ConcurrentUsers: 80, RampUpPeriod: 5}, NonstopTimer{})
	<-time.After(2 * time.Second)
	commander.Command(&FixedConcurrentUsers{ConcurrentUsers: 30, RampUpPeriod: 7}, NonstopTimer{})
	commander.Close()
}

func TestFCUSBenchmark(t *testing.T) {
	commander := newFixedConcurrentUsersStrategyCommander()
	task := NewTask()
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

	commander.Command(&FixedConcurrentUsers{ConcurrentUsers: 100, RampUpPeriod: 0}, NonstopTimer{})
	<-time.After(5 * time.Second)
	commander.Close()
	wg.Wait()

	report := sg.Report(true) // tps理论最大值10000, 1.6.0该配置均值在8000左右
	Logger.Info("report", zap.Float64("tps", report.TotalTPS), zap.Time("first_attack", report.FirstAttack), zap.Time("last_attack", report.LastAttack))
}
