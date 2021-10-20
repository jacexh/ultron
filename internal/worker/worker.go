package worker

import (
	"context"
	"log"
	"math/rand"
	"sync"
	"sync/atomic"
	"time"

	"github.com/wosai/ultron/v2"
	"github.com/wosai/ultron/v2/pkg/attacker"
	"github.com/wosai/ultron/v2/pkg/statistics"
)

type (
	WorkShop interface {
		Open(context.Context, *attacker.Task) <-chan *statistics.AttackResult
		Execute(ultron.StageConfig)
		Close()
	}

	FixedSizeWorkShop struct {
		ctx     context.Context
		cancel  context.CancelFunc
		config  ultron.StageConfig
		output  chan *statistics.AttackResult
		task    *attacker.Task
		counter uint32
		pool    map[uint32]*simpleWorker
		wg      *sync.WaitGroup
	}

	simpleWorker struct {
		id     uint32
		cancel context.CancelFunc
		parent *FixedSizeWorkShop
	}

	Timer interface {
		Sleep()
	}

	ConcurrencePolicy interface {
		RampUp() int
		RampDown() int
	}
)

// todo:
func (sw *simpleWorker) start(ctx context.Context, task *attacker.Task, output chan<- *statistics.AttackResult) error {
	ctx, sw.cancel = context.WithCancel(ctx)
	defer func() {
		if rec := recover(); rec != nil {
			log.Println(rec)
		}
	}()

	for {
		select {
		case <-ctx.Done():
			return ctx.Err()
		default:
		}

		start := time.Now()
		att := task.PickUp()
		err := att.Fire(ctx)
		output <- &statistics.AttackResult{
			Name:     att.Name(),
			Duration: time.Since(start),
			Error:    err,
		}

		select {
		case <-ctx.Done():
			return ctx.Err()
		default:
		}

		if sw.parent.config.MaxWait > 0 {
			time.Sleep(sw.parent.config.MinWait + time.Duration(rand.Int63n(int64(sw.parent.config.MaxWait-sw.parent.config.MinWait))))
		}
	}
}

func (sw *simpleWorker) kill() {
	sw.cancel()
}

func NewFixedSizeWorkShop() WorkShop {
	return &FixedSizeWorkShop{
		output: make(chan *statistics.AttackResult, 100),
		pool:   make(map[uint32]*simpleWorker),
		wg:     new(sync.WaitGroup),
	}

}

func (fs *FixedSizeWorkShop) Open(ctx context.Context, task *attacker.Task) <-chan *statistics.AttackResult {
	fs.ctx, fs.cancel = context.WithCancel(ctx)
	fs.task = task
	return fs.output
}

func (fs *FixedSizeWorkShop) Execute(config ultron.StageConfig) {
	fs.config = config

	for i := 0; i < config.Concurrence; i++ {
		worker := &simpleWorker{
			id:     atomic.AddUint32(&fs.counter, 1) - 1,
			parent: fs,
		}
		fs.pool[worker.id] = worker
		go func(w *simpleWorker) {
			fs.wg.Add(1)
			defer fs.wg.Done()

			if err := worker.start(fs.ctx, fs.task, fs.output); err != nil {
			}
		}(worker)
	}
}

func (fs *FixedSizeWorkShop) Close() {
	fs.cancel()
	fs.wg.Wait()
	close(fs.output)
}
