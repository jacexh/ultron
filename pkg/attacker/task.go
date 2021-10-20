package attacker

import (
	"sync"
	"sync/atomic"
)

type (
	Task struct {
		attacker    []*swrrAttacker
		totalWeight uint32
		counts      uint32
		preempted   []Attacker
		once        sync.Once
	}

	// swrrAttacker 平滑的加权请求
	swrrAttacker struct {
		attacker Attacker
		weight   uint32
		current  int32
	}
)

func NewTask() *Task {
	return &Task{
		attacker: make([]*swrrAttacker, 0),
	}
}

func (t *Task) Add(a Attacker, weight uint32) {
	t.totalWeight += weight
	t.attacker = append(t.attacker, &swrrAttacker{
		attacker: a,
		weight:   weight,
	})
}

// https://tenfy.cn/2018/11/12/smooth-weighted-round-robin/
func (t *Task) swrr() Attacker {
	var best *swrrAttacker

	for _, attacker := range t.attacker {
		attacker.current += int32(attacker.weight)
		if best == nil || attacker.current > best.current {
			best = attacker
		}
	}

	best.current -= int32(t.totalWeight)
	return best.attacker
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
