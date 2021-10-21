package ultron

import (
	"context"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestFixedConcurrentUsers_Spawn(t *testing.T) {
	s := &FixedConcurrentUsers{
		ConcurrentUsers: 1000,
		RampUpPeriod:    3,
	}
	waves := s.Spawn()
	assert.EqualValues(t, waves, []*RampUpStep{
		{N: 333, Interval: 1 * time.Second},
		{N: 333, Interval: 1 * time.Second},
		{N: 334, Interval: 1 * time.Second},
	})
}

func TestFixedConcurrentUsers_Switch(t *testing.T) {
	s1 := &FixedConcurrentUsers{
		ConcurrentUsers: 1000,
		RampUpPeriod:    3,
	}
	s2 := &FixedConcurrentUsers{
		ConcurrentUsers: 600,
		RampUpPeriod:    6,
	}
	waves := s1.Switch(s2)
	assert.EqualValues(t, waves, []*RampUpStep{
		{N: -66, Interval: 1 * time.Second},
		{N: -66, Interval: 1 * time.Second},
		{N: -66, Interval: 1 * time.Second},
		{N: -66, Interval: 1 * time.Second},
		{N: -66, Interval: 1 * time.Second},
		{N: -70, Interval: 1 * time.Second},
	})
}

func TestFixedConcurrentUsers_Spilt(t *testing.T) {
	fx := &FixedConcurrentUsers{
		ConcurrentUsers: 1000,
		RampUpPeriod:    3,
	}
	subs := fx.Split(3)
	assert.EqualValues(t, subs, []AttackStrategyDescriber{
		&FixedConcurrentUsers{ConcurrentUsers: 334, RampUpPeriod: 3},
		&FixedConcurrentUsers{ConcurrentUsers: 333, RampUpPeriod: 3},
		&FixedConcurrentUsers{ConcurrentUsers: 333, RampUpPeriod: 3},
	})
}

func TestFCUExecutor(t *testing.T) {
	commander := newFixedConcurrentUsersStrategyCommander()
	task := NewTask()
	task.Add(&fakeAttacker{}, 10)

	output := commander.Open(context.Background(), task)
	go func() {
		for range output {
		}
	}()

	commander.Command(&FixedConcurrentUsers{ConcurrentUsers: 50, RampUpPeriod: 3}, NonstopTimer{})
	<-time.After(2 * time.Second)
	commander.Command(&FixedConcurrentUsers{ConcurrentUsers: 80, RampUpPeriod: 5}, NonstopTimer{})
	<-time.After(2 * time.Second)
	commander.Command(&FixedConcurrentUsers{ConcurrentUsers: 30, RampUpPeriod: 7}, NonstopTimer{})
	commander.Close()
}
