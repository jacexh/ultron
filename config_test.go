package ultron

import (
	"testing"
	"time"

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

func TestStageHatchWorkerCounts_zeroHatchRate(t *testing.T) {
	stage := &Stage{
		Concurrence: 100,
		HatchRate:   0,
	}
	ret := stage.hatchWorkerCounts()
	assert.Equal(t, ret, []int{100})
}

func TestStageHatchWorkerCounts_hatchAll(t *testing.T) {
	stage := &Stage{
		Concurrence: 100,
		HatchRate:   200,
	}
	ret := stage.hatchWorkerCounts()
	assert.Equal(t, ret, []int{100})

	stage = &Stage{
		Concurrence:         100,
		HatchRate:           200,
		previousConcurrence: 200,
	}
	ret = stage.hatchWorkerCounts()
	assert.Equal(t, ret, []int{-100})
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

func TestRunnerConfig_block(t *testing.T) {
	conf := &RunnerConfig{MinWait: ZeroDuration, MaxWait: ZeroDuration}
	start := time.Now()
	conf.block()
	d := time.Since(start)
	assert.True(t, d < time.Millisecond)

	conf = &RunnerConfig{MinWait: time.Second, MaxWait: time.Second}
	start = time.Now()
	conf.block()
	d = time.Since(start)
	assert.True(t, d >= time.Second && d <= time.Second+4*time.Millisecond)

	conf = &RunnerConfig{MinWait: time.Second, MaxWait: 2 * time.Second}
	start = time.Now()
	conf.block()
	d = time.Since(start)
	assert.True(t, !(d < time.Second))
	assert.True(t, !(d > 2*time.Second+2*time.Millisecond))
}

func TestRunnerConfig_checkTime(t *testing.T) {
	conf := &RunnerConfig{MinWait: -1}
	err := conf.check()
	assert.EqualError(t, err, "invalid RunnerConfig.MinWait/MaxWait")

	conf = &RunnerConfig{MaxWait: -1}
	err = conf.check()
	assert.EqualError(t, err, "invalid RunnerConfig.MinWait/MaxWait")

	conf = &RunnerConfig{MinWait: time.Second, MaxWait: time.Millisecond}
	err = conf.check()
	assert.EqualError(t, err, "invalid RunnerConfig.MinWait/MaxWait")
}

func TestRunnerConfig_checkStage(t *testing.T) {
	conf := NewRunnerConfig()
	conf.Concurrence = 0
	err := conf.check()

	assert.EqualError(t, err, "invalid RunnerConfig.Stages")
}

func TestRunnerConfig_checkConcurrence(t *testing.T) {
	conf := NewRunnerConfig()
	conf.AppendStages(
		&Stage{Duration: 100, Concurrence: 100},
		&Stage{Duration: 100, Concurrence: 0},
	)
	err := conf.check()
	assert.EqualError(t, err, "invalid Stage.Concurrency")
}

func TestRunnerConfig_checkHatchRate(t *testing.T) {
	conf := NewRunnerConfig()
	conf.AppendStages(
		&Stage{Duration: 10, Concurrence: 100, HatchRate: 100},
		&Stage{Duration: 10, Concurrence: 100, HatchRate: 0},
	)
	err := conf.check()
	assert.Nil(t, err)

	conf.AppendStage(&Stage{Duration: 10, Concurrence: 100, HatchRate: -1})
	err = conf.check()
	assert.EqualError(t, err, "invalid Stage.HatchRate")
}

func TestRunnerConfig_checkRequests(t *testing.T) {
	conf := NewRunnerConfig()
	conf.AppendStages(
		&Stage{Concurrence: 100, HatchRate: 10, Requests: 100},
		&Stage{Concurrence: 100, HatchRate: 10, Requests: 0},
	)
	err := conf.check()
	assert.Nil(t, err)

	//conf.AppendStage(&Stage{Requests: -1})
	//err = conf.check()
	//assert.EqualError(t, err, "invalid Stage.HatchRate")
}

func TestRunnerConfig_lastStage(t *testing.T) {
	conf := NewRunnerConfig()
	conf.AppendStages(
		&Stage{Concurrence: 100, HatchRate: 100},
		&Stage{Duration: time.Second, Requests: 1000, Concurrence: 200, HatchRate: 10},
	)
	err := conf.check()
	assert.EqualError(t, err, "cannot break current stage")

	conf = NewRunnerConfig()
	conf.AppendStages(
		&Stage{Duration: time.Second, Requests: 1000, Concurrence: 200, HatchRate: 10},
		&Stage{Concurrence: 100, HatchRate: 100},
	)
	err = conf.check()
	assert.Nil(t, err)
}

func TestRunnerConfig_switchStage(t *testing.T) {
	conf := NewRunnerConfig()
	conf.AppendStages(
		&Stage{Duration: time.Second, Requests: 1000, Concurrence: 200, HatchRate: 10},
		&Stage{Duration: time.Second, Requests: 2000, Concurrence: 300, HatchRate: 30},
		&Stage{Concurrence: 100, HatchRate: 100},
	)
	err := conf.check()
	assert.Nil(t, err)

	n, stage := conf.CurrentStage()
	assert.EqualValues(t, n, 0)
	assert.Equal(t, stage, conf.Stages[0])

	n, stage, f := conf.finishCurrentStage(n)
	assert.EqualValues(t, n, 1)
	assert.Equal(t, stage, conf.Stages[1])
	assert.False(t, f)

	i, s2 := conf.CurrentStage()
	assert.Equal(t, i, n)
	assert.Equal(t, s2, stage)

	j, s3, f2 := conf.finishCurrentStage(0)
	assert.Equal(t, j, i)
	assert.Equal(t, s3, s2)
	assert.Equal(t, f2, f)

	i, _, f = conf.finishCurrentStage(i)
	assert.False(t, f)
	i, _, f = conf.finishCurrentStage(i)
	assert.True(t, f)
}

func TestRunnerConfig_maxConcurrence(t *testing.T) {
	conf := NewRunnerConfig()
	conf.AppendStages(
		&Stage{Duration: time.Second, Requests: 1000, Concurrence: 200, HatchRate: 10},
		&Stage{Duration: time.Second, Requests: 2000, Concurrence: 300, HatchRate: 30},
		&Stage{Concurrence: 100, HatchRate: 100},
	)
	err := conf.check()
	assert.Nil(t, err)

	i := conf.findMaxConcurrence()
	assert.EqualValues(t, i, 300)
}
