package ultron

import (
	"errors"
	"fmt"
	"math/rand"
	"sync"
)

type (
	// Task Attacker集合
	Task struct {
		attackers   map[Attacker]int
		totalWeight int
		mu          sync.RWMutex
	}
)

// NewTask 创建一个Task对象
func NewTask() *Task {
	return &Task{attackers: map[Attacker]int{}}
}

// Add 往Task中添加一个Attacker对象, weight 表示该Attacker的权重
func (t *Task) Add(a Attacker, weight int) {
	t.mu.Lock()
	defer t.mu.Unlock()

	if weight <= 0 {
		Logger.Warn(fmt.Sprintf("Attacker named %s with invalid weight: %d", a.Name(), weight))
		return
	}

	t.totalWeight += weight
	t.attackers[a] = weight
}

// Del 从Task中移除一个Attacker对象
func (t *Task) Del(a Attacker) {
	t.mu.Lock()
	defer t.mu.Unlock()

	if weight, ok := t.attackers[a]; ok {
		t.totalWeight -= weight
		delete(t.attackers, a)
	}
}

func (t *Task) pickUp() Attacker {
	hit := rand.Intn(t.totalWeight) + 1

	for a, w := range t.attackers {
		if hit <= w {
			return a
		}
		hit -= w
	}
	panic(errors.New("unreachable code"))
}
