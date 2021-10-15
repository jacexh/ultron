package statistics

import (
	"math/rand"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestFindResponseBucket(t *testing.T) {
	t1 := 61 * time.Millisecond
	assert.EqualValues(t, t1, findReponseBucket(t1))

	t2 := 121 * time.Millisecond
	assert.EqualValues(t, 120*time.Millisecond, findReponseBucket(t2))

	t3 := 1111 * time.Millisecond
	assert.EqualValues(t, 1100*time.Millisecond, findReponseBucket(t3))

	t4 := 3 * time.Second * 333 * time.Microsecond
	assert.NotEqualValues(t, 3300*time.Microsecond, findReponseBucket(t4))
}

func BenchmarkAttackResultAggregator_RecordSuccess(b *testing.B) {
	agg := NewAttackResultAggregator("benchmark")
	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			agg.recordSuccess(&AttackResut{
				Name:     "benchmark",
				Duration: 111 * time.Millisecond,
			})
		}
	})
}

func BenchmarkAttackResultAggregator_Percent(b *testing.B) {
	agg := NewAttackResultAggregator("benchmark")
	for i := 0; i < 1000*1000; i++ {
		agg.Record(&AttackResut{
			Name:     "benchmark",
			Duration: time.Duration(rand.Int63n(2000)) * time.Millisecond,
		})
	}

	for i := 0; i < b.N; i++ {
		agg.percentile(.9)
	}
}

func BenchmarkAttackResultAggregator_Percent2(b *testing.B) {
	agg := NewAttackResultAggregator("benchmark")
	for i := 0; i < 1000*1000; i++ {
		agg.Record(&AttackResut{
			Name:     "benchmark",
			Duration: time.Duration(rand.Int63n(2000)) * time.Millisecond,
		})
	}

	for i := 0; i < b.N; i++ {
		agg.percentile(.5, .6, .7, .8, .9, .95, .99)
	}
}

func TestAttackResultAggregator_Percent(t *testing.T) {
	agg := NewAttackResultAggregator("benchmark")
	for i := 1; i <= 11; i++ {
		agg.Record(&AttackResut{
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
	a1 := NewAttackResultAggregator("test")
	for i := 0; i < 10; i++ {
		a1.Record(&AttackResut{Name: "test", Duration: time.Duration(i) * time.Millisecond})
	}
	a2 := NewAttackResultAggregator("test")
	for j := 10; j < 20; j++ {
		a2.Record(&AttackResut{Name: "test", Duration: time.Duration(j) * time.Millisecond})
	}

	a1.merge(a2)
	report := a1.Report(true)

	assert.EqualValues(t, report.Average, 9500*time.Microsecond)
	assert.EqualValues(t, report.Median, 9*time.Millisecond)
	assert.EqualValues(t, report.Requests, 20)
}
