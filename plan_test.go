package ultron

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/wosai/ultron/v2/pkg/statistics"
)

func TestPlan_AddStages(t *testing.T) {
	plan := newPlan()
	plan.addStage(
		BuildStage().WithAttackStrategy(&FixedConcurrentUsers{ConcurrentUsers: 100, RampUpPeriod: 5}),
	)

	assert.Nil(t, plan.check())
}

func TestPlan_startNextStage(t *testing.T) {
	plan := newPlan()
	plan.addStage(
		BuildStage().WithAttackStrategy(&FixedConcurrentUsers{ConcurrentUsers: 100}).
			WithExitConditions(&UniversalExitConditions{Duration: 1 * time.Hour}),
	)
	plan.addStage(
		BuildStage().WithExitConditions(&UniversalExitConditions{Requests: 1024 * 1024}).
			WithAttackStrategy(&FixedConcurrentUsers{ConcurrentUsers: 200}),
	)

	assert.Nil(t, plan.check())

	stopped, i, sta, timer, err := plan.stopCurrentAndStartNext(-1, nil)
	assert.Nil(t, err)
	assert.EqualValues(t, i, 0)
	assert.EqualValues(t, sta, plan.stages[0].strategy)
	assert.EqualValues(t, timer, plan.stages[0].timer)
	assert.True(t, stopped)

	// 尚未超时
	stopped, i, sta, timer, err = plan.stopCurrentAndStartNext(i, &statistics.SummaryReport{
		LastAttack:    time.Now(),
		FirstAttack:   time.Now().Add(-30 * time.Minute),
		TotalRequests: 10000,
		Reports:       map[string]*statistics.AttackReport{},
	})
	assert.False(t, stopped)
	assert.Nil(t, err)

	// 已经超时
	stopped, i, sta, timer, err = plan.stopCurrentAndStartNext(i, &statistics.SummaryReport{
		LastAttack:    time.Now(),
		FirstAttack:   time.Now().Add(-61 * time.Minute),
		TotalRequests: 10000,
		Reports:       map[string]*statistics.AttackReport{},
	})
	assert.Nil(t, err)
	assert.EqualValues(t, i, 1)
	assert.EqualValues(t, sta, plan.stages[1].strategy)
	assert.EqualValues(t, timer, plan.stages[1].timer)
	assert.True(t, stopped)

	// 第二阶段累计请求数
	stopped, i, sta, timer, err = plan.stopCurrentAndStartNext(1, &statistics.SummaryReport{
		LastAttack:    time.Now(),
		FirstAttack:   time.Now().Add(-10 * time.Minute),
		TotalRequests: 10000 + 1024*1024 - 1,
		Reports:       map[string]*statistics.AttackReport{},
	})
	assert.False(t, stopped)
	assert.Nil(t, err)

	stopped, i, sta, timer, err = plan.stopCurrentAndStartNext(1, &statistics.SummaryReport{
		LastAttack:    time.Now(),
		FirstAttack:   time.Now().Add(-10 * time.Minute),
		TotalRequests: 10000 + 1024*1024,
		Reports:       map[string]*statistics.AttackReport{},
	})
	assert.True(t, stopped)
	assert.Error(t, ErrPlanClosed)
}
