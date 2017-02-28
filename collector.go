package ultron

import (
	"sync"
	"time"

	"go.uber.org/zap"
)

type statsCollector struct {
	entries  map[string]*statsEntry
	receiver chan *QueryResult
	lock     sync.RWMutex
}

func newStatsCollector() *statsCollector {
	return &statsCollector{
		entries:  map[string]*statsEntry{},
		receiver: make(chan *QueryResult),
	}
}

func (c *statsCollector) logSuccess(name string, t time.Duration) {
	if _, ok := c.entries[name]; !ok {
		c.lock.Lock()
		defer c.lock.Unlock()
		c.entries[name] = newStatsEntry(name)
	}
	c.entries[name].logSuccess(t)
}

func (c *statsCollector) logFailure(name string, err error) {
	if _, ok := c.entries[name]; !ok {
		c.lock.Lock()
		defer c.lock.Unlock()
		c.entries[name] = newStatsEntry(name)
	}
	Logger.Warn("occure error", zap.String("error", err.Error()))
	c.entries[name].logFailure(err)
}

// Receiving 主函数，监听channel进行统计
func (c *statsCollector) Receiving() {
	for r := range c.receiver {
		// Todo: ctx
		if r.Error == nil {
			go c.logSuccess(r.Name, r.Duration)
		} else {
			go c.logFailure(r.Name, r.Error)
		}
	}
}

// Receiver 返回接收事务结果通道
func (c *statsCollector) Receiver() chan<- *QueryResult {
	return c.receiver
}
