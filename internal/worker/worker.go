package worker

import (
	"context"
	"math/rand"
	"runtime/debug"
	"sync"
	"time"

	"github.com/wosai/ultron/pkg/attacker"
	"github.com/wosai/ultron/pkg/statistics"
)

type (
	Worker interface {
		Start(<-chan attacker.Attacker, chan<- *statistics.StatisticianGroup)
		Kill()
	}

	worker struct {
		planCtx context.Context
		cancel  context.CancelFunc
		task    *attacker.Task
		sg      *statistics.StatisticianGroup
		min     time.Duration
		max     time.Duration
	}

	WorkerFacotry interface {
		Build() <-chan Worker
		Recyle() <-chan struct{}
	}
)

func (w *worker) Kill() {
	w.cancel()
}

func (w *worker) Start(input <-chan attacker.Attacker, output chan<- *statistics.AttackResult) {
	var attacker attacker.Attacker
	var err error

	for {
		select {
		case <-w.planCtx.Done():
			return
		case attacker = <-input:
		}

		start := time.Now()
		err = attacker.Fire(w.planCtx)
		output <- &statistics.AttackResult{
			Name:     attacker.Name(),
			Duration: time.Since(start),
			Error:    err,
		}

		select {
		case <-w.planCtx.Done():
			return
		default:
		}

		if w.max > 0 {
			time.Sleep(w.min + time.Duration(rand.Int63n(int64(w.max-w.min)+1)))
		}
	}
}

func (w *worker) DoWork(min, max time.Duration) {
	var wg sync.WaitGroup
	wg.Add(1)
	defer func() {
		if rec := recover(); rec != nil {
			debug.PrintStack()
		}
		wg.Done()
	}()

	var err error
	for {
		attacker := w.task.PickUp()
		start := time.Now()

		select {
		case <-w.planCtx.Done():
			return
		default:
		}

		err = attacker.Fire(w.planCtx)
		w.sg.Record(&statistics.AttackResult{
			Name:     attacker.Name(),
			Duration: time.Since(start),
			Error:    err,
		})

		select {
		case <-w.planCtx.Done():
			return
		default:
		}

		if max > 0 {
			time.Sleep(min + time.Duration(rand.Int63n(int64(max-min)+1)))
		}
	}
}
