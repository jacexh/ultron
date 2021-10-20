package plan

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/wosai/ultron/v2"
	"github.com/wosai/ultron/v2/pkg/statistics"
)

func TestPlan_AddStages(t *testing.T) {
	plan := NewPlan()
	_ = plan.AddStages(
		ultron.StageConfig{Duration: 1 * time.Hour, Concurrence: 100},
		ultron.StageConfig{Requests: 1024 * 1024, Concurrence: 200},
	)

	assert.Nil(t, plan.Check())
}

func TestPlan_startNextStage(t *testing.T) {
	plan := NewPlan()
	_ = plan.AddStages(
		ultron.StageConfig{Duration: 1 * time.Hour, Concurrence: 100},
		ultron.StageConfig{Requests: 1024 * 1024, Concurrence: 200},
	)
	assert.Nil(t, plan.Check())

	stopped, i, conf, err := plan.StopCurrentAndStartNext(-1, nil)
	assert.Nil(t, err)
	assert.EqualValues(t, i, 0)
	assert.EqualValues(t, conf, plan.stages[0])
	assert.True(t, stopped)

	// 尚未超时
	stopped, i, conf, err = plan.StopCurrentAndStartNext(i, &statistics.SummaryReport{
		LastAttack:    time.Now(),
		FirstAttack:   time.Now().Add(-30 * time.Minute),
		TotalRequests: 10000,
		Reports:       map[string]*statistics.AttackReport{},
	})
	assert.False(t, stopped)
	assert.Nil(t, err)

	// 已经超时
	stopped, i, conf, err = plan.StopCurrentAndStartNext(i, &statistics.SummaryReport{
		LastAttack:    time.Now(),
		FirstAttack:   time.Now().Add(-61 * time.Minute),
		TotalRequests: 10000,
		Reports:       map[string]*statistics.AttackReport{},
	})
	assert.Nil(t, err)
	assert.EqualValues(t, i, 1)
	assert.EqualValues(t, conf, plan.stages[1])
	assert.True(t, stopped)

	// 第二阶段累计请求数
	stopped, i, conf, err = plan.StopCurrentAndStartNext(1, &statistics.SummaryReport{
		LastAttack:    time.Now(),
		FirstAttack:   time.Now().Add(-10 * time.Minute),
		TotalRequests: 10000 + 1024*1024 - 1,
		Reports:       map[string]*statistics.AttackReport{},
	})
	assert.False(t, stopped)
	assert.Nil(t, err)

	stopped, i, conf, err = plan.StopCurrentAndStartNext(1, &statistics.SummaryReport{
		LastAttack:    time.Now(),
		FirstAttack:   time.Now().Add(-10 * time.Minute),
		TotalRequests: 10000 + 1024*1024,
		Reports:       map[string]*statistics.AttackReport{},
	})
	assert.True(t, stopped)
	assert.Error(t, ultron.ErrPlanClosed)
}

func TestPlan_Stages(t *testing.T) {
	plan := NewPlan()
	_ = plan.AddStages(
		ultron.StageConfig{Duration: 1 * time.Hour, Concurrence: 100},
		ultron.StageConfig{Requests: 1024 * 1024, Concurrence: 200},
	)

	stages := plan.Stages()
	assert.EqualValues(t, plan.stages, stages)
	stages[0].Duration = 2 * time.Hour
	assert.EqualValues(t, plan.stages[0].Duration, 1*time.Hour)
}
