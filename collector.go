package ultron

import (
	"sync"
	"time"

	"go.uber.org/zap"
)

type statsCollector struct {
	entries map[string]*statsEntry
	lock    sync.RWMutex
}

var defaultStatsCollector *statsCollector

func newStatsCollector() *statsCollector {
	return &statsCollector{
		entries: map[string]*statsEntry{},
	}
}

func (c *statsCollector) createEntries(n ...string) {
	c.lock.Lock()
	defer c.lock.Unlock()

	for _, i := range n {
		c.entries[i] = newStatsEntry(i)
	}
}

func (c *statsCollector) logSuccess(name string, t time.Duration) {
	c.entries[name].logSuccess(t)
}

func (c *statsCollector) logFailure(name string, err error) {
	Logger.Warn("occure error", zap.String("error", err.Error()))
	c.entries[name].logFailure(err)
}

// Receiving 主函数，监听channel进行统计
func (c *statsCollector) log(ret *RequestResult) {
	if ret.Error == nil {
		c.logSuccess(ret.Name, ret.Duration)
	} else {
		c.logFailure(ret.Name, ret.Error)
	}
}

func (c *statsCollector) report(full bool) map[string]*StatsReport {
	r := map[string]*StatsReport{}
	c.lock.RLock()
	defer c.lock.RUnlock()
	for k, v := range c.entries {
		r[k] = v.report(full)
	}
	return r
}

func init() {
	defaultStatsCollector = newStatsCollector()
}
