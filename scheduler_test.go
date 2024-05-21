package ultron

import (
	"context"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/wosai/ultron/v2/pkg/genproto"
)

func TestScheduler_Start(t *testing.T) {
	supervisor := newSlaveSupervisor()
	sa := newSlaveAgent(&genproto.SubscribeRequest{SlaveId: "abc"})
	go func() {
		for range sa.input {
		}
	}()
	supervisor.Add(sa)
	scheduler := newScheduler(supervisor)
	err := scheduler.start(nil)
	assert.NotNil(t, err)

	plan := NewPlan("")
	plan.AddStages(&V1StageConfig{Duration: 1000, ConcurrentUsers: 200})
	err = scheduler.start(plan)
	assert.Nil(t, err)
	i, _ := plan.Current()
	assert.EqualValues(t, i, 0)
}

func TestScheduler_StartWithoudSA(t *testing.T) {
	supervisor := newSlaveSupervisor()
	scheduler := newScheduler(supervisor)
	plan := NewPlan("")
	plan.AddStages(&V1StageConfig{Duration: 1000, ConcurrentUsers: 200})
	err := scheduler.start(plan)
	assert.NotNil(t, err)
}

func TestScheduler_Stop(t *testing.T) {
	supervisor := newSlaveSupervisor()
	sa := newSlaveAgent(&genproto.SubscribeRequest{SlaveId: "abc"})
	go func() {
		for range sa.input {

		}
	}()
	supervisor.Add(sa)
	scheduler := newScheduler(supervisor)

	plan := NewPlan("")
	plan.AddStages(&V1StageConfig{Duration: 1000, ConcurrentUsers: 200})
	err := scheduler.start(plan)
	assert.Nil(t, err)

	err = scheduler.stop(false)
	assert.ErrorIs(t, err, context.DeadlineExceeded) // 聚合超时
}

func TestScheduler_NextStage(t *testing.T) {
	supervisor := newSlaveSupervisor()
	sa := newSlaveAgent(&genproto.SubscribeRequest{SlaveId: "abc"})
	go func() {
		for range sa.input {
		}
	}()
	supervisor.Add(sa)
	scheduler := newScheduler(supervisor)
	err := scheduler.start(nil)
	assert.NotNil(t, err)

	plan := NewPlan("")
	plan.AddStages(&V1StageConfig{Duration: 1000, ConcurrentUsers: 200})
	err = scheduler.start(plan)
	assert.Nil(t, err)

	err = scheduler.nextStage(&V1StageConfig{ConcurrentUsers: 100})
	assert.Nil(t, err)
}

func TestScheduler_StopPatrol(t *testing.T) {
	supervisor := newSlaveSupervisor()
	sa := newSlaveAgent(&genproto.SubscribeRequest{SlaveId: "abc"})
	go func() {
		for range sa.input {
		}
	}()
	supervisor.Add(sa)
	scheduler := newScheduler(supervisor)
	err := scheduler.start(nil)
	assert.NotNil(t, err)

	plan := NewPlan("")
	plan.AddStages(&V1StageConfig{Duration: 1000, ConcurrentUsers: 200})
	err = scheduler.start(plan)
	assert.Nil(t, err)

	go scheduler.patrol(1 * time.Second)
	scheduler.cancel()
}

func TestScheduler_Patrol(t *testing.T) {
	supervisor := newSlaveSupervisor()
	sa := newSlaveAgent(&genproto.SubscribeRequest{SlaveId: "abc"})
	go func() {
		for range sa.input {
		}
	}()
	supervisor.Add(sa)
	scheduler := newScheduler(supervisor)
	err := scheduler.start(nil)
	assert.NotNil(t, err)

	plan := NewPlan("")
	plan.AddStages(&V1StageConfig{Duration: 1000, ConcurrentUsers: 200})
	err = scheduler.start(plan)
	assert.Nil(t, err)

	go scheduler.patrol(1 * time.Second)
	<-time.After(3500 * time.Millisecond)
}
