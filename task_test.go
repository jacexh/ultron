package ultron

import (
	"time"
	"testing"
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
