package statistics

import (
	"errors"
	"sort"
	"strconv"
	"sync"
	"time"
)

var (
	timeDistributions = []float64{0.5, 0.6, 0.7, 0.8, 0.9, 0.95, 0.97, 0.98, 0.99, 1.0}
)

const (
	CurrentTPSTimeRange = 12 * time.Second
)

type (
	// AttackResult 事务执行结果
	AttackResult struct {
		Name     string
		Duration time.Duration
		Error    error
	}

	AttackStatistician struct {
		name                string                   // 事务名称
		requests            uint64                   // 成功请求数
		failures            uint64                   // 失败请求数
		totalResponseTime   time.Duration            // 原始响应时间汇总
		minResponseTime     time.Duration            // 最小响应时间
		maxResponseTime     time.Duration            // 最长响应时间
		recentSuccessBucket *timeRangeContainer      // 最近的成功请求数量
		recentFailureBucket *timeRangeContainer      // 最近的失败请求数量
		responseBucket      map[time.Duration]uint64 // 成功请求的响应时间桶
		failureBucket       map[string]uint64        // 失败请求的错误原因桶
		firstAttack         time.Time                // 请求开始时间
		lastAttack          time.Time                // 最后一次收到响应结果的时间
		interval            time.Duration            // 统计CurrentTPS（）的时间区间
		mu                  sync.Mutex
	}

	// AttackReport 聚合报告
	AttackReport struct {
		Name           string                   `json:"name"`                      // 事务名称
		Requests       uint64                   `json:"requests,omitempty"`        // 成功请求总数
		Failures       uint64                   `json:"failures,omitempty"`        // 失败请求总数
		Min            time.Duration            `json:"min,omitempty"`             // 最小延迟
		Max            time.Duration            `json:"max,omitempty"`             // 最大延迟
		Median         time.Duration            `json:"median,omitempty"`          // 中位数
		Average        time.Duration            `json:"average,omitempty"`         // 平均数
		TPS            float64                  `json:"tps,omitempty"`             // 每秒事务数
		Distributions  map[string]time.Duration `json:"distributions,omitempty"`   // 百分位分布
		FailRatio      float64                  `json:"fail_ratio,omitempty"`      // 错误率
		FailureDetails map[string]uint64        `json:"failure_details,omitempty"` // 错误详情分布
		FullHistory    bool                     `json:"full_history"`              // 是否是该阶段完整的报告
		FirstAttack    time.Time                `json:"first_attack,omitempty"`    // 第一请求发生时间
		LastAttack     time.Time                `json:"last_attack,omitempty"`     // 最后一次请求结束时间
	}

	SummaryReport struct {
		FirstAttack   time.Time               `json:"first_attack,omitempty"`
		LastAttack    time.Time               `json:"last_attack,omitempty"`
		TotalRequests uint64                  `json:"total_requests,omitempty"`
		TotalFailures uint64                  `json:"total_failures,omitempty"`
		TotalTPS      float64                 `json:"total_tps,omitempty"`
		FullHistory   bool                    `json:"full_history"`
		Reports       map[string]AttackReport `json:"reports,omitempty"`
		Extras        map[string]string       `json:"extras,omitempty"`
	}

	timeRangeContainer struct {
		container map[int64]int64
		timeRange int64
	}

	StatisticianGroup struct {
		tags      map[string]Tag
		container map[string]*AttackStatistician // 优于sync.Map
		mu        sync.Mutex                     // 写多读少场景，互斥锁更好
	}

	Tag struct {
		Key   string `json:"key"`
		Value string `json:"value"`
	}

	Tags map[string]Tag
)

func newTimeRangeContainer(n int64) *timeRangeContainer {
	return &timeRangeContainer{
		container: make(map[int64]int64),
		timeRange: n,
	}
}

func (ls *timeRangeContainer) accumulate(k, v int64) {
	_, ok := ls.container[k]
	ls.container[k] += v

	if !ok {
		for key := range ls.container {
			if (k - key) > ls.timeRange {
				delete(ls.container, key)
			}
		}
	}
}

