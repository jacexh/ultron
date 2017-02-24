package ultron

import (
	"sync"
	"sync/atomic"
	"time"
)

type (
	// RoundedMillisecond 将time.Duration四舍五入到ms后的对象
	RoundedMillisecond int64

	// StatsEntry represents a single stats entry
	StatsEntry struct {
		Name              string
		NumRequests       uint64
		NumFailures       uint64
		TotalResponseTime time.Duration
		MinResponseTime   time.Duration
		MaxResponseTime   time.Duration
		TPSTrend          map[int64]uint64
		ResponseTimes     map[RoundedMillisecond]uint64
		StartTime         time.Time
		LastRequestTime   time.Time
		interval          time.Duration // interval 统计TPS的时间间隔
		lock              *sync.RWMutex
	}

	// StatsCollector 统计集合
	StatsCollector map[string]*StatsEntry
)

func timeDurationToRoudedMilliSecond(t time.Duration) RoundedMillisecond {
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

// NewStatsEntry create a new StatsEntry instance
func NewStatsEntry(n string) *StatsEntry {
	return &StatsEntry{
		Name:          n,
		TPSTrend:      map[int64]uint64{},
		ResponseTimes: map[RoundedMillisecond]uint64{},
		interval:      time.Second * 2,
		lock:          &sync.RWMutex{},
	}
}

func (s *StatsEntry) logSuccess(t time.Duration) {
	s.lock.Lock()
	defer s.lock.Unlock()

	atomic.AddUint64(&s.NumRequests, 1)

	now := time.Now()
	if s.LastRequestTime.IsZero() {
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
	if _, ok := s.TPSTrend[sec]; ok {
		s.TPSTrend[sec]++
	} else {
		s.TPSTrend[sec] = 1
	}

	rm := timeDurationToRoudedMilliSecond(t)
	if _, ok := s.ResponseTimes[rm]; ok {
		s.ResponseTimes[rm]++
	} else {
		s.ResponseTimes[rm] = 1
	}
}

func (s *StatsEntry) logFailure(t time.Duration, e error) {

}

// TotalTPS 获取总的TPS
func (s *StatsEntry) TotalTPS() float64 {
	s.lock.RLock()
	defer s.lock.RUnlock()

	return float64(s.NumRequests) / s.LastRequestTime.Sub(s.StartTime).Seconds()
}
