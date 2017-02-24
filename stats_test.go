package ultron

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestNewStatsEntry(t *testing.T) {
	e := NewStatsEntry("test")
	assert.Equal(t, int(e.minResponseTime), 0, "should be 0s")
}

func TestRoudedMilliSecond(t *testing.T) {
	d := time.Millisecond * 54
	rd := timeDurationToRoudedMillisecond(d)
	assert.True(t, (rd == 54), "not rounded")

	d = time.Millisecond * 124
	rd = timeDurationToRoudedMillisecond(d)
	assert.True(t, (rd == 120), "equal 120")

	d = time.Millisecond * 125
	rd = timeDurationToRoudedMillisecond(d)
	assert.True(t, (rd == 130), "equal 130")

	d = time.Millisecond * 1230
	rd = timeDurationToRoudedMillisecond(d)
	assert.True(t, (rd == 1200), "equal 1200")

	d = time.Millisecond * 1250
	rd = timeDurationToRoudedMillisecond(d)
	assert.True(t, (rd == 1300), "equal 1300")
}

func TestTotalQPS(t *testing.T) {
	s := NewStatsEntry("test")
	for i := 0; i < 10000; i++ {
		s.logSuccess(time.Millisecond * 64)
	}
	time.Sleep(time.Second * 1)
	s.logSuccess(time.Millisecond * 84)
	assert.True(t, (s.TotalQPS() > 9900.0), "greater than 9900")
}

func TestPercentile(t *testing.T) {
	s := NewStatsEntry("test")

	s.logSuccess(time.Millisecond * 10)
	s.logSuccess(time.Millisecond * 20)
	s.logSuccess(time.Millisecond * 20)
	s.logSuccess(time.Millisecond * 30)
	s.logSuccess(time.Millisecond * 50)

	assert.Equal(t, time.Millisecond*10, s.Percentile(0))
	assert.Equal(t, time.Millisecond*20, s.Percentile(.4))
	assert.Equal(t, time.Millisecond*50, s.Percentile(1))
	assert.Equal(t, time.Millisecond*30, s.Percentile(.8))
	assert.Equal(t, time.Millisecond*50, s.Percentile(.99))
}

func TestCurrentQPS(t *testing.T) {
	s := NewStatsEntry("test")

	for i := 0; i < 10; i++ {
		s.logSuccess(time.Millisecond * 10)
	}
	time.Sleep(time.Second * 2)

	for i := 0; i < 15; i++ {
		s.logSuccess(time.Millisecond * 10)
	}

	time.Sleep(time.Second * 4)
	s.logSuccess(time.Millisecond * 20) // refresh the last response time

	assert.True(t, (s.CurrentQPS() > 3.0), "greater than 3")
}

func TestAverage(t *testing.T) {
	s := NewStatsEntry("test")
	s.logSuccess(time.Millisecond * 10)
	s.logSuccess(time.Millisecond * 20)
	s.logSuccess(time.Millisecond * 30)
	s.logSuccess(time.Millisecond * 80)

	assert.Equal(t, time.Millisecond*35, s.Average())
}
