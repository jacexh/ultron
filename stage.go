package ultron

import (
	"time"
)

type (
	// Stage 每个阶段必须包含
	Stage interface {
		ExitCondition
		Timer
	}

	// ExitCondition 阶段退出条件
	ExitCondition interface {
		Exit(ExitCondition) bool
		Endless() bool
	}

	UniversalExitConditions struct {
		Requests uint64        `json:"requests,omitempty"`
		Duration time.Duration `json:"duration,omitempty"`
	}

	stage struct {
		timer         Timer
		exitCondition ExitCondition
	}
)

// func BuildStage() Stage {
// 	return &stage{
// 		timer:         NonstopTimer{},
// 		exitCondition: UniversalExitConditions{},
// 		strategy:      FixedUserStrategy{},
// 	}
// }

// func (s *stage) WithExitCondition(requests uint64, duration time.Duration) Stage {
// 	return s
// }

func (s stage) Endless() bool {
	return s.Endless()
}

func (sec UniversalExitConditions) Exit(actual ExitCondition) bool {
	if sec.Endless() {
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

func (sec UniversalExitConditions) Endless() bool {
	if sec.Duration <= 0 && sec.Requests <= 0 {
		return true
	}
	return false
}
