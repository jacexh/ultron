package ultron

import (
	"net"
	"net/http"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestMasterRunner_Launch(t *testing.T) {
	runner := NewMasterRunner()
	err := runner.Launch()
	assert.Nil(t, err)

	<-time.After(1 * time.Second)
	resp, err := http.Get("http://localhost:2017")
	assert.Nil(t, err)
	assert.EqualValues(t, resp.StatusCode, 200)

	_, err = net.Listen("TCP", ":2021")
	assert.NotNil(t, err)
}

func TestMasterRunner_StartPlan(t *testing.T) {
	runner := NewMasterRunner()
	err := runner.Launch()
	assert.Nil(t, err)
	<-time.After(1 * time.Second)

	err = runner.StartPlan(nil)
	assert.NotNil(t, err)

	plan := NewPlan("foobar")
	plan.AddStages(&V1StageConfig{ConcurrentUsers: 100})
	err = runner.StartPlan(plan)
	assert.NotNil(t, err.Error(), "cannot batch send event to empty slave agent")
}

func TestMasterRunner_StopPlan(t *testing.T) {
	runner := NewMasterRunner()
	err := runner.Launch()
	assert.Nil(t, err)
	<-time.After(1 * time.Second)

	runner.StopPlan()
}
