package ultron

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

type (
	toTestAttacker struct {
		name string
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
	return nil
}

func BenchmarkPickUp(b *testing.B) {
	task := NewTask()
	task.Add(newAttacker("a"), 10)
	task.Add(newAttacker("b"), 20)
	task.Add(newAttacker("c"), 3)
	for i := 0; i < b.N; i++ {
		task.pickUp()
	}
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
