package ultron

import (
	"fmt"
	"sync"
	"sync/atomic"
)

type (
	// Task Attacker集合
	Task struct {
		attackers   []*smoothAttack
		totalWeight uint64
		once        sync.Once
		counts      uint64
		preempted   []Attacker
	}

	// smoothAttack 平滑的加权请求
	smoothAttack struct {
		attacker      Attacker
		weight        int
		currentWeight int
	}
)

// NewTask 创建一个Task对象
func NewTask() *Task {
	return &Task{attackers: []*smoothAttack{}, totalWeight: 0}
}

// Add 往Task中添加一个Attacker对象, weight 表示该Attacker的权重
func (t *Task) Add(a Attacker, weight int) {
	if weight <= 0 {
		Logger.Warn(fmt.Sprintf("Attacker named %s with invalid weight: %d", a.Name(), weight))
		return
	}

	t.totalWeight += uint64(weight)
	sa := &smoothAttack{attacker: a, weight: weight}
	t.attackers = append(t.attackers, sa)
}

// Del 从Task中移除一个Attacker对象
func (t *Task) Del(a Attacker) {
	for i, attacker := range t.attackers {
		if attacker.attacker == a {
			switch i {
			case 0:
				t.attackers = t.attackers[i+1:]
			case len(t.attackers) - 1:
				t.attackers = t.attackers[:i]
			default:
				t.attackers = append(t.attackers[:i], t.attackers[i+1:]...)
			}
			t.totalWeight -= uint64(attacker.weight)
			return
		}
	}
}

func (t *Task) smoothWeigh() Attacker {
	maxIndex := 0
	maxWeight := 0
	for i, attacker := range t.attackers {
		attacker.currentWeight += attacker.weight
		if attacker.currentWeight > maxWeight {
			maxWeight = attacker.currentWeight
			maxIndex = i
		}
	}
	sa := t.attackers[maxIndex]
	sa.currentWeight -= int(t.totalWeight)
	return sa.attacker
}

// pickUp 按照权重获取attacker
func (t *Task) pickUp() Attacker {
	t.once.Do(func() {
		t.preempted = make([]Attacker, int(t.totalWeight))
		for i := 0; i < int(t.totalWeight); i++ {
			t.preempted[i] = t.smoothWeigh()
		}
	})
	attacker := t.preempted[atomic.LoadUint64(&t.counts)%t.totalWeight]
	atomic.AddUint64(&t.counts, 1)
	return attacker
}