func findResponseBucket(t time.Duration) time.Duration {
	if t <= 100*time.Millisecond {
		return (t + 500*time.Microsecond) / 1e6 * 1e6
	}
	if t <= 1000*time.Millisecond {
		return (t + 5*time.Millisecond) / 1e7 * 1e7
	}
	return (t + 50*time.Millisecond) / 1e8 * 1e8
}

// IsFailure 事务是否执行失败
func (ar *AttackResult) IsFailure() bool {
	return ar.Error != nil
}

func NewAttackStatistician(name string) *AttackStatistician {
	return &AttackStatistician{
		name:                name,
		recentSuccessBucket: newTimeRangeContainer(15),
		recentFailureBucket: newTimeRangeContainer(15),
		responseBucket:      make(map[time.Duration]uint64),
		failureBucket:       make(map[string]uint64),
		interval:            CurrentTPSTimeRange,
	}
}

func (ara *AttackStatistician) recordSuccess(ret AttackResult) {
	if ara.name != ret.Name {
		return
	}
	ara.mu.Lock()
	defer ara.mu.Unlock()

	ara.requests++
	ara.totalResponseTime += ret.Duration

	now := time.Now()
	if ara.firstAttack.IsZero() { // 第一次记录，且是成功请求
		ara.firstAttack = now
		ara.minResponseTime = ret.Duration
		ara.maxResponseTime = ret.Duration
	} else if ara.minResponseTime == 0 && ara.maxResponseTime == 0 { // 如果第一次先到的是错误请求，则min repsone 必然为0
		ara.minResponseTime = ret.Duration
		ara.maxResponseTime = ret.Duration
	} else {
		if ret.Duration < ara.minResponseTime {
			ara.minResponseTime = ret.Duration
		}
		if ret.Duration > ara.maxResponseTime {
			ara.maxResponseTime = ret.Duration
		}
	}
	ara.lastAttack = now

	ara.recentSuccessBucket.accumulate(now.Unix(), 1)
	ara.responseBucket[findResponseBucket(ret.Duration)]++
}

func (ara *AttackStatistician) recordFailure(ret AttackResult) {
	if ara.name != ret.Name {
		return
	}

	ara.mu.Lock()
	defer ara.mu.Unlock()

	ara.failures++

	now := time.Now()
	if ara.firstAttack.IsZero() {
		ara.firstAttack = now
	}
	ara.lastAttack = now

	ara.failureBucket[ret.Error.Error()]++
	ara.recentFailureBucket.accumulate(now.Unix(), 1)
}

func (ara *AttackStatistician) Record(ret AttackResult) {
	if ret.IsFailure() {
		ara.recordFailure(ret)
		return
	}
	ara.recordSuccess(ret)
}

// TotalTPS 全程TPS
func (ara *AttackStatistician) totalTPS() float64 {
	if ara.lastAttack == ara.firstAttack {
		return 0
	}
	return float64(ara.requests) / ara.lastAttack.Sub(ara.firstAttack).Seconds()
}

// CurrentTPS 最近12秒的TPS
func (ara *AttackStatistician) currentTPS() float64 {
	if ara.lastAttack == ara.firstAttack {
		return 0
	}

	end := time.Now().Add(-1 * time.Second) // 当前一秒未完成，往前推一秒作为统计终点
	if end.Before(ara.firstAttack) {        // 尚未执行满一秒，不统计
		return 0
	}
	start := end.Add(-1 * (ara.interval - time.Second))
	if start.Before(ara.firstAttack) {
		start = ara.firstAttack
	}
	if end.Sub(start) <= 0 {
		return 0
	}

	var count int64
	for i := start.Unix(); i <= end.Unix(); i++ {
		if v, ok := ara.recentSuccessBucket.container[i]; ok {
			count += v
		}
	}
	return float64(count) / float64(end.Unix()-start.Unix()+1)
}

