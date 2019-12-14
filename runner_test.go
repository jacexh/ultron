package ultron

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestBaseRunner_WithConfig(t *testing.T) {
	runner := newBaseRunner()
	conf := NewRunnerConfig()
	runner.WithConfig(conf)
	assert.Equal(t, runner.GetConfig(), conf)
}

func TestBaseRunner_WithTask(t *testing.T) {
	runner := newBaseRunner()
	task := NewTask()
	runner.WithTask(task)
	assert.Equal(t, runner.task, task)
}

func TestBaseRunner_getStatus(t *testing.T) {
	runner := newBaseRunner()
	assert.Equal(t, runner.GetStatus(), StatusIdle)

	runner.Done()
	assert.Equal(t, runner.GetStatus(), StatusStopped)
}

func TestBaseRunner_check(t *testing.T) {
	runner := newBaseRunner()
	err := checkRunner(runner)
	assert.EqualError(t, err, "no Task provided")

	task := NewTask()
	runner.WithTask(task)
	runner.WithConfig(nil)
	err = checkRunner(runner)
	assert.EqualError(t, err, "no RunnerConfig provided")

	runner.WithConfig(NewRunnerConfig())
	err = checkRunner(runner)
	assert.Nil(t, err)
}

func TestBaseRunner_isFinishedCurrentStage(t *testing.T) {
	runner := newBaseRunner()
	runner.Config.AppendStages(
		&Stage{Concurrence: 200, Duration: 5 * time.Minute, HatchRate: 10, Requests: 1000},
		&Stage{Concurrence: 300, Duration: 6 * time.Minute, HatchRate: 20, Requests: 2000},
	)

	s, r := runner.isFinishedCurrentStage()
	assert.False(t, s)
	assert.False(t, r)

	runner.Config.Stages[0].counts = 10001 // 满足条件
	s, r = runner.isFinishedCurrentStage()
	assert.True(t, s)
	assert.False(t, r)

	s, r = runner.isFinishedCurrentStage() // stage 2
	assert.False(t, s)
	assert.False(t, r)

	runner.Config.Stages[1].expired = StageExpired
	s, r = runner.isFinishedCurrentStage()
	assert.True(t, s)
	assert.True(t, r)
}

func TestBaseRunner_isFinishedCurrentStageStatusDone(t *testing.T) {
	runner := newBaseRunner()
	runner.Config.AppendStages(
		&Stage{Concurrence: 200, Duration: 5 * time.Minute, HatchRate: 10, Requests: 1000},
		&Stage{Concurrence: 300, Duration: 6 * time.Minute, HatchRate: 20, Requests: 2000},
	)

	runner.Done()
	s, r := runner.isFinishedCurrentStage()
	assert.True(t, s)
	assert.True(t, r)
}
