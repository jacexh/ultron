package ultron

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestStage_WithTimer(t *testing.T) {
	stage := BuildStage()
	stage.WithTimer(nil)
	timer := stage.GetTimer()
	assert.NotNil(t, timer)
}

func TestStage_WithEC(t *testing.T) {
	stage := BuildStage()
	stage.WithExitConditions(nil)
	assert.NotNil(t, stage.GetExitConditions())
}

func TestUniversalExitConditions_NeverStop(t *testing.T) {
	ec := &UniversalExitConditions{}
	assert.True(t, ec.NeverStop())

	ec.Duration = 1 * time.Second
	assert.False(t, ec.NeverStop())
}

func TestUniversalExitConditions_Check(t *testing.T) {
	expected := &UniversalExitConditions{Requests: 1000}
	actual := &UniversalExitConditions{Requests: 500}
	assert.False(t, expected.Check(actual))
	assert.True(t, expected.Check(&UniversalExitConditions{Requests: 1000}))

	expected = &UniversalExitConditions{Duration: 2 * time.Second}
	assert.False(t, expected.Check(&UniversalExitConditions{Duration: 1 * time.Second}))
	assert.True(t, expected.Check(&UniversalExitConditions{Duration: 3 * time.Second}))

	expected = &UniversalExitConditions{}
	assert.False(t, expected.Check(&UniversalExitConditions{Requests: 1000, Duration: 5 * time.Second}))
}

func TestV1StageConfig(t *testing.T) {
	conf := &V1StageConfig{
		Requests:        1000,
		Duration:        3 * time.Second,
		ConcurrentUsers: 100,
		RampUpPeriod:    10,
		MinWait:         3 * time.Second,
		MaxWait:         5 * time.Second,
	}

	timer := conf.GetTimer()
	assert.EqualValues(t, timer, &UniformRandomTimer{MinWait: 3 * time.Second, MaxWait: 5 * time.Second})
	ec := conf.GetExitConditions()
	assert.EqualValues(t, ec, &UniversalExitConditions{Requests: 1000, Duration: 3 * time.Second})
	as := conf.GetStrategy()
	assert.EqualValues(t, as, &FixedConcurrentUsers{ConcurrentUsers: 100, RampUpPeriod: 10})
}
