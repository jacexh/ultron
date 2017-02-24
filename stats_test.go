package ultron

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestNewStatsEntry(t *testing.T) {
	e := NewStatsEntry("test")
	assert.Equal(t, int(e.MinResponseTime), 0, "should be 0s")
}

func TestRoudedMilliSecond(t *testing.T) {
	d := time.Millisecond * 54
	rd := timeDurationToRoudedMilliSecond(d)
	assert.True(t, (rd == 54), "not rounded")

	d = time.Millisecond * 124
	rd = timeDurationToRoudedMilliSecond(d)
	assert.True(t, (rd == 120), "equal 120")

	d = time.Millisecond * 125
	rd = timeDurationToRoudedMilliSecond(d)
	assert.True(t, (rd == 130), "equal 130")

	d = time.Millisecond * 1230
	rd = timeDurationToRoudedMilliSecond(d)
	assert.True(t, (rd == 1200), "equal 1200")

	d = time.Millisecond * 1250
	rd = timeDurationToRoudedMilliSecond(d)
	assert.True(t, (rd == 1300), "equal 1300")
}

func TestLogSuccess(t *testing.T) {
	s := NewStatsEntry("test")
	for i := 0; i < 10000; i++ {
		s.logSuccess(time.Millisecond * 64)
	}
	time.Sleep(time.Second * 1)
	s.logSuccess(time.Millisecond * 84)
	assert.True(t, (s.TotalTPS() > 9900.0), "greater than 9900")
}
