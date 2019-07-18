package ultron

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestNewRunnerConfig(t *testing.T) {
	conf := NewRunnerConfig()
	assert.EqualValues(t, conf, DefaultRunnerConfig)
	assert.EqualValues(t, conf.Duration, DefaultDuration)
	assert.EqualValues(t, conf.Concurrence, DefaultConcurrence)
	assert.EqualValues(t, conf.Requests, DefaultRequests)
	assert.EqualValues(t, conf.HatchRate, DefaultHatchRate)
	assert.EqualValues(t, conf.MinWait, DefaultMinWait)
	assert.EqualValues(t, conf.MaxWait, DefaultMaxWait)
}

func TestNewStage(t *testing.T) {
	stage := NewStage()
	assert.EqualValues(t, stage.HatchRate, DefaultHatchRate)
	assert.EqualValues(t, stage.Duration, DefaultDuration)
	assert.EqualValues(t, stage.Requests, DefaultRequests)
	assert.EqualValues(t, stage.Concurrence, DefaultConcurrence)
}

func TestStage_hatchWorkerCounts_noChange(t *testing.T) {
	stage := NewStage()
	stage.previousConcurrence = DefaultConcurrence
	ret := stage.hatchWorkerCounts()
	assert.Equal(t, len(ret), 0)
}

func TestStageHatchWorkerCounts_increase(t *testing.T) {
	stage := &Stage{
		Concurrence:         100,
		HatchRate:           10,
		previousConcurrence: 50,
	}
	ret := stage.hatchWorkerCounts()
	assert.Equal(t, ret, []int{10, 10, 10, 10, 10})
}

func TestStageHatchWorkerCounts_increase2(t *testing.T) {
	stage := &Stage{
		Concurrence:         100,
		HatchRate:           30,
		previousConcurrence: 50,
	}
	ret := stage.hatchWorkerCounts()
	assert.Equal(t, ret, []int{30, 20})
}

func TestStageHatchWorkerCounts_reduce(t *testing.T) {
	stage := &Stage{
		Concurrence:         100,
		HatchRate:           10,
		previousConcurrence: 150,
	}
	ret := stage.hatchWorkerCounts()
	assert.Equal(t, ret, []int{-10, -10, -10, -10, -10})
}

func TestStageHatchWorkerCounts_reduce2(t *testing.T) {
	stage := &Stage{
		Concurrence:         100,
		HatchRate:           30,
		previousConcurrence: 150,
	}
	ret := stage.hatchWorkerCounts()
	assert.Equal(t, ret, []int{-30, -20})
}

func TestRunnerConfig_initV1(t *testing.T) {
	conf := NewRunnerConfig()
	conf.initialization()
	assert.EqualValues(t, 1, len(conf.Stages))
	stage := conf.Stages[0]
	assert.EqualValues(t, stage.HatchRate, DefaultHatchRate)
	assert.EqualValues(t, stage.Duration, DefaultDuration)
	assert.EqualValues(t, stage.Requests, DefaultRequests)
	assert.EqualValues(t, stage.Concurrence, DefaultConcurrence)
}

func TestRunnerConfig_initV2(t *testing.T) {
	conf := NewRunnerConfig()
	conf.AppendStage(&Stage{}).AppendStage(&Stage{})
	conf.initialization()
	assert.EqualValues(t, 2, len(conf.Stages))
	assert.EqualValues(t, conf.Stages[0].Concurrence, 0)
}

func TestRunnerConfig_initBadValue(t *testing.T) {
	conf := NewRunnerConfig()
	conf.Concurrence = 0
	conf.initialization()
	assert.Nil(t, conf.Stages)
}

func TestRunnerConfig_initPreviousConcurrence(t *testing.T) {
	conf := NewRunnerConfig()
	conf.AppendStages(
		&Stage{Concurrence: 100},
		&Stage{Concurrence: 200},
		&Stage{Concurrence: 50},
	)
	conf.initialization()

	assert.EqualValues(t, conf.Stages[0].previousConcurrence, 0)
	assert.EqualValues(t, conf.Stages[1].previousConcurrence, 100)
	assert.EqualValues(t, conf.Stages[2].previousConcurrence, 200)
}
