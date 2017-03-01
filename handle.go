package ultron

import (
	"encoding/json"
	"fmt"
	"sync"
	"time"

	"go.uber.org/zap"
)

type (
	// HandleResult type of result handle function
	HandleResult func(*QueryResult)
	// HandleReport type of report handle function
	HandleReport func(map[string]*StatsReport)

	resultHandleChain struct {
		handles []HandleResult
		ch      chan *QueryResult
		wg      *sync.WaitGroup
	}

	reportHandleChain struct {
		handles []HandleReport
		ch      chan map[string]*StatsReport
		wg      *sync.WaitGroup
	}
)

const (
	// StatsReportInterval time interval
	StatsReportInterval = time.Second * 5
)

var (
	// ResultHandleChain handlers for each request
	ResultHandleChain *resultHandleChain
	// ReportHandleChain handlers for each timing stats report
	ReportHandleChain *reportHandleChain
)

func (rc *resultHandleChain) AddHandle(fn HandleResult) {
	rc.handles = append(rc.handles, fn)
}

func (rc *resultHandleChain) channel() chan *QueryResult {
	return rc.ch
}

func (rc *resultHandleChain) listening() {
	for msg := range rc.ch {
		rc.wg.Add(1)
		go func(ret *QueryResult) {
			defer rc.wg.Done()
			for _, f := range rc.handles {
				f(ret)
			}
		}(msg)
	}
}

func (rc *resultHandleChain) safeClose() {
	rc.wg.Wait()
	close(rc.ch)
}

func (re *reportHandleChain) AddHandle(fn HandleReport) {
	re.handles = append(re.handles, fn)
}

func (re *reportHandleChain) channel() chan map[string]*StatsReport {
	return re.ch
}

func (re *reportHandleChain) listening() {
	for msg := range re.ch {
		re.wg.Add(1)
		go func(s map[string]*StatsReport) {
			defer re.wg.Done()
			for _, h := range re.handles {
				h(s)
			}
		}(msg)
	}
}

func (re *reportHandleChain) safeClose() {
	re.wg.Wait()
	close(re.ch)
}

func printReportToConsole(report map[string]*StatsReport) {
	start := time.Now()
	data, err := json.Marshal(report)
	if err == nil {
		fmt.Println(string(data))
		Logger.Info("print", zap.Duration("elasped", time.Since(start)))
		return
	}
}

func init() {
	ResultHandleChain = &resultHandleChain{
		handles: []HandleResult{defaultStatsCollector.log},
		ch:      make(chan *QueryResult),
		wg:      &sync.WaitGroup{},
	}
	ReportHandleChain = &reportHandleChain{
		handles: []HandleReport{printReportToConsole},
		ch:      make(chan map[string]*StatsReport),
		wg:      &sync.WaitGroup{},
	}
}
