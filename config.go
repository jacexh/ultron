package ultron

import (
	"errors"
	"math/rand"
	"sync"
	"time"

	"go.uber.org/zap"
)

type (
	// RunnerConfig runner配置参数
	RunnerConfig struct {
		Duration     time.Duration `json:"duration,omitempty"`    //v2废弃，但兼容V1
		Requests     uint64        `json:"requests,omitempty"`    //总请求数
		Concurrence  int           `json:"concurrence,omitempty"` //v2废弃，但兼容V1
		HatchRate    int           `json:"hatch_rate,omitempty"`  //v2废弃，但兼容V1
		MinWait      time.Duration `json:"min_wait,omitempty"`
		MaxWait      time.Duration `json:"max_wait,omitempty"`
		Stages       []*Stage      `json:"stages,omitempty"`
		currentStage int
		initialized  sync.Once
		mu           sync.RWMutex
	}

	// Stage 压测阶段配置参数
	Stage struct {
		Duration            time.Duration `json:"duration,omitempty"` // 阶段持续时间，不严格控制
		Requests            uint64        `json:"requests,omitempty"` // 阶段请求总数，不严格控制
		Concurrence         int           `json:"concurrence"`        // 阶段目标并发数
		previousConcurrence int           `json:"-"`
		HatchRate           int           `json:"hatch_rate,omitempty"` // 阶段增压/降压频率，为0不表示不控制，对于降压阶段，无需使用负数来表示降压频率
	}
)

const (
	// DefaultHatchRate 默认的增压/降压幅度
	DefaultHatchRate = 10
	// DefaultDuration 默认的压测持续时间，ZeroDuration表示不控制
	DefaultDuration = ZeroDuration
	// DefaultConcurrence 默认并发数
	DefaultConcurrence = 100
	// DefaultRequests 默认请求总次数，0表示不限制
	DefaultRequests = 0
	// DefaultMinWait 默认最小等待时间
	DefaultMinWait = 3 * time.Second
	// DefaultMaxWait 默认最大等待时间
	DefaultMaxWait = 5 * time.Second
)

var (
	// DefaultRunnerConfig 默认执行器配置
	DefaultRunnerConfig = &RunnerConfig{
		Duration:    DefaultDuration,    // 默认不控制压测时长
		Requests:    DefaultRequests,    // 请求总数，默认不控制，**而且无法严格控制**
		Concurrence: DefaultConcurrence, // 并发数，默认0并发    **19/3/2修改 原：100，改为0。为了判断是否是有效的配置。**
		HatchRate:   DefaultHatchRate,   // 加压频率，表示每秒启动多少goroutine，直到达到Concurrence的值；0 表示不控制，所有的并发goroutine在瞬间启动
		MinWait:     DefaultMinWait,     // 在单独的goroutine中，两次请求之间最少等待的时间
		MaxWait:     DefaultMaxWait,     // 在单独的goroutine中，两次请求之间最长等待的时间
	}
)

// NewRunnerConfig 创建新的执行器配置
func NewRunnerConfig() *RunnerConfig {
	return &RunnerConfig{
		Duration:    DefaultDuration,    // 默认不控制压测时长
		Requests:    DefaultRequests,    // 请求总数，默认不控制，**而且无法严格控制**
		Concurrence: DefaultConcurrence, // 并发数，默认100并发
		HatchRate:   DefaultHatchRate,   // 加压频率，表示每秒启动多少goroutine，直到达到Concurrence的值；0 表示不控制，所有的并发goroutine在瞬间启动
		MinWait:     DefaultMinWait,     // 在单独的goroutine中，两次请求之间最少等待的时间
		MaxWait:     DefaultMaxWait,     // 在单独的goroutine中，两次请求之间最长等待的时间
	}
}

// block 根据配置中的MinWait、MaxWait阻塞一段时间 [MinWait, MaxWait]
func (rc *RunnerConfig) block() {
	if rc.MinWait == ZeroDuration && rc.MaxWait == ZeroDuration {
		return
	}

	time.Sleep(rc.MinWait + time.Duration(rand.Int63n(int64(rc.MaxWait-rc.MinWait))+1))
}

// initialization 负责将默认的RunnerConfig转换成StageConfig
func (rc *RunnerConfig) initialization() {
	if rc.Stages == nil && rc.Concurrence <= 0 {
		return
	}

	// 不存在Stage时，初始化
	if rc.Stages == nil || len(rc.Stages) == 0 {
		rc.Stages = []*Stage{
			{Duration: rc.Duration, Requests: rc.Requests, Concurrence: rc.Concurrence, HatchRate: rc.HatchRate},
		}
		return
	}

	// 存在Stage时，确认previousConcurrence已经被设置
	var previous int
	for _, stage := range rc.Stages {
		stage.previousConcurrence = previous
		previous = stage.Concurrence
	}
}

