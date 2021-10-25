package eventbus

import (
	"context"
	"sync"
	"sync/atomic"

	"github.com/wosai/ultron/v2"
	"github.com/wosai/ultron/v2/pkg/statistics"
)

type (
	// IEventBus 对eventbus的实现
	IEventBus struct {
		cancel                context.CancelFunc
		reportBus             chan statistics.SummaryReport
		resultBuses           []chan statistics.AttackResult
		otherBus              chan ultron.Event
		reportHandlers        []ultron.ReportHandleFunc
		resultHandlers        []ultron.ResultHandleFunc
		eventHandlers         map[ultron.EventType][]ultron.EventHandleFunc
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
	_ ultron.EventBus  = (*IEventBus)(nil)
	_ ultron.ReportBus = (*IEventBus)(nil)
	_ ultron.ResultBus = (*IEventBus)(nil)
)

func newEventBus() *IEventBus {
	bus := &IEventBus{
		reportBus:           make(chan statistics.SummaryReport, 3), // 低频通道
		otherBus:            make(chan ultron.Event, 3),             // 低频通道
		reportHandlers:      make([]ultron.ReportHandleFunc, 0),
		resultHandlers:      make([]ultron.ResultHandleFunc, 0),
		eventHandlers:       make(map[ultron.EventType][]ultron.EventHandleFunc),
		numberOfSubchannels: 20,
	}
	bus.resultBuses = make([]chan statistics.AttackResult, bus.numberOfSubchannels)
	for i := 0; i < int(bus.numberOfSubchannels); i++ {
		bus.resultBuses[i] = make(chan statistics.AttackResult, 100)
	}
	return bus
}

func (bus *IEventBus) SubscribeReport(fn ultron.ReportHandleFunc) {
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

func (bus *IEventBus) SubscribeResult(fn ultron.ResultHandleFunc) {
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

func (bus *IEventBus) Subscribe(et ultron.EventType, fn ultron.EventHandleFunc) {
	if fn == nil {
		return
	}
	if _, exists := bus.eventHandlers[et]; !exists {
		bus.eventHandlers[et] = []ultron.EventHandleFunc{fn}
		return
	}
	bus.eventHandlers[et] = append(bus.eventHandlers[et], fn)
}

func (bus *IEventBus) Publish(e ultron.Event) {
	if atomic.LoadUint32(&bus.closed) == 0 {
		bus.otherBus <- e
	}
}

func (bus *IEventBus) Start() {
	bus.once.Do(func() {
		ctx, cancel := context.WithCancel(context.Background())
		bus.cancel = cancel

		bus.wg.Add(1)
		go func() {
			defer bus.wg.Done()
			for report := range bus.reportBus {
				for _, fn := range bus.reportHandlers {
					fn(ctx, report)
				}
			}
		}()

		bus.wg.Add(1)
		go func() {
			defer bus.wg.Done()
			for event := range bus.otherBus {
				if handlers, ok := bus.eventHandlers[event.Type()]; ok {
					for _, handler := range handlers {
						handler(ctx, event)
					}
				}
			}
		}()

		for _, sub := range bus.resultBuses {
			bus.wg.Add(1)
			go func(c <-chan statistics.AttackResult) {
				defer bus.wg.Done()
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
		bus.wg.Wait()
		close(bus.reportBus)
		close(bus.otherBus)
		for _, sub := range bus.resultBuses {
			close(sub)
		}
	}
}

func init() {
	DefaultEventBus = newEventBus()
}
