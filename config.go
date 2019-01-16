package ultron

import (
	"Richard1ybb/ultron/utils"
	"errors"
	"go.uber.org/zap"
	"math/rand"
	"time"
)

type (
	// RunnerConfig runner配置参数
	RunnerConfig struct {
		Duration    time.Duration `json:"duration"`      //v2废弃，但兼容V1
		Requests    uint64        `json:"requests"`      //总请求数
		Concurrence int           `json:"concurrence"`   //v2废弃，但兼容V1
		HatchRate   int           `json:"hatch_rate"`    //v2废弃，但兼容V1
		MinWait     time.Duration `json:"min_wait"`
		MaxWait     time.Duration `json:"max_wait"`
		Stages       []*StageConfig
	}

	StageConfig struct {
		Duration           time.Duration `json:"duration"`
		//InitConcurrence    int           `json:"init_concurrence"`  //初始并发数
		Concurrence        int           `json:"concurrence"`
		HatchRate          int           `json:"hatch_rate"`
	}


	////阶段运行配置
	//StageRunnerConfig struct {
	//	Requests     		uint64        `json:"requests"`
	//	MinWait      		time.Duration `json:"min_wait"`
	//	MaxWait      		time.Duration `json:"max_wait"`
	//	StageConfigs 		[]*StageConfig
	//	StageConfigsChanged []*StageConfigsChanged
	//	mu                  sync.Mutex
	//}


	StageConfigsChanged StageConfig
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
		Concurrence: 0,             // 并发数，默认100并发
		HatchRate:   0,               // 加压频率，表示每秒启动多少goroutine，直到达到Concurrence的值；0 表示不控制，所有的并发goroutine在瞬间启动
		MinWait:     time.Second * 3, // 在单独的goroutine中，两次请求之间最少等待的时间
		MaxWait:     time.Second * 5, // 在单独的goroutine中，两次请求之间最长等待的时间
		Stages:      []*StageConfig{},
	}
}


