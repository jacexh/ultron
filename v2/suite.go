package ultron

import (
	"errors"
	"math/rand"
	"sync"
	"time"
)

type (
	// AttackerSuite Attacker集合
	AttackerSuite struct {
		attackers   map[Attacker]int
		totalWeight int
		mu          sync.RWMutex
	}
)

// NewAttackerSuite 创建一个AttackerSuite对象
func NewAttackerSuite() *AttackerSuite {
	return &AttackerSuite{attackers: map[Attacker]int{}}
}

// Add 往AttackerSuite中添加一个Attacker对象, weight 表示该Attacker的权重
func (as *AttackerSuite) Add(a Attacker, weight int) {
	as.mu.Lock()
	defer as.mu.Unlock()

	if weight >= 0 {
		as.totalWeight += weight
		as.attackers[a] = weight
	}
}

// Del 从AttackerSuite中移除一个Attack对象
func (as *AttackerSuite) Del(a Attacker) {
	as.mu.Lock()
	defer as.mu.Unlock()

	if weight, ok := as.attackers[a]; ok {
		as.totalWeight -= weight
		delete(as.attackers, a)
	}
}

func (as *AttackerSuite) pickUp() Attacker {
	hit := rand.Intn(as.totalWeight) + 1

	for a, w := range as.attackers {
		if w > 0 {
			if hit <= w {
				return a
			}
			hit -= w
		}
	}
	panic(errors.New("unreachable code"))
}

func init() {
	rand.Seed(time.Now().UnixNano())
}
