package eventbus

import (
	"context"
	"runtime/debug"
	"sync"
	"sync/atomic"

	"github.com/wosai/ultron/v2/log"
	"github.com/wosai/ultron/v2/pkg/statistics"
	"go.uber.org/zap"
)

type (
	// IEventBus 对eventbus的实现
	IEventBus struct {
		cancel                context.CancelFunc
		reportBus             chan statistics.SummaryReport
		resultBuses           []chan statistics.AttackResult
		reportHandlers        []statistics.ReportHandleFunc
		resultHandlers        []statistics.ResultHandleFunc
		numberOfSubchannels   uint32
		counterForSubchannels uint32
		closed                uint32
		once                  sync.Once
		wg                    sync.WaitGroup
	}
)

var (
	DefaultEventBus *IEventBus
)

var (
	_ statistics.ReportBus = (*IEventBus)(nil)
	_ statistics.ResultBus = (*IEventBus)(nil)
)

func newEventBus() *IEventBus {
	bus := &IEventBus{
		reportBus:           make(chan statistics.SummaryReport, 3), // 低频通道
		reportHandlers:      make([]statistics.ReportHandleFunc, 0),
		resultHandlers:      make([]statistics.ResultHandleFunc, 0),
		numberOfSubchannels: 25,
	}
	bus.resultBuses = make([]chan statistics.AttackResult, bus.numberOfSubchannels)
	for i := 0; i < int(bus.numberOfSubchannels); i++ {
		bus.resultBuses[i] = make(chan statistics.AttackResult, 200)
	}
	return bus
}

func (bus *IEventBus) SubscribeReport(fn statistics.ReportHandleFunc) {
	if fn == nil {
		return
	}
	bus.reportHandlers = append(bus.reportHandlers, fn)
}

func (bus *IEventBus) PublishReport(report statistics.SummaryReport) {
	if atomic.LoadUint32(&bus.closed) == 0 {
		bus.reportBus <- report
	}
}

func (bus *IEventBus) SubscribeResult(fn statistics.ResultHandleFunc) {
	if fn == nil {
		return
	}
	bus.resultHandlers = append(bus.resultHandlers, fn)
}

func (bus *IEventBus) PublishResult(ret statistics.AttackResult) {
	if atomic.LoadUint32(&bus.closed) == 0 {
		v := atomic.AddUint32(&bus.counterForSubchannels, 1)
		bus.resultBuses[int((v-1)%bus.numberOfSubchannels)] <- ret
	}
}

func (bus *IEventBus) Start() {
	bus.once.Do(func() {
		ctx, cancel := context.WithCancel(context.Background())
		bus.cancel = cancel

		bus.wg.Add(1)
		go func() {
			defer func() {
				if rec := recover(); rec != nil {
					debug.PrintStack()
					log.DPanic("report bus is quit", zap.Any("recover", rec))
				}
				bus.wg.Done()
				bus.Close()
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
						log.DPanic("one result bus is quit", zap.Any("recover", rec))
					}
					bus.wg.Done()
					bus.Close()
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

func (bus *IEventBus) Close() {
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
	DefaultEventBus = newEventBus()
}
