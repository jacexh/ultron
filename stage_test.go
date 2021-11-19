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