func NewStageConfig(Duration time.Duration, Concurrence int, HatchRate int) *StageConfig {
	return &StageConfig{
		Duration:         Duration,
		//InitConcurrence:  InitConcurrence,
		Concurrence: 	  Concurrence,
		HatchRate:   	  HatchRate,
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

	// stage模式
	if len(rc.Stages) != 0 && rc.Concurrence == 0 {
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
	}

	//v1Runner
	if len(rc.Stages) == 0 && rc.Concurrence != 0 {
		if rc.Concurrence <= 0 {
			Logger.Error("invalid Concurrence value, it should be greater than 0", zap.Int("Concurrence", rc.Concurrence))
			return errors.New("invalid Concurrency value")
		}
		if rc.MaxWait < rc.MinWait || rc.MaxWait < ZeroDuration {
			Logger.Error("invalid MaxWait/MinWait value")
			return errors.New("invalid MaxWait/MinWait value")
		}
		rc.v1Runner2Stage()
	}
	if (len(rc.Stages) == 0 && rc.Concurrence == 0) || (len(rc.Stages) != 0 && rc.Concurrence != 0) {
		Logger.Error("invalid runnerConfig, something wrong ", zap.Any("runnerConfig", rc))
		return errors.New("invalid runnerConfig")
	}

	return nil
}

// stage兼容v1版本
func (rc *RunnerConfig) v1Runner2Stage() {
	rc.AppendStage(&StageConfig{
		rc.Duration,
		rc.Concurrence,
		rc.HatchRate,
	})
	rc.Concurrence = 0
}

//// 清除Concurrence，运行stage
//func (rc *RunnerConfig) clearConcurrence()  {
//}


//TODO
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


//计算出每秒启动协程的数量
func (sc *StageConfig) hatchWorkerCounts() []int {
	rounds := 1
	var ret []int
	//var localConcurrence = sc.Concurrence

	//if scc.Concurrence <= 0  {
	//	Logger.Error("invalid Concurrence value, it should be greater than 0", zap.Int("Concurrence", scc.Concurrence))
	//	//return errors.New("invalid Concurrency value")
	//}

	if sc.Concurrence > 0 {
		if sc.HatchRate > 0 && sc.HatchRate < utils.Abs(sc.Concurrence) {
			rounds = sc.Concurrence / sc.HatchRate
			for i := 0; i < utils.Abs(rounds); i++ {
				ret = append(ret, sc.HatchRate)
			}
			last := sc.Concurrence % sc.HatchRate
			if utils.Abs(last) > 0 {
				ret = append(ret, last)
			}
		} else {
			ret = append(ret, sc.Concurrence)
		}
	} else {
		if sc.HatchRate > 0 && sc.HatchRate < utils.Abs(sc.Concurrence) {
			rounds = sc.Concurrence / sc.HatchRate
			for i := 0; i < utils.Abs(rounds); i++ {
				ret = append(ret, - sc.HatchRate)
			}
			last := sc.Concurrence % sc.HatchRate
			if utils.Abs(last) > 0 {
				ret = append(ret, last)
			}
		} else {
			ret = append(ret, sc.Concurrence)
		}
	}

	return ret
}


// 每个stage，协程变更数量  Concurrence  in:[100, 50, 70, 30] out:[100, -50, 20, -40]
func (rc *RunnerConfig) UpdateStageConfig() {
	var stageConfigChangeds = []*StageConfig{}

	currentConcurrence := 0
	for _, sc := range rc.Stages {
		concurrenceChanged := sc.Concurrence - currentConcurrence
		currentConcurrence = sc.Concurrence

		scc := &StageConfig{
			Duration:   	  sc.Duration,
			Concurrence:	  concurrenceChanged,
			HatchRate:  	  sc.HatchRate,
		}

		stageConfigChangeds = append(stageConfigChangeds, scc)
	}
	rc.Stages = stageConfigChangeds
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
	var stageconfig [][]*StageConfig
	rc.check()
	req := split(rc.Requests, uint64(n))

	for _, stage := range rc.Stages {
		stp := stage.split(n)
		for i := range stp {
			stageconfig = append(stageconfig, []*StageConfig{})
			stageconfig[i] = append(stageconfig[i], stp[i])
		}
	}

	for i := 0; i < n; i++ {
		r := &RunnerConfig{
			Duration:   rc.Duration,
			Concurrence:rc.Concurrence,
			HatchRate:  rc.HatchRate,
			MinWait:    rc.MinWait,
			MaxWait:    rc.MaxWait,
		}
		r.Requests = req[i]
		r.Stages = stageconfig[i]
		ret = append(ret, r)
	}
	return ret
}

func (sc *StageConfig) split(n int) []*StageConfig {
	//n <= 0时，会返回空[]*StageConfig
	var ret []*StageConfig
	c := split(uint64(sc.Concurrence), uint64(n))
	h := split(uint64(sc.HatchRate), uint64(n))

	for i := 0; i < n; i++ {
		s := &StageConfig{
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


//func NewStageRunnerConfig() *StageRunnerConfig {
//	return &StageRunnerConfig{
//		Requests:    0,               // 请求总数，默认不控制，**而且无法严格控制**
//		MinWait:     time.Second * 3, // 在单独的goroutine中，两次请求之间最少等待的时间
//		MaxWait:     time.Second * 5, // 在单独的goroutine中，两次请求之间最长等待的时间
//		StageConfigs:[]*StageConfig{}, // 各阶段的并发量及加压频率
//	}
//}

func (rc *RunnerConfig) AppendStage(sc ...*StageConfig) (rrc *RunnerConfig) {
	rc.Stages = append(rc.Stages, sc...)
	return rc
}


//func (src *StageRunnerConfig) check() error {
//	if src.MaxWait < src.MinWait || src.MaxWait < ZeroDuration {
//		Logger.Error("invalid MaxWait/MinWait value")
//		return errors.New("invalid MaxWait/MinWait value")
//	}
//	if src.StageConfigs == nil {
//		Logger.Error("empty StageConfigs")
//		return errors.New("empty StageConfigs")
//	}
//
//	return nil
//}




//func (src *StageRunnerConfig) block() {
//	if src.MinWait == ZeroDuration && src.MaxWait == ZeroDuration {
//		return
//	}
//
//	time.Sleep(src.MinWait + time.Duration(rand.Int63n(int64(src.MaxWait-src.MinWait))+1))
//}



