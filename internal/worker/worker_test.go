package worker

import (
	"context"
	"encoding/json"
	"errors"
	"log"
	"runtime"
	"sync"
	"sync/atomic"
	"testing"
	"time"

	"github.com/wosai/ultron/pkg/attacker"
	"github.com/wosai/ultron/pkg/statistics"
	"github.com/wosai/ultron/types"
)

type (
	forLoopWorker struct {
		ctx context.Context
	}

	channelWorker struct {
		ctx context.Context
	}

	transaction struct {
		name string
	}
)

func (t transaction) Name() string {
	return t.name
}

func (t transaction) Fire(ctx context.Context) error {
	select {
	case <-ctx.Done():
		return ctx.Err()
	default:
	}

	time.Sleep(1 * time.Millisecond)
	return nil
}

func (fw *forLoopWorker) do(task *attacker.Task, stats *statistics.StatisticianGroup, c *uint32) {
	var err error
	for {
		select {
		case <-fw.ctx.Done():
			return
		default:
		}

		at := task.PickUp()
		start := time.Now()
		err = at.Fire(fw.ctx)
		stats.Record(&statistics.AttackResult{
			Name:     at.Name(),
			Duration: time.Since(start),
			Error:    err,
		})
		atomic.AddUint32(c, 1)

		select {
		case <-fw.ctx.Done():
			return
		default:
		}
	}
}

func (cw *channelWorker) do(input <-chan attacker.Attacker, output chan<- *statistics.AttackResult) error {
	var at attacker.Attacker
	var err error
	var opening bool

	for {
		select {
		case <-cw.ctx.Done():
			return cw.ctx.Err()
		case <-time.After(3 * time.Second):
			return errors.New("kill lazy worker")
		case at, opening = <-input:
			if !opening {
				return errors.New("input channel closed")
			}
		}

		start := time.Now()
		err = at.Fire(cw.ctx)
		output <- &statistics.AttackResult{
			Name:     at.Name(),
			Duration: time.Since(start),
			Error:    err,
		}

		select {
		case <-cw.ctx.Done():
			return cw.ctx.Err()
		default:
		}
	}
}

func (cw *channelWorker) do2(input <-chan attacker.Attacker, stats *statistics.StatisticianGroup) error {
	var at attacker.Attacker
	var err error
	var opening bool

	for {
		select {
		case <-cw.ctx.Done():
			return cw.ctx.Err()
		case <-time.After(3 * time.Second):
			return errors.New("kill lazy worker")
		case at, opening = <-input:
			if !opening {
				return errors.New("input channel closed")
			}
		}

		start := time.Now()
		err = at.Fire(cw.ctx)
		stats.Record(&statistics.AttackResult{
			Name:     at.Name(),
			Duration: time.Since(start),
			Error:    err,
		})

		select {
		case <-cw.ctx.Done():
			return cw.ctx.Err()
		default:
		}
	}
}

func TestChannelWork(t *testing.T) {
	var wg = &sync.WaitGroup{}
	sg := statistics.NewStatisticianGroup()
	output := make(chan *statistics.AttackResult, 100)
	input := make(chan attacker.Attacker, 100)
	task := attacker.NewTask()
	task.Add(transaction{}, 5)

	go func() {
		for result := range output {
			sg.Record(result)
		}
	}()

	for i := 0; i < 100; i++ {
		go func() {
			wg.Add(1)
			defer wg.Done()

			cw := &channelWorker{ctx: context.Background()}
			cw.do(input, output)
		}()
	}

	for i := 0; i < 1000*1000; i++ {
		input <- task.PickUp()
	}
	close(input)
	wg.Wait()
	close(output)
	report := sg.Report(true)
	log.Println(report.TotalTPS)
}

func TestChannelWork2(t *testing.T) {
	var wg = &sync.WaitGroup{}
	sg := statistics.NewStatisticianGroup()
	input := make(chan attacker.Attacker, 100)
	task := attacker.NewTask()
	task.Add(transaction{}, 5)

	for i := 0; i < 100; i++ {
		go func() {
			wg.Add(1)
			defer wg.Done()

			cw := &channelWorker{ctx: context.Background()}
			cw.do2(input, sg)
		}()
	}

	for i := 0; i < 1000*1000; i++ {
		input <- task.PickUp()
	}
	close(input)
	wg.Wait()
	report := sg.Report(true)
	log.Println(report.TotalTPS)
}

func TestForLoopWorker(t *testing.T) {
	var wg = &sync.WaitGroup{}
	sg := statistics.NewStatisticianGroup()
	task := attacker.NewTask()
	task.Add(transaction{}, 5)

	ctx, cancel := context.WithCancel(context.Background())
	var counts uint32

	for i := 0; i < 100; i++ {
		go func() {
			wg.Add(1)
			defer wg.Done()

			cw := &forLoopWorker{ctx: ctx}
			cw.do(task, sg, &counts)
		}()
	}
	for {
		if atomic.LoadUint32(&counts) >= 1000*1000 {
			cancel()
			break
		}
		runtime.Gosched()
	}
	wg.Wait()
	report := sg.Report(true)
	log.Println(report.TotalTPS)
}

func TestFixedSizeWorkShop_Finish(t *testing.T) {
	wr := NewFixedSizeWorkShop()
	sg := statistics.NewStatisticianGroup()
	task := attacker.NewTask()
	task.Add(transaction{name: "test-0"}, 5)
	task.Add(transaction{name: "test-1"}, 12)

	output := wr.Start(context.Background(), task)
	go wr.Execute(types.StageConfig{
		Concurrence: 200,
		MinWait:     10 * time.Millisecond,
		MaxWait:     15 * time.Millisecond,
	})

	go func() {
		<-time.After(3 * time.Second)
		wr.Finish()
	}()

	for ret := range output {
		sg.Record(ret)
	}

	report := sg.Report(true)
	data, _ := json.Marshal(report)
	log.Println(string(data))
}
