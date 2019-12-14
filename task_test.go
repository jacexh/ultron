package ultron

import (
	"fmt"
	"sync/atomic"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

type (
	toTestAttacker struct {
		name   string
		counts uint32
	}
)

func newAttacker(n string) *toTestAttacker {
	return &toTestAttacker{name: n}
}

func (t *toTestAttacker) Name() string {
	return t.name
}

func (t *toTestAttacker) Fire() error {
	time.Sleep(time.Millisecond)
	atomic.AddUint32(&t.counts, 1)
	return nil
}

func BenchmarkPickUp(b *testing.B) {
	task := NewTask()
	a := newAttacker("a")
	d := newAttacker("d")
	c := newAttacker("c")
	task.Add(a, 10)
	task.Add(d, 20)
	task.Add(c, 3)

	for i := 0; i < b.N; i++ {
		switch task.pickUp() {
		case a:
			atomic.AddUint32(&a.counts, 1)
		case d:
			atomic.AddUint32(&d.counts, 1)
		case c:
			atomic.AddUint32(&c.counts, 1)
		}
	}
	fmt.Printf("%d - %d - %d\n", a.counts, d.counts, c.counts)
}

func BenchmarkPickUpParallel(b *testing.B) {
	task := NewTask()
	a := newAttacker("a")
	d := newAttacker("d")
	c := newAttacker("c")
	task.Add(a, 10)
	task.Add(d, 20)
	task.Add(c, 3)
	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			switch task.pickUp() {
			case a:
				atomic.AddUint32(&a.counts, 1)
			case d:
				atomic.AddUint32(&d.counts, 1)
			case c:
				atomic.AddUint32(&c.counts, 1)
			}
		}
	})

}

func TestTask_Add(t *testing.T) {
	task := NewTask()
	attacker := newAttacker("hello")
	task.Add(attacker, 1)
	actual := task.pickUp()
	assert.Equal(t, attacker, actual)
}

func TestTask_NoAttacker(t *testing.T) {
	task := NewTask()
	attacker := newAttacker("hello")
	task.Add(attacker, 1)
	task.Del(attacker)
	assert.Panics(t, func() { task.pickUp() })
}

func TestTask_BadWeight(t *testing.T) {
	task := NewTask()
	attacker := newAttacker("hello")
	task.Add(attacker, -1)
	assert.Panics(t, func() {
		task.pickUp()
	})
}
