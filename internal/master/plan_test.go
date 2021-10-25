package master

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/wosai/ultron/v2"
	"github.com/wosai/ultron/v2/pkg/statistics"
)

func TestPlan_AddStages(t *testing.T) {
	plan := NewPlan()
	plan.AddStages(
		ultron.V1StageConfig{ConcurrentUsers: 100, RampUpPeriod: 3},
	)

	assert.Nil(t, plan.Start())
}

func TestPlan_startNextStage(t *testing.T) {
	p1 := NewPlan()
	p1.AddStages(
		ultron.BuildStage().WithAttackStrategy(&ultron.FixedConcurrentUsers{ConcurrentUsers: 100}).
			WithExitConditions(&ultron.UniversalExitConditions{Duration: 1 * time.Hour}),
		ultron.BuildStage().WithExitConditions(&ultron.UniversalExitConditions{Requests: 1024 * 1024}).
			WithAttackStrategy(&ultron.FixedConcurrentUsers{ConcurrentUsers: 200}),
	)
	assert.Nil(t, p1.Start())
	assert.EqualValues(t, p1.Status(), ultron.StatusReady)

	stopped, i, stage, err := p1.StopCurrentAndStartNext(-1, statistics.SummaryReport{})
	assert.Nil(t, err)
	assert.EqualValues(t, i, 0)
	assert.EqualValues(t, stage, p1.(*plan).stages[0])
	assert.True(t, stopped)

	assert.EqualValues(t, p1.Status(), ultron.StatusRunning)

	// 尚未超时
	stopped, i, stage, err = p1.StopCurrentAndStartNext(i, statistics.SummaryReport{
		LastAttack:    time.Now(),
		FirstAttack:   time.Now().Add(-30 * time.Minute),
		TotalRequests: 10000,
		Reports:       map[string]statistics.AttackReport{},
	})
	assert.False(t, stopped)
	assert.Nil(t, err)

	// 已经超时
	stopped, i, stage, err = p1.StopCurrentAndStartNext(i, statistics.SummaryReport{
		LastAttack:    time.Now(),
		FirstAttack:   time.Now().Add(-61 * time.Minute),
		TotalRequests: 10000,
		Reports:       map[string]statistics.AttackReport{},
	})
	assert.Nil(t, err)
	assert.EqualValues(t, i, 1)
	assert.EqualValues(t, stage, p1.(*plan).stages[1])
	assert.True(t, stopped)

	// 第二阶段累计请求数
	stopped, i, stage, err = p1.StopCurrentAndStartNext(1, statistics.SummaryReport{
		LastAttack:    time.Now(),
		FirstAttack:   time.Now().Add(-10 * time.Minute),
		TotalRequests: 10000 + 1024*1024 - 1,
		Reports:       map[string]statistics.AttackReport{},
	})
	assert.False(t, stopped)
	assert.Nil(t, err)

	stopped, i, stage, err = p1.StopCurrentAndStartNext(1, statistics.SummaryReport{
		LastAttack:    time.Now(),
		FirstAttack:   time.Now().Add(-10 * time.Minute),
		TotalRequests: 10000 + 1024*1024,
		Reports:       map[string]statistics.AttackReport{},
	})
	assert.True(t, stopped)
	assert.Error(t, ultron.ErrPlanClosed)
	assert.EqualValues(t, p1.Status(), ultron.StatusFinished)
}
