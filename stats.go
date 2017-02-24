package ultron

import (
	"sort"
	"sync"
	"sync/atomic"
	"time"

	"go.uber.org/zap"
)

type (
	// RoundedMillisecond 将time.Duration四舍五入到ms后的对象
	RoundedMillisecond int64

	// StatsEntry represents a single stats entry
	StatsEntry struct {
		name              string                       // 统计对象名称
		numRequests       int64                        // 成功请求次数
		numFailures       int64                        // 失败次数
		totalResponseTime time.Duration                // 成功请求的response time总和，基于原始的响应时间
		minResponseTime   time.Duration                // 最快的响应时间
		maxResponseTime   time.Duration                // 最慢的响应时间
		trend             map[int64]int64              // 按时间轴（秒级）记录成功请求次数
		responseTimes     map[RoundedMillisecond]int64 // 按优化后的响应时间记录成功请求次数
		startTime         time.Time                    // 第一次收到请求的时间
		lastRequestTime   time.Time                    // 最后一次收到请求的时间
		interval          time.Duration                // 统计时间间隔，影响 CurrentQPS()
		lock              *sync.RWMutex
	}

	// StatsCollector 统计集合
	StatsCollector map[string]*StatsEntry
)

// NewStatsEntry create a new StatsEntry instance
func NewStatsEntry(n string) *StatsEntry {
	return &StatsEntry{
		name:          n,
		trend:         map[int64]int64{},
		responseTimes: map[RoundedMillisecond]int64{},
		interval:      time.Second * 5,
		lock:          &sync.RWMutex{},
	}
}

func (s *StatsEntry) logSuccess(t time.Duration) {
	s.lock.Lock()
	defer s.lock.Unlock()

	s.numRequests++

	now := time.Now()
	if s.lastRequestTime.IsZero() {
		s.minResponseTime = t
		s.startTime = now
	}
	s.lastRequestTime = now

	if t < s.minResponseTime {
		s.minResponseTime = t
	}
	if t > s.maxResponseTime {
		s.maxResponseTime = t
	}
	s.totalResponseTime += t

	sec := now.Unix()
	if _, ok := s.trend[sec]; ok {
		s.trend[sec]++
	} else {
		s.trend[sec] = 1
	}

	rm := timeDurationToRoudedMillisecond(t)
	if _, ok := s.responseTimes[rm]; ok {
		s.responseTimes[rm]++
	} else {
		s.responseTimes[rm] = 1
	}
}

func (s *StatsEntry) logFailure(t time.Duration, e error) {
	atomic.AddInt64(&s.numFailures, 1)
	// todo: handle error
}

// TotalQPS 获取总的QPS
func (s *StatsEntry) TotalQPS() float64 {
	s.lock.RLock()
	defer s.lock.RUnlock()

	return float64(s.numRequests) / s.lastRequestTime.Sub(s.startTime).Seconds()
}

// CurrentQPS 最近5秒的QPS
func (s *StatsEntry) CurrentQPS() float64 {
	s.lock.RLock()
	defer s.lock.RUnlock()

	if s.lastRequestTime.IsZero() {
		return 0
	}
	end := s.lastRequestTime.Unix()
	start := s.lastRequestTime.Add(-s.interval).Unix()
	var total int64

	for k, v := range s.trend {
		if k >= start && k <= end {
			total += v
		}
	}
	return float64(total) / float64(s.interval/time.Second)
}

// Percentile 获取x%的响应时间
func (s *StatsEntry) Percentile(f float64) time.Duration {
	s.lock.RLock()
	defer s.lock.RUnlock()

	if f <= 0.0 {
		return s.minResponseTime
	}

	if f >= 1.0 {
		return s.maxResponseTime
	}

	hint := int64(float64(s.numRequests)*f + .5)
	if hint == s.numRequests {
		return s.maxResponseTime
	}

	var times []RoundedMillisecond
	for k := range s.responseTimes {
		times = append(times, k)
	}

	sort.Slice(times, func(i, j int) bool { return times[i] < times[j] })
	for _, val := range times {
		counts := s.responseTimes[val]
		hint -= counts
		if hint <= 0 {
			return roundedMillisecondToDuration(val)
		}
	}
	Logger.Warn("occer error", zap.Int64("hint", hint))
	return time.Nanosecond // occur error
}

// Min 最快响应时间
func (s *StatsEntry) Min() time.Duration {
	return s.minResponseTime
}

// Max 最慢响应时间
func (s *StatsEntry) Max() time.Duration {
	return s.maxResponseTime
}

// Average 平均响应时间
func (s *StatsEntry) Average() time.Duration {
	return time.Duration(int64(s.totalResponseTime) / int64(s.numRequests))
}

// Median 响应时间中位数
func (s *StatsEntry) Median() time.Duration {
	return s.Percentile(.5)
}

// FailRation 错误率
func (s *StatsEntry) FailRation() float64 {
	s.lock.RLock()
	defer s.lock.RUnlock()

	return float64(s.numFailures) / float64(s.numRequests+s.numFailures)
}
