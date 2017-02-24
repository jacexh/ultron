package ultron

import (
	"sort"
	"sync"
	"time"

	"go.uber.org/zap"
)

type (
	// RoundedMillisecond 将time.Duration四舍五入到ms后的对象
	RoundedMillisecond int64

	// StatsEntry represents a single stats entry
	StatsEntry struct {
		Name              string                     // 事务名称
		NumRequests       int                        // 成功请求次数
		NumFailures       int                        // 失败次数
		TotalResponseTime time.Duration              // 成功请求的response time总和，基于原始的响应时间
		MinResponseTime   time.Duration              // 最快的响应时间
		MaxResponseTime   time.Duration              // 最慢的响应时间
		Trend             map[int64]int              // 按时间轴（秒级）记录成功请求次数
		ResponseTimes     map[RoundedMillisecond]int // 按优化后的响应时间记录成功请求次数
		StartTime         time.Time                  // 第一次收到请求的时间
		LastRequestTime   time.Time                  // 最后一次收到请求的时间
		interval          time.Duration              // 统计时间间隔，影响 CurrentQPS()
		lock              *sync.RWMutex
	}

	// StatsCollector 统计集合
	StatsCollector map[string]*StatsEntry
)

func timeDurationToRoudedMillisecond(t time.Duration) RoundedMillisecond {
	ms := int64(t.Seconds()*1000 + 0.5)
	var rm RoundedMillisecond
	if ms < 100 {
		rm = RoundedMillisecond(ms)
	} else if ms < 1000 {
		rm = RoundedMillisecond(((ms + 5) / 10) * 10)
	} else {
		rm = RoundedMillisecond(((ms + 50) / 100) * 100)
	}
	return rm
}

func roundedMillisecondToDuration(r RoundedMillisecond) time.Duration {
	return time.Duration(r * 1000 * 1000)
}

// NewStatsEntry create a new StatsEntry instance
func NewStatsEntry(n string) *StatsEntry {
	return &StatsEntry{
		Name:          n,
		Trend:         map[int64]int{},
		ResponseTimes: map[RoundedMillisecond]int{},
		interval:      time.Second * 2,
		lock:          &sync.RWMutex{},
	}
}

func (s *StatsEntry) logSuccess(t time.Duration) {
	s.lock.Lock()
	defer s.lock.Unlock()

	s.NumRequests++

	now := time.Now()
	if s.LastRequestTime.IsZero() {
		s.MinResponseTime = t
		s.StartTime = now
	}
	s.LastRequestTime = now

	if t < s.MinResponseTime {
		s.MinResponseTime = t
	}
	if t > s.MaxResponseTime {
		s.MaxResponseTime = t
	}
	s.TotalResponseTime += t

	sec := now.Unix()
	if _, ok := s.Trend[sec]; ok {
		s.Trend[sec]++
	} else {
		s.Trend[sec] = 1
	}

	rm := timeDurationToRoudedMillisecond(t)
	if _, ok := s.ResponseTimes[rm]; ok {
		s.ResponseTimes[rm]++
	} else {
		s.ResponseTimes[rm] = 1
	}
}

func (s *StatsEntry) logFailure(t time.Duration, e error) {

}

// TotalQPS 获取总的QPS
func (s *StatsEntry) TotalQPS() float64 {
	s.lock.RLock()
	defer s.lock.RUnlock()

	return float64(s.NumRequests) / s.LastRequestTime.Sub(s.StartTime).Seconds()
}

// Percentile 获取x%的响应时间
func (s *StatsEntry) Percentile(f float64) time.Duration {
	s.lock.RLock()
	defer s.lock.RUnlock()

	if f <= 0.0 {
		return s.MinResponseTime
	}

	if f >= 1.0 {
		return s.MaxResponseTime
	}

	hint := int(float64(s.NumRequests)*f + .5)
	if hint == s.NumRequests {
		return s.MaxResponseTime
	}

	var times []RoundedMillisecond
	for k := range s.ResponseTimes {
		times = append(times, k)
	}

	sort.Slice(times, func(i, j int) bool { return times[i] < times[j] })
	for _, val := range times {
		counts := s.ResponseTimes[val]
		hint -= counts
		if hint <= 0 {
			return roundedMillisecondToDuration(val)
		}
	}
	Logger.Warn("occer error", zap.Int("hint", hint))
	return time.Nanosecond // occur error
}
