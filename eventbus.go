package ultron

import (
	"context"
	"runtime/debug"
	"sync"
	"sync/atomic"

	"github.com/wosai/ultron/pkg/statistics"
	"go.uber.org/zap"
)

type (
	// ReportHandleFunc 聚合报告处理函数
	ReportHandleFunc func(context.Context, statistics.SummaryReport)

	// reportBus 聚合报告事件总线
	reportBus interface {
		subscribeReport(ReportHandleFunc)
		publishReport(statistics.SummaryReport)
	}

	// ResultHandleFunc 请求结果处理函数
	ResultHandleFunc func(context.Context, statistics.AttackResult)

	// resultBus 压测结果事件总线
	resultBus interface {
		subscribeResult(ResultHandleFunc)
		publishResult(statistics.AttackResult)
	}

	eventbus struct {
		cancel                context.CancelFunc
		reportBus             chan statistics.SummaryReport
		resultBuses           []chan statistics.AttackResult
		reportHandlers        []ReportHandleFunc
		resultHandlers        []ResultHandleFunc
		numberOfSubchannels   uint32
		counterForSubchannels uint32
		closed                uint32
		once                  sync.Once
		wg                    sync.WaitGroup
	}
)

var (
	defaultEventBus *eventbus
	_               reportBus = (*eventbus)(nil)
	_               resultBus = (*eventbus)(nil)
)

func newEventBus() *eventbus {
	bus := &eventbus{
		reportBus:           make(chan statistics.SummaryReport, 3), // 低频通道
		reportHandlers:      make([]ReportHandleFunc, 0),
		resultHandlers:      make([]ResultHandleFunc, 0),
		numberOfSubchannels: 30,
	}
	bus.resultBuses = make([]chan statistics.AttackResult, bus.numberOfSubchannels)
	for i := 0; i < int(bus.numberOfSubchannels); i++ {
		bus.resultBuses[i] = make(chan statistics.AttackResult, 200)
	}
	return bus
}

func (bus *eventbus) subscribeReport(fn ReportHandleFunc) {
	if fn == nil {
		return
	}
	bus.reportHandlers = append(bus.reportHandlers, fn)
}

func (bus *eventbus) publishReport(report statistics.SummaryReport) {
	if atomic.LoadUint32(&bus.closed) == 0 {
		bus.reportBus <- report
	}
}

func (bus *eventbus) subscribeResult(fn ResultHandleFunc) {
	if fn == nil {
		return
	}
	bus.resultHandlers = append(bus.resultHandlers, fn)
}

func (bus *eventbus) publishResult(ret statistics.AttackResult) {
	if atomic.LoadUint32(&bus.closed) == 0 {
		v := atomic.AddUint32(&bus.counterForSubchannels, 1)
		bus.resultBuses[int((v-1)%bus.numberOfSubchannels)] <- ret
	}
}

func (bus *eventbus) start() {
	bus.once.Do(func() {
		ctx, cancel := context.WithCancel(context.Background())
		bus.cancel = cancel

		bus.wg.Add(1)
		go func() {
			defer func() {
				if rec := recover(); rec != nil {
					debug.PrintStack()
					Logger.DPanic("report bus is quit", zap.Any("recover", rec))
				}
				bus.wg.Done()
				bus.close()
			}()

			for report := range bus.reportBus {
				for _, fn := range bus.reportHandlers {
					fn(ctx, report)
				}
			}
		}()

		for _, sub := range bus.resultBuses {
			bus.wg.Add(1)
			go func(c <-chan statistics.AttackResult) {
				defer func() {
					if rec := recover(); rec != nil {
						debug.PrintStack()
						Logger.DPanic("one result bus is quit", zap.Any("recover", rec))
					}
					bus.wg.Done()
					bus.close()
				}()

				for result := range c {
					for _, handler := range bus.resultHandlers {
						handler(ctx, result)
					}
				}
			}(sub)
		}
	})
}

func (bus *eventbus) close() {
	if atomic.CompareAndSwapUint32(&bus.closed, 0, 1) {
		if bus.cancel != nil {
			bus.cancel()
		}

		close(bus.reportBus)
		for _, sub := range bus.resultBuses {
			close(sub)
		}
		bus.wg.Wait()
	}
}

func init() {
	defaultEventBus = newEventBus()
}
