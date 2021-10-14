package plan

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/wosai/ultron/types"
)

func TestPlan_AddStages(t *testing.T) {
	plan := NewPlan()
	plan.AddStages(
		types.StageConfig{Duration: 1 * time.Hour, Concurrence: 100},
		types.StageConfig{Requests: 1024 * 1024, Concurrence: 200},
	)

	assert.Nil(t, plan.check())
}

func TestPlan_startNextStage(t *testing.T) {
	plan := NewPlan()
	plan.AddStages(
		types.StageConfig{Duration: 1 * time.Hour, Concurrence: 100},
		types.StageConfig{Requests: 1024 * 1024, Concurrence: 200},
	)

	i, conf, err := plan.finishAndStartNextStage(-1)
	assert.Nil(t, err)
	assert.EqualValues(t, i, 0)
	assert.EqualValues(t, conf, plan.stages[0])

	i, conf, err = plan.finishAndStartNextStage(i)
	assert.Nil(t, err)
	assert.EqualValues(t, i, 1)
	assert.EqualValues(t, conf, plan.stages[1])

	i, conf, err = plan.finishAndStartNextStage(1)
	assert.NotNil(t, err)
}
