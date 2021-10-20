package attacker

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestTask_PickUp(t *testing.T) {
	task := NewTask()
	task.Add(NewHTTPAttacker("task-1"), 5)
	task.Add(NewHTTPAttacker("task-2"), 10)
	task.PickUp()

	actual := make(map[string]int)
	for _, attacker := range task.preempted {
		actual[attacker.Name()]++
	}
	assert.EqualValues(t, actual["task-1"], 5)
	assert.EqualValues(t, actual["task-2"], 10)
}

func BenchmarkTest_PickUp(b *testing.B) {
	task := NewTask()
	task.Add(NewHTTPAttacker("task-1"), 5)
	task.Add(NewHTTPAttacker("task-2"), 10)
	task.Add(NewHTTPAttacker("task-3"), 20)

	b.RunParallel(func(p *testing.PB) {
		for p.Next() {
			task.PickUp()
		}
	})
}