func (ara *AttackStatistician) percentile(ps ...float64) []time.Duration {
	var bucketKeys []time.Duration
	for k := range ara.responseBucket {
		bucketKeys = append(bucketKeys, k)
	}
	sort.Slice(bucketKeys, func(i, j int) bool {
		return bucketKeys[i] < bucketKeys[j]
	})

	results := make([]time.Duration, len(ps))

percent:
	for n, per := range ps {
		index := int64(float64(ara.requests)*per + .5)
		if index >= int64(ara.requests) {
			results[n] = ara.maxResponseTime
			continue percent
		}
		if index <= 1 {
			results[n] = ara.minResponseTime
			continue percent
		}

		for _, key := range bucketKeys {
			index -= int64(ara.responseBucket[key])
			if index <= 0 {
				results[n] = key
				continue percent
			}
		}
		panic("unreachable code")
	}
	return results
}

func (ara *AttackStatistician) min() time.Duration {
	return ara.minResponseTime
}

func (ara *AttackStatistician) max() time.Duration {
	return ara.maxResponseTime
}

func (ara *AttackStatistician) average() time.Duration {
	if ara.requests == 0 {
		return 0
	}
	return ara.totalResponseTime / time.Duration(ara.requests)
}

func (ara *AttackStatistician) failRatio() float64 {
	total := float64(ara.requests) + float64(ara.failures)
	if total == 0 {
		return 0.0
	}
	return float64(ara.failures) / total
}

func (ara *AttackStatistician) Report(full bool) AttackReport {
	ara.mu.Lock()
	defer ara.mu.Unlock()

	report := AttackReport{
		Name:           ara.name,
		Requests:       ara.requests,
		Failures:       ara.failures,
		Min:            ara.min(),
		Max:            ara.max(),
		Average:        ara.average(),
		Distributions:  make(map[string]time.Duration),
		FailRatio:      ara.failRatio(),
		FailureDetails: make(map[string]uint64),
		FullHistory:    full,
		FirstAttack:    ara.firstAttack,
		LastAttack:     ara.lastAttack,
	}
	if full {
		report.TPS = ara.totalTPS()
	} else {
		report.TPS = ara.currentTPS()
	}
	pers := ara.percentile(timeDistributions...)
	for index, d := range timeDistributions {
		report.Distributions[strconv.FormatFloat(d, 'f', 2, 64)] = pers[index]
	}
	report.Median = pers[0]

	// failure details
	for key, value := range ara.failureBucket {
		report.FailureDetails[key] = value
	}
	return report
}

func (ara *AttackStatistician) merge(other *AttackStatistician) error {
	if other == nil {
		return nil
	}
	if ara.name != other.name {
		return errors.New("cannot merge two different types report")
	}

	ara.mu.Lock()
	defer ara.mu.Unlock()
	other.mu.Lock()
	defer other.mu.Unlock()

	ara.requests += other.requests
	ara.failures += other.failures
	ara.totalResponseTime += other.totalResponseTime
	if (other.minResponseTime < ara.minResponseTime && other.minResponseTime > 0) || ara.minResponseTime == 0 {
		ara.minResponseTime = other.minResponseTime
	}
	if other.maxResponseTime > ara.maxResponseTime {
		ara.maxResponseTime = other.maxResponseTime
	}
	for k, v := range other.recentSuccessBucket.container {
		ara.recentSuccessBucket.accumulate(k, v)
	}
	for k, v := range other.recentFailureBucket.container {
		ara.recentFailureBucket.accumulate(k, v)
	}
	for k, v := range other.responseBucket {
		ara.responseBucket[k] += v
	}
	for k, v := range other.failureBucket {
		ara.failureBucket[k] += v
	}
	if (!other.firstAttack.IsZero() && other.firstAttack.Before(ara.firstAttack)) || ara.firstAttack.IsZero() {
		ara.firstAttack = other.firstAttack
	}
	if ara.lastAttack.Before(other.lastAttack) {
		ara.lastAttack = other.lastAttack
	}
	return nil
}

