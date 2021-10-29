package ultron

import (
	"time"
)

type (
	// ExitCondition 阶段退出条件
	ExitConditions interface {
		NeverStop() bool
		Check(ExitConditions) bool
	}

	// Stage 描述一个压测阶段，需要包含并发策略、延时器、退出条件
	Stage interface {
		GetTimer() Timer
		GetExitConditions() ExitConditions
		GetStrategy() AttackStrategy
	}

	// stage 通用的stage对象
	stage struct {
		timer    Timer
		checker  ExitConditions
		strategy AttackStrategy
	}

	// exitConditions 通用的退出条件
	UniversalExitConditions struct {
		Requests uint64        `json:"requests,omitempty"` // 请求总数，不严格控制
		Duration time.Duration `json:"duration,omitempty"` // 持续时长，不严格控制
	}

	V1StageConfig struct {
		Requests        uint64
		Duration        time.Duration
		ConcurrentUsers int
		RampUpPeriod    int // 单位秒
		MinWait         time.Duration
		MaxWait         time.Duration
	}
)

func (sec *UniversalExitConditions) Check(actual ExitConditions) bool {
	if sec.NeverStop() {
		return false
	}
	if a, ok := actual.(*UniversalExitConditions); ok {
		if sec.Duration > 0 && sec.Duration <= a.Duration {
			return true
		}
		if sec.Requests > 0 && sec.Requests <= a.Requests {
			return true
		}
	}
	return false
}

func (sec *UniversalExitConditions) NeverStop() bool {
	if sec.Duration <= 0 && sec.Requests <= 0 {
		return true
	}
	return false
}

func BuildStage() *stage {
	return &stage{}
}

func (s *stage) WithTimer(t Timer) *stage {
	if t == nil {
		s.timer = NonstopTimer{}
	} else {
		s.timer = t
	}
	return s
}

func (s *stage) WithExitConditions(ec ExitConditions) *stage {
	if ec == nil {
		s.checker = &UniversalExitConditions{}
	} else {
		s.checker = ec
	}
	return s
}

func (s *stage) WithAttackStrategy(d AttackStrategy) *stage {
	if d == nil {
		panic("bad attack strategy")
	}
	s.strategy = d
	return s
}

func (s *stage) GetTimer() Timer {
	return s.timer
}

func (s *stage) GetExitConditions() ExitConditions {
	return s.checker
}

func (s *stage) GetStrategy() AttackStrategy {
	return s.strategy
}

func (v1 *V1StageConfig) GetTimer() Timer {
	return &UniformRandomTimer{MinWait: v1.MinWait, MaxWait: v1.MaxWait}
}

func (v1 *V1StageConfig) GetExitConditions() ExitConditions {
	return &UniversalExitConditions{Requests: v1.Requests, Duration: v1.Duration}
}

func (v1 *V1StageConfig) GetStrategy() AttackStrategy {
	return &FixedConcurrentUsers{ConcurrentUsers: v1.ConcurrentUsers, RampUpPeriod: v1.RampUpPeriod}
}
