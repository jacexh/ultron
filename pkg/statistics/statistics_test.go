package statistics

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"math/rand"
	"sync"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestFindResponseBucket(t *testing.T) {
	t1 := 61 * time.Millisecond
	assert.EqualValues(t, t1, findResponseBucket(t1))

	t2 := 121 * time.Millisecond
	assert.EqualValues(t, 120*time.Millisecond, findResponseBucket(t2))

	t3 := 1111 * time.Millisecond
	assert.EqualValues(t, 1100*time.Millisecond, findResponseBucket(t3))

	t4 := 3 * time.Second * 333 * time.Microsecond
	assert.NotEqualValues(t, 3300*time.Microsecond, findResponseBucket(t4))
}

func TestAttackStatistician_Record(t *testing.T) {
	as := NewAttackStatistician("foobar")
	as.Record(AttackResult{
		Name:     "foobar",
		Duration: 200 * time.Millisecond,
	})
	as.Record(AttackResult{
		Name:     "foobar",
		Duration: 300 * time.Millisecond,
		Error:    errors.New("unknown"),
	})
	report := as.Report(false)
	assert.EqualValues(t, report.Requests, 1)
	assert.EqualValues(t, report.Failures, 1)
	assert.EqualValues(t, report.FailRation, .5)
}

func TestAttackStatistician_SyncRecord(t *testing.T) {
	as := NewAttackStatistician("foobar")
	ctx, cancel := context.WithTimeout(context.Background(), 18*time.Second)
	defer cancel()
	var wg sync.WaitGroup
	for i := 0; i < 20; i++ {
		wg.Add(1)
		func() {
			defer wg.Done()
			for {
				select {
				case <-ctx.Done():
					return
				default:
					as.Record(AttackResult{Name: "foobar", Duration: 150 * time.Second})
				}
			}

		}()
	}
	wg.Wait()
	report := as.Report(true)
	data, err := json.Marshal(report)
	assert.Nil(t, err)
	fmt.Println(string(data))
}

func BenchmarkAttackResultAggregator_RecordSuccess(b *testing.B) {
	agg := NewAttackStatistician("benchmark")
	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			agg.recordSuccess(AttackResult{
				Name:     "benchmark",
				Duration: 111 * time.Millisecond,
			})
		}
	})
}

func BenchmarkStatistician_SyncRecord(b *testing.B) {
	s := NewStatisticianGroup()
	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			s.Record(AttackResult{
				Name:     "benchmark",
				Duration: 111 * time.Millisecond,
			})
		}
	})
}

func BenchmarkAttackResultAggregator_Percent(b *testing.B) {
	agg := NewAttackStatistician("benchmark")
	for i := 0; i < 1000*1000; i++ {
		agg.Record(AttackResult{
			Name:     "benchmark",
			Duration: time.Duration(rand.Int63n(2000)) * time.Millisecond,
		})
	}

	for i := 0; i < b.N; i++ {
		agg.percentile(.9)
	}
}

func BenchmarkAttackResultAggregator_Percent2(b *testing.B) {
	agg := NewAttackStatistician("benchmark")
	for i := 0; i < 1000*1000; i++ {
		agg.Record(AttackResult{
			Name:     "benchmark",
			Duration: time.Duration(rand.Int63n(2000)) * time.Millisecond,
		})
	}

	for i := 0; i < b.N; i++ {
		agg.percentile(.5, .6, .7, .8, .9, .95, .99)
	}
}

func TestAttackResultAggregator_Percent(t *testing.T) {
	agg := NewAttackStatistician("benchmark")
	for i := 1; i <= 11; i++ {
		agg.Record(AttackResult{
			Name:     "benchmark",
			Duration: time.Duration(i) * time.Millisecond,
		})
	}

	assert.EqualValues(t, agg.percentile(.0)[0], 1*time.Millisecond)
	assert.EqualValues(t, agg.percentile(.1)[0], 1*time.Millisecond)
	assert.EqualValues(t, agg.percentile(.5)[0], 6*time.Millisecond)
	assert.EqualValues(t, agg.percentile(.9)[0], 10*time.Millisecond)
	assert.EqualValues(t, agg.percentile(.97)[0], 11*time.Millisecond)
}

func TestAttackResultAggregator_merge(t *testing.T) {
	a1 := NewAttackStatistician("test")
	for i := 0; i < 10; i++ {
		a1.Record(AttackResult{Name: "test", Duration: time.Duration(i) * time.Millisecond})
	}
	a2 := NewAttackStatistician("test")
	for j := 10; j < 20; j++ {
		a2.Record(AttackResult{Name: "test", Duration: time.Duration(j) * time.Millisecond})
	}

	a1.merge(a2)
	report := a1.Report(true)

	assert.False(t, report.FirstAttack.IsZero())
	assert.EqualValues(t, report.Average, 9500*time.Microsecond)
	assert.EqualValues(t, report.Median, 9*time.Millisecond)
	assert.EqualValues(t, report.Requests, 20)
}

func TestAttackResultAggregator_MergeEmptry(t *testing.T) {
	a1 := NewAttackStatistician("test")
	a2 := NewAttackStatistician("test")
	for i := 0; i < 20; i++ {
		l := rand.Float64()
		var err error
		if l <= 0.3 {
			err = errors.New("unknown error")
		}
		a2.Record(AttackResult{Name: "test", Error: err, Duration: 1 * time.Millisecond})
	}
	a1.merge(a2)

	assert.EqualValues(t, a1, a2)
	report := a1.Report(true)
	data, _ := json.Marshal(report)
	log.Println(string(data))
}

func TestStatisticianGroup_Attach(t *testing.T) {
	sg := NewStatisticianGroup()
	sg.Attach(Tag{Key: "plan", Value: "hello"})
	report := sg.Report(true)
	assert.EqualValues(t, report.Extras["plan"], "hello")
}
