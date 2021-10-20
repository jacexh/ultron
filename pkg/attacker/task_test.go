package attacker

import (
	"fmt"
	"log"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestTask_PickUp(t *testing.T) {
	task := NewTask()
	task.Add(NewHTTPAttacker("task-1"), 5)
	task.Add(NewHTTPAttacker("task-2"), 12)
	task.PickUp()

	actual := make(map[string]int)
	for _, attacker := range task.preempted {
		actual[attacker.Name()]++
	}
	log.Println(actual)
	assert.EqualValues(t, actual["task-1"], 5)
	assert.EqualValues(t, actual["task-2"], 12)
}

func TestTask_PickUp2(t *testing.T) {
	task := NewTask()
	task.Add(NewHTTPAttacker("task-1"), 5)
	task.Add(NewHTTPAttacker("task-2"), 12)

	counter := make(map[string]uint32)

	for i := 0; i < 1000*1000; i++ {
		attacker := task.PickUp()
		counter[attacker.Name()] += 1
	}
	fmt.Println(counter)
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
