package scheduler

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/wosai/ultron/v2"
)

func TestStageConfiguration_SplitOne(t *testing.T) {
	conf := ultron.StageConfig{
		Duration:    2 * time.Hour,
		Requests:    40000000,
		Concurrence: 1000,
		HatchRate:   200,
		MinWait:     1 * time.Second,
		MaxWait:     3 * time.Second,
	}

	sub := SplitStageConfiguration(conf, 1)
	assert.EqualValues(t, conf, sub[0])
}

func TestStageConfiguration_SplitTwo(t *testing.T) {
	conf := ultron.StageConfig{
		Duration:    2 * time.Hour,
		Requests:    40000000,
		Concurrence: 1000,
		HatchRate:   100,
		MinWait:     1 * time.Second,
		MaxWait:     3 * time.Second,
	}

	subs := SplitStageConfiguration(conf, 2)
	expected := ultron.StageConfig{
		Duration:    2 * time.Hour,
		Requests:    20000000,
		Concurrence: 500,
		HatchRate:   50,
		MinWait:     1 * time.Second,
		MaxWait:     3 * time.Second,
	}
	assert.EqualValues(t, expected, subs[0])
	assert.EqualValues(t, subs[0], subs[1])
}

func TestStageConfiguration_SplitThree(t *testing.T) {
	conf := ultron.StageConfig{
		Duration:    2 * time.Hour,
		Requests:    2000,
		Concurrence: 1000,
		HatchRate:   100,
		MinWait:     1 * time.Second,
		MaxWait:     3 * time.Second,
	}

	subs := SplitStageConfiguration(conf, 3)
	expected := ultron.StageConfig{
		Duration:    2 * time.Hour,
		Requests:    666,
		Concurrence: 333,
		HatchRate:   33,
		MinWait:     1 * time.Second,
		MaxWait:     3 * time.Second,
	}
	assert.EqualValues(t, expected, subs[2])
	assert.EqualValues(t, subs[0], ultron.StageConfig{
		Duration:    2 * time.Hour,
		Requests:    667,
		Concurrence: 334,
		HatchRate:   34,
		MinWait:     1 * time.Second,
		MaxWait:     3 * time.Second,
	})
}
