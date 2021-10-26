package ultron

import (
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
	assert.EqualValues(t, subs, []AttackStrategy{
		&FixedConcurrentUsers{ConcurrentUsers: 334, RampUpPeriod: 3},
		&FixedConcurrentUsers{ConcurrentUsers: 333, RampUpPeriod: 3},
		&FixedConcurrentUsers{ConcurrentUsers: 333, RampUpPeriod: 3},
	})
}

func TestAttackStrategyConverter(t *testing.T) {
	converter := newAttackStrategyConverter()
	as := &FixedConcurrentUsers{ConcurrentUsers: 200, RampUpPeriod: 4}
	dto, err := converter.ConvertAttackStrategy(as)
	assert.Nil(t, err)
	as2, err := converter.ConvertDTO(dto)
	assert.Nil(t, err)
	assert.EqualValues(t, as, as2)
}