// BatchMerge 合并多个AttackStatistician对象
func (ara *AttackStatistician) BatchMerge(others ...*AttackStatistician) error {
	for _, other := range others {
		if err := ara.merge(other); err != nil {
			return err
		}
	}
	return nil
}

func NewStatisticianGroup() *StatisticianGroup {
	return &StatisticianGroup{
		container: make(map[string]*AttackStatistician),
		tags:      make(map[string]Tag),
	}
}

// Report 输出统计报表
func (s *StatisticianGroup) Report(full bool) SummaryReport {
	sr := SummaryReport{
		FullHistory: full,
		Reports:     make(map[string]AttackReport),
		Extras:      make(map[string]string),
	}

	s.mu.Lock()
	defer s.mu.Unlock()

	for key, value := range s.container {
		sr.Reports[key] = value.Report(full)
		sr.TotalRequests += sr.Reports[key].Requests
		sr.TotalFailures += sr.Reports[key].Failures
		sr.TotalTPS += sr.Reports[key].TPS

		if sr.FirstAttack.IsZero() && !sr.Reports[key].FirstAttack.IsZero() {
			sr.FirstAttack = sr.Reports[key].FirstAttack
		}
		if sr.LastAttack.IsZero() && !sr.Reports[key].LastAttack.IsZero() {
			sr.LastAttack = sr.Reports[key].LastAttack
		}
		if s := sr.Reports[key].FirstAttack; !s.IsZero() && s.Before(sr.FirstAttack) {
			sr.FirstAttack = s
		}
		if !sr.LastAttack.IsZero() && sr.LastAttack.Before(sr.Reports[key].LastAttack) {
			sr.LastAttack = sr.Reports[key].LastAttack
		}
	}

	for key, tag := range s.tags {
		sr.Extras[key] = tag.Value
	}
	return sr
}

// Record 记录一次请求结果
func (s *StatisticianGroup) Record(result AttackResult) {
	s.mu.Lock()
	defer s.mu.Unlock()

	if agg, ok := s.container[result.Name]; !ok {
		agg = NewAttackStatistician(result.Name)
		agg.Record(result)
		s.container[result.Name] = agg
	} else {
		agg.Record(result)
	}
}

// Reset 重置统计组状态
func (s *StatisticianGroup) Reset() {
	s.mu.Lock()
	defer s.mu.Unlock()

	for key := range s.container {
		delete(s.container, key)
	}

	for key := range s.tags {
		delete(s.tags, key)
	}
}

// ReplaceStatistician 替换某个事务的Statistician
func (s *StatisticianGroup) ReplaceStatistician(agg *AttackStatistician) error {
	if agg == nil {
		return errors.New("cannot replace with nil pointer")
	}
	s.mu.Lock()
	defer s.mu.Unlock()

	s.container[agg.name] = agg
	return nil
}

// Attach 附加tag
func (s *StatisticianGroup) Attach(tag Tag) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.tags[tag.Key] = tag
}

// Tags 返回当前统计组的所有tag
func (s *StatisticianGroup) Tags() Tags {
	s.mu.Lock()
	defer s.mu.Unlock()
	return s.tags
}

func (s *StatisticianGroup) SetTag(key, value string) {
	s.mu.Lock()
	defer s.mu.Unlock()

	s.tags[key] = Tag{Key: key, Value: value}
}

func (s *StatisticianGroup) Merge(other *StatisticianGroup) {
	if other == nil {
		return
	}
	s.mu.Lock()
	defer s.mu.Unlock()
	other.mu.Lock()
	defer other.mu.Unlock()

	for key, value := range other.tags {
		s.tags[key] = value
	}

	for key, value := range other.container {
		if _, ok := s.container[key]; !ok {
			s.container[key] = NewAttackStatistician(key)
		}
		s.container[key].merge(value)
	}
}
