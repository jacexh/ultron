package ultron

import (
	"errors"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestNewLimitedSizeMap(t *testing.T) {
	m := newLimitedSizeMap(3)
	assert.EqualValues(t, 3, m.size)
}

func TestLimitedSizeMap_accumulateOK(t *testing.T) {
	m := newLimitedSizeMap(3)
	m.accumulate(100, 1)
	assert.EqualValues(t, 1, m.content[100])

	m.accumulate(100, 3)
	assert.EqualValues(t, 4, m.content[100])
}

func TestLimitedSizeMap_accumulateToRemoveEle(t *testing.T) {
	m := newLimitedSizeMap(3)
	m.accumulate(100, 1)
	assert.EqualValues(t, 1, m.content[100])
	assert.EqualValues(t, 1, len(m.content))

	m.accumulate(103, 1)
	assert.EqualValues(t, 2, len(m.content))

	m.accumulate(104, 1)
	assert.EqualValues(t, 2, len(m.content))
	_, ok := m.content[100]
	assert.False(t, ok)
}

func TestNewAttackerStatistics(t *testing.T) {
	s := newAttackerStatistics("foobar")
	assert.Equal(t, s.name, "foobar")
	assert.Equal(t, s.interval, 12*time.Second)
	assert.EqualValues(t, s.trendFailures.size, 20)
	assert.EqualValues(t, s.trendSuccess.size, 20)
	assert.True(t, s.lastRequestTime.IsZero())
	assert.True(t, s.startTime.IsZero())
}

func TestAttackerStatistics_logSuccess(t *testing.T) {
	name := "foobar"
	stats := newAttackerStatistics(name)

	ret := &Result{Name: name, Duration: int64(10 * time.Millisecond)}
	stats.logSuccess(ret)

	assert.False(t, stats.startTime.IsZero())
	assert.False(t, stats.lastRequestTime.IsZero())
	assert.EqualValues(t, stats.minResponseTime, 10*time.Millisecond)
	assert.EqualValues(t, stats.minResponseTime, stats.maxResponseTime)
	assert.EqualValues(t, stats.totalResponseTime, 10*time.Millisecond)
	assert.EqualValues(t, stats.numRequests, 1)

	ret = &Result{Name: name, Duration: int64(20 * time.Millisecond)}
	time.Sleep(time.Second)
	stats.log(ret)
	assert.True(t, stats.lastRequestTime.After(stats.startTime))
	assert.EqualValues(t, stats.maxResponseTime, 20*time.Millisecond)
	assert.EqualValues(t, stats.minResponseTime, 10*time.Millisecond)
	assert.EqualValues(t, stats.totalResponseTime, 30*time.Millisecond)
	assert.EqualValues(t, stats.numRequests, 2)
	assert.EqualValues(t, 2, len(stats.trendSuccess.content))

	ret = &Result{Name: name, Duration: int64(5 * time.Millisecond)}
	stats.log(ret)
	assert.True(t, stats.lastRequestTime.After(stats.startTime))
	assert.EqualValues(t, stats.minResponseTime, 5*time.Millisecond)
	assert.EqualValues(t, stats.totalResponseTime, 35*time.Millisecond)
	assert.EqualValues(t, stats.numRequests, 3)
}

func TestAttackerStatistics_logFailure(t *testing.T) {
	name := "foobar"
	stats := newAttackerStatistics(name)

	ret := &Result{Name: name, Duration: int64(10 * time.Millisecond), Error: newAttackerError(name, errors.New("error"))}
	stats.log(ret)
	assert.False(t, stats.startTime.IsZero())
	assert.Equal(t, stats.startTime, stats.lastRequestTime)
	assert.EqualValues(t, 1, stats.numFailures)
	assert.EqualValues(t, 0, stats.numRequests)
	assert.EqualValues(t, 1, stats.failuresTimes["error"])
	assert.EqualValues(t, 1, len(stats.trendFailures.content))
}
