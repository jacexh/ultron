package attacker

import (
	"sync"
	"sync/atomic"
)

type (
	Task struct {
		attacker    []*smoothAttacker
		totalWeight uint32
		counts      uint32
		preempted   []Attacker
		once        sync.Once
	}

	// smoothAttacker 平滑的加权请求
	smoothAttacker struct {
		attacker      Attacker
		weight        uint32
		currentWeight uint32
	}
)

func NewTask() *Task {
	return &Task{
		attacker: make([]*smoothAttacker, 0),
	}
}

func (t *Task) Add(a Attacker, weight uint32) {
	t.totalWeight += weight
	t.attacker = append(t.attacker, &smoothAttacker{
		attacker: a,
		weight:   weight,
	})
}

// https://en.wikipedia.org/wiki/Weighted_round_robin
func (t *Task) swrr() Attacker {
	var maxIndex int = 0
	var maxWeight uint32 = 0

	for i, attacker := range t.attacker {
		attacker.currentWeight += attacker.weight
		if attacker.currentWeight > maxWeight {
			maxWeight = attacker.currentWeight
			maxIndex = i
		}
	}

	sa := t.attacker[maxIndex]
	sa.currentWeight -= t.totalWeight
	return sa.attacker
}

func (t *Task) PickUp() Attacker {
	t.once.Do(func() {
		t.preempted = make([]Attacker, int(t.totalWeight))
		for i := 0; i < int(t.totalWeight); i++ {
			t.preempted[i] = t.swrr()
		}
	})
	v := atomic.AddUint32(&t.counts, 1)
	return t.preempted[(v-1)%t.totalWeight]
}
