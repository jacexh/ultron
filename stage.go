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

	stage struct {
		timer    Timer
		checker  ExitConditions
		strategy AttackStrategyDescriber
	}

	// exitConditions 通用的退出条件
	UniversalExitConditions struct {
		Requests uint64        `json:"requests,omitempty"` // 请求总数，不严格控制
		Duration time.Duration `json:"duration,omitempty"` // 持续时长，不严格控制
	}
)

func (sec UniversalExitConditions) Check(actual ExitConditions) bool {
	if sec.NeverStop() {
		return false
	}
	if a, ok := actual.(UniversalExitConditions); ok {
		if sec.Duration > 0 && sec.Duration <= a.Duration {
			return true
		}
		if sec.Requests > 0 && sec.Requests <= a.Requests {
			return true
		}
	}
	return false
}

func (sec UniversalExitConditions) NeverStop() bool {
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
		s.checker = UniversalExitConditions{}
	} else {
		s.checker = ec
	}
	return s
}

func (s *stage) WithAttackStrategy(d AttackStrategyDescriber) *stage {
	if d == nil {
		panic("bad attack strategy")
	}
	s.strategy = d
	return s
}
