package ultron

import "time"

type (
	// Stage 每个阶段必须包含
	Stage interface {
		ExitCondition
		AttackStrategy
		Timer
	}

	// ExitCondition 阶段退出条件
	ExitCondition interface {
		Exit(ExitCondition) bool
		Endless() bool
	}

	// AttackStrategy 压测策略
	AttackStrategy interface {
		GenWaves() []AttackWave
		Switch(next AttackStrategy) []AttackWave
	}

	// AttackWave 基于压测策略描述产生的增压、降压变化
	AttackWave struct {
		N        int
		Interval time.Duration
	}

	FixedUsers struct {
		User int `json:"user"`
	}

	FixedUserStage struct {
		FixedUsers              `json:",inline"`
		UniformRandomTimer      `json:",inline"`
		UniversalExitConditions `json:",inline"`
	}

	UniversalExitConditions struct {
		Requests uint64        `json:"requests,omitempty"`
		Duration time.Duration `json:"duration,omitempty"`
	}
)

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
