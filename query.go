package ultron

import (
	"errors"
	"math/rand"
	"sync"
	"time"
)

const (
	// NoDuration 无等待
	NoDuration time.Duration = time.Duration(0)
	// DefaultMinWait 默认最小等待时间
	DefaultMinWait time.Duration = time.Second * 1
	// DefaultMaxWait 默认最大等待时间
	DefaultMaxWait time.Duration = time.Second * 5
)

type (
	// Query .
	Query interface {
		Name() string
		Fire() (time.Duration, error)
	}

	// // QuerySet 查询事件集合
	// QuerySet interface {
	// 	OnStart()
	// 	Queries() map[Query]int
	// 	MinWait() time.Duration
	// 	MaxWait() time.Duration
	// }

	// TaskSet 任务集
	TaskSet struct {
		queries     map[Query]int
		MinWait     time.Duration
		MaxWait     time.Duration
		totalWeight int
		lock        *sync.RWMutex
	}
)

// NewTaskSet 新建任务集
func NewTaskSet() *TaskSet {
	return &TaskSet{
		queries:     map[Query]int{},
		MinWait:     DefaultMinWait,
		MaxWait:     DefaultMaxWait,
		totalWeight: 0,
		lock:        &sync.RWMutex{},
	}
}

// Add 添加Query以及权重
func (t *TaskSet) Add(q Query, w int) *TaskSet {
	t.lock.Lock()
	defer t.lock.Unlock()

	if w > 0 {
		t.totalWeight += w
	}
	t.queries[q] = w
	return t
}

// Choice 根据权重获取一个Query对象
func (t *TaskSet) Choice() Query {
	t.lock.RLock()
	defer t.lock.RUnlock()

	hint := rand.Intn(t.totalWeight) + 1
	for q, w := range t.queries {
		if w > 0 {
			if hint <= w {
				return q
			}
			hint -= w
		}
	}
	panic(errors.New("what happened?"))
}

// Wait return wait time
func (t *TaskSet) Wait() time.Duration {
	t.lock.RLock()
	defer t.lock.RUnlock()

	delta := NoDuration
	if t.MaxWait == t.MinWait || t.MinWait < t.MaxWait {
	} else {
		delta = time.Duration(rand.Int63n(int64(t.MaxWait-t.MinWait)) + 1)
	}
	return t.MinWait + delta
}
