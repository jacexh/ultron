package ultron

import (
	"errors"
	"math/rand"
	"time"

	"go.uber.org/zap"
)

type (
	// RunnerConfig runner配置参数
	RunnerConfig struct {
		Duration    time.Duration `json:"duration"`
		Requests    uint64        `json:"requests"`
		Concurrence int           `json:"concurrence"`
		HatchRate   int           `json:"hatch_rate"`
		MinWait     time.Duration `json:"min_wait"`
		MaxWait     time.Duration `json:"max_wait"`
	}
)

var (
	// DefaultRunnerConfig 默认执行器配置
	DefaultRunnerConfig = &RunnerConfig{
		Duration:    ZeroDuration,    // 默认不控制压测时长
		Requests:    0,               // 请求总数，默认不控制，**而且无法严格控制**
		Concurrence: 100,             // 并发数，默认100并发
		HatchRate:   0,               // 加压频率，表示每秒启动多少goroutine，直到达到Concurrence的值；0 表示不控制，所有的并发goroutine在瞬间启动
		MinWait:     time.Second * 3, // 在单独的goroutine中，两次请求之间最少等待的时间
		MaxWait:     time.Second * 5, // 在单独的goroutine中，两次请求之间最长等待的时间
	}
)

// NewRunnerConfig 创建新的执行器配置
func NewRunnerConfig() *RunnerConfig {
	return &RunnerConfig{
		Duration:    ZeroDuration,    // 默认不控制压测时长
		Requests:    0,               // 请求总数，默认不控制，**而且无法严格控制**
		Concurrence: 100,             // 并发数，默认100并发
		HatchRate:   0,               // 加压频率，表示每秒启动多少goroutine，直到达到Concurrence的值；0 表示不控制，所有的并发goroutine在瞬间启动
		MinWait:     time.Second * 3, // 在单独的goroutine中，两次请求之间最少等待的时间
		MaxWait:     time.Second * 5, // 在单独的goroutine中，两次请求之间最长等待的时间
	}
}

// block 根据配置中的MinWait、MaxWait阻塞一段时间 [MinWait, MaxWait]
func (rc *RunnerConfig) block() {
	if rc.MinWait == ZeroDuration && rc.MaxWait == ZeroDuration {
		return
	}

	time.Sleep(rc.MinWait + time.Duration(rand.Int63n(int64(rc.MaxWait-rc.MinWait))+1))
}

// check 检查当前RunnerConfig配置是否合理
func (rc *RunnerConfig) check() error {
	if rc.Concurrence <= 0 {
		Logger.Error("invalid Concurrence value, it should be greater than 0", zap.Int("Concurrence", rc.Concurrence))
		return errors.New("invalid Concurrency value")
	}
	if rc.MaxWait < rc.MinWait || rc.MaxWait < ZeroDuration {
		Logger.Error("invalid MaxWait/MinWait value")
		return errors.New("invalid MaxWait/MinWait value")
	}
	return nil
}

// hatchWorkerCounts 根据HatchRate和Concurrence的值，计算出每秒启动的worker(goroutine)数量
func (rc *RunnerConfig) hatchWorkerCounts() []int {
	rounds := 1
	var ret []int

	if rc.HatchRate > 0 && rc.HatchRate < rc.Concurrence {
		rounds = rc.Concurrence / rc.HatchRate
		for i := 0; i < rounds; i++ {
			ret = append(ret, rc.HatchRate)
		}
		last := rc.Concurrence % rc.HatchRate
		if last > 0 {
			ret = append(ret, last)
		}
	} else {
		ret = append(ret, rc.Concurrence)
	}
	return ret
}

func split(total uint64, n uint64) []uint64 {
	if n <= 1 {
		return []uint64{total}
	}

	var ret []uint64
	size := total / n
	for k := uint64(0); k < n; k++ {
		ret = append(ret, size)
	}
	ret[n-1] += total % n
	return ret
}

func (rc *RunnerConfig) split(n int) []*RunnerConfig {
	var ret []*RunnerConfig
	c := split(uint64(rc.Concurrence), uint64(n))
	h := split(uint64(rc.HatchRate), uint64(n))

	for i := 0; i < n; i++ {
		r := &RunnerConfig{
			Duration: rc.Duration,
			MinWait:  rc.MinWait,
			MaxWait:  rc.MaxWait,
		}

		if rc.Concurrence == 0 {
			r.Concurrence = 0
		} else {
			r.Concurrence = int(c[i])
		}

		if rc.HatchRate == 0 {
			r.HatchRate = 0
		} else {
			r.HatchRate = int(h[i])
		}

		ret = append(ret, r)
	}
	return ret
}
