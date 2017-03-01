package ultron

import (
	"sort"
	"strconv"
	"sync"
	"time"
)

var timeDistributions = [10]float64{0.5, 0.6, 0.7, 0.8, 0.9, 0.95, 0.97, 0.98, 0.99, 1.0}

type (
	roundedMillisecond int64

	statsEntry struct {
		name              string                       // 统计对象名称
		numRequests       int64                        // 成功请求次数
		numFailures       int64                        // 失败次数
		totalResponseTime time.Duration                // 成功请求的response time总和，基于原始的响应时间
		minResponseTime   time.Duration                // 最快的响应时间
		maxResponseTime   time.Duration                // 最慢的响应时间
		trend             map[int64]int64              // 按时间轴（秒级）记录成功请求次数
		failuresTrend     map[int64]int64              // 按时间轴（妙计）记录错误的请求次数
		responseTimes     map[roundedMillisecond]int64 // 按优化后的响应时间记录成功请求次数
		failuresTimes     map[string]int64             // 记录不同错误的次数
		startTime         time.Time                    // 第一次收到请求的时间
		lastRequestTime   time.Time                    // 最后一次收到请求的时间
		interval          time.Duration                // 统计时间间隔，影响 CurrentQPS()
		lock              sync.RWMutex
	}

	// QueryResult 查询事件结果
	QueryResult struct {
		Name     string
		Duration time.Duration
		Error    error
	}

	// StatsReport 输出报告
	StatsReport struct {
		Name           string           `json:"name"`
		Requests       int64            `json:"requests"`
		Failures       int64            `json:"failures"`
		Min            int64            `json:"min"`
		Max            int64            `json:"max"`
		Median         int64            `json:"median"`
		Average        int64            `json:"average"`
		QPS            int64            `json:"qps"`
		Distributions  map[string]int64 `json:"distributions"`
		FailureDetails map[string]int64 `json:"failure_details"`
		FullHistory    bool             `json:"full_history"`
	}
)

func newStatsEntry(n string) *statsEntry {
	return &statsEntry{
		name:          n,
		trend:         map[int64]int64{},
		failuresTrend: map[int64]int64{},
		responseTimes: map[roundedMillisecond]int64{},
		failuresTimes: map[string]int64{},
		interval:      time.Second * 12,
	}
}

func (s *statsEntry) logSuccess(t time.Duration) {
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

func (s *statsEntry) logFailure(e error) {
	s.lock.Lock()
	defer s.lock.Unlock()

	sec := time.Now().Unix()
	s.numFailures++
	info := e.Error()

	if _, ok := s.failuresTimes[info]; ok {
		s.failuresTimes[info]++
	} else {
		s.failuresTimes[info] = 1
	}

	if _, ok := s.failuresTrend[sec]; ok {
		s.failuresTrend[sec]++
	} else {
		s.failuresTrend[sec] = 1
	}
}

// TotalQPS 获取总的QPS
func (s *statsEntry) TotalQPS() float64 {
	return float64(s.numRequests) / s.lastRequestTime.Sub(s.startTime).Seconds()
}

// CurrentQPS 最近12秒的QPS
func (s *statsEntry) CurrentQPS() float64 {
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
func (s *statsEntry) Percentile(f float64) time.Duration {
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

	times := []roundedMillisecond{}
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
	return ZeroDuration // occur error
}

// Min 最快响应时间
func (s *statsEntry) Min() time.Duration {
	return s.minResponseTime
}

// Max 最慢响应时间
func (s *statsEntry) Max() time.Duration {
	return s.maxResponseTime
}

// Average 平均响应时间
func (s *statsEntry) Average() time.Duration {
	if s.numRequests == 0 {
		return ZeroDuration
	}
	return time.Duration(int64(s.totalResponseTime) / int64(s.numRequests))
}

// Median 响应时间中位数
func (s *statsEntry) Median() time.Duration {
	return s.Percentile(.5)
}

// FailRation 错误率
func (s *statsEntry) FailRation() float64 {
	return float64(s.numFailures) / float64(s.numRequests+s.numFailures)
}

// Report 打印统计结果
func (s *statsEntry) Report(full bool) *StatsReport {
	r := &StatsReport{
		Name:           s.name,
		Requests:       s.numRequests,
		Failures:       s.numFailures,
		Min:            timeDurationToMillsecond(s.Min()),
		Max:            timeDurationToMillsecond(s.Max()),
		Median:         timeDurationToMillsecond(s.Median()),
		Average:        timeDurationToMillsecond(s.Average()),
		Distributions:  map[string]int64{},
		FailureDetails: s.failuresTimes,
		FullHistory:    full,
	}

	if full {
		r.QPS = int64(s.TotalQPS())
	} else {
		r.QPS = int64(s.CurrentQPS())
	}

	for _, percent := range timeDistributions {
		r.Distributions[strconv.FormatFloat(percent, 'f', 2, 64)] = timeDurationToMillsecond(s.Percentile(percent))
	}
	return r
}
