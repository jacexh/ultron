package ultron

import (
	"errors"
	"math/rand"
	"sync"
	"time"
)

// TaskSet 任务集
type TaskSet struct {
	requests    map[Request]int
	totalWeight int
	MinWait     time.Duration
	MaxWait     time.Duration
	OnStart     func() error
	lock        sync.RWMutex
	ctx         map[string]interface{}
}

// NewTaskSet 新建任务集
func NewTaskSet() *TaskSet {
	return &TaskSet{
		requests: map[Request]int{},
		MinWait:  DefaultMinWait,
		MaxWait:  DefaultMaxWait,
		ctx:      map[string]interface{}{},
	}
}

// Add 添加Query以及权重
func (t *TaskSet) Add(q Request, w int) *TaskSet {
	t.lock.Lock()
	defer t.lock.Unlock()

	q.SetParent(t)

	if w > 0 {
		t.totalWeight += w
	}
	t.requests[q] = w
	return t
}

// PickUp 根据权重获取一个Query对象
func (t *TaskSet) PickUp() Request {
	t.lock.RLock()
	defer t.lock.RUnlock()

	hint := rand.Intn(t.totalWeight) + 1
	for q, w := range t.requests {
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

	delta := ZeroDuration
	if t.MaxWait == t.MinWait || t.MinWait < t.MaxWait {
	} else {
		delta = time.Duration(rand.Int63n(int64(t.MaxWait-t.MinWait)) + 1)
	}
	return t.MinWait + delta
}

// Set 在TaskSet中写入一条可供上下文读取的记录
func (t *TaskSet) Set(key string, value interface{}) {
	t.lock.Lock()
	defer t.lock.Unlock()

	t.ctx[key] = value
}

// Get 在TaskSet中读取一条记录
func (t *TaskSet) Get(key string) interface{} {
	t.lock.RLock()
	defer t.lock.RUnlock()

	if val, ok := t.ctx[key]; ok {
		return val
	}
	return nil

}

func init() {
	rand.Seed(time.Now().UnixNano())
}