// check 检查当前RunnerConfig配置是否合理
func (rc *RunnerConfig) check() error {
	// 兼容原配置文件，自动填充StageConfig
	rc.initialized.Do(rc.initialization)

	if (rc.Stages == nil || len(rc.Stages) == 0) && rc.Concurrence <= 0 {
		Logger.Error("invalid runnerConfig, something wrong ", zap.Any("runnerConfig", rc))
		return errors.New("invalid runnerConfig")
	}

	for num, sc := range rc.Stages {
		if sc.Concurrence <= 0 {
			Logger.Error("invalid Stage Concurrence value, it should be greater than 0 or InitConcurrence", zap.Int("Concurrence", sc.Concurrence))
			return errors.New("invalid Concurrency value")
		}
		if sc.HatchRate < 0 {
			Logger.Error("invalid Stage HatchRate value, it should be litter than 0 ", zap.Int("HatchRate", sc.HatchRate))
			return errors.New("invalid HatchRate value")
		}
		// 只有最后一阶段可以不控制压测时长
		if num < len(rc.Stages)-1 {
			if sc.Duration <= ZeroDuration {
				Logger.Error("invalid Stage Duration value, it should be greater than 0", zap.Duration("Duration", sc.Duration))
				return errors.New("invalid Concurrency value")
			}
		} else {
			if sc.Duration < ZeroDuration {
				Logger.Error("invalid Stage Duration value, last stage's Duration should be greater than 0 or equal to 0", zap.Duration("Duration", sc.Duration))
				return errors.New("invalid Concurrency value")
			}
		}
	}

	return nil
}

// AppendStage 添加下一个阶段配置
func (rc *RunnerConfig) AppendStage(sc *Stage) *RunnerConfig {
	if rc.Stages == nil {
		rc.Stages = []*Stage{}
	}
	sc.previousConcurrence = rc.Stages[len(rc.Stages)-1].Concurrence
	rc.Stages = append(rc.Stages, sc)
	return rc
}

// AppendStages 批量添加StageConfig
func (rc *RunnerConfig) AppendStages(sc ...*Stage) *RunnerConfig {
	for _, s := range sc {
		rc.AppendStage(s)
	}
	return rc
}

// FinishStage 通知完成当前Stage，如果已经是最后一个stage，返回error
func (rc *RunnerConfig) FinishStage(s int) (int, *Stage, error) {
	rc.mu.Lock()
	defer rc.mu.Unlock()

	if rc.currentStage == len(rc.Stages)-1 {
		return 0, nil, errors.New("reached the end of stages")
	}

	switch {
	case s == rc.currentStage:
		rc.currentStage++
		return rc.currentStage, rc.Stages[rc.currentStage], nil

	default: //   s 小于或者大于 rc.currentStage  无视
		return rc.currentStage, rc.Stages[rc.currentStage], nil
	}

}

// CurrentStage 获取当前Stage
func (rc *RunnerConfig) CurrentStage() (int, *Stage) {
	rc.mu.RLock()
	defer rc.mu.RUnlock()
	return rc.currentStage, rc.Stages[rc.currentStage]
}

// hatchWorkerCounts 计算出每秒启动/关闭的goroutine数量
func (sc *Stage) hatchWorkerCounts() []int {
	var ret []int
	increment := sc.Concurrence - sc.previousConcurrence
	if increment == 0 { // 无增减
		return ret
	}

	if sc.HatchRate == 0 || abs(increment) <= sc.HatchRate {
		ret = append(ret, increment)
		return ret
	}

	rounds := abs(increment) / sc.HatchRate
	perRound := sc.HatchRate
	if increment < 0 {
		perRound = -sc.HatchRate
	}
	for i := 0; i < rounds; i++ {
		ret = append(ret, perRound)
	}

	if increment%sc.HatchRate != 0 {
		ret = append(ret, increment-(rounds*perRound))
	}
	return ret
}

func (rc *RunnerConfig) split(n int) []*RunnerConfig {
	rc.initialized.Do(rc.initialization)

	var ret []*RunnerConfig
	var stageconfig [][]*Stage

	//err := rc.check()
	//if err != nil {
	//	Logger.Error("bad RunnerConfig", zap.Error(err))
	//}
	req := split(rc.Requests, uint64(n))

	for _, stage := range rc.Stages {
		stp := stage.split(n)
		for i := range stp {
			stageconfig = append(stageconfig, []*Stage{})
			stageconfig[i] = append(stageconfig[i], stp[i])
		}
	}

	for i := 0; i < n; i++ {
		r := &RunnerConfig{
			Duration:    rc.Duration,
			Concurrence: rc.Concurrence,
			HatchRate:   rc.HatchRate,
			MinWait:     rc.MinWait,
			MaxWait:     rc.MaxWait,
		}
		r.Requests = req[i]
		r.Stages = stageconfig[i]
		ret = append(ret, r)
	}
	return ret
}

// NewStage 实例化Stage
func NewStage(c int) *Stage {
	return &Stage{
		Duration:    DefaultDuration,
		Requests:    DefaultRequests,
		Concurrence: c,
		HatchRate:   DefaultHatchRate,
	}
}

func (sc *Stage) split(n int) []*Stage {
	//n <= 0时，会返回空[]*Stage
	var ret []*Stage
	c := split(uint64(sc.Concurrence), uint64(n))
	h := split(uint64(sc.HatchRate), uint64(n))

	for i := 0; i < n; i++ {
		s := &Stage{
			Duration: sc.Duration,
		}
		if sc.Concurrence == 0 {
			s.Concurrence = 0
		} else {
			s.Concurrence = int(c[i])
		}

		if sc.HatchRate == 0 {
			s.HatchRate = 0
		} else {
			s.HatchRate = int(h[i])
		}
		ret = append(ret, s)
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
