package ultron

import (
	"encoding/json"
	"fmt"
	"sync/atomic"
	"time"
)

type (
	// HandleResult type of result handle function
	HandleResult func(*QueryResult)
	// HandleReport type of report handle function
	HandleReport func(map[string]*StatsReport)

	resultHandleChain struct {
		handles []HandleResult
		ch      chan *QueryResult
	}

	reportHandleChain struct {
		handles  []HandleReport
		ch       chan map[string]*StatsReport
		children int64
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
		go func(ret *QueryResult) {
			for _, f := range rc.handles {
				f(ret)
			}
		}(msg)
	}
}

func (re *reportHandleChain) AddHandle(fn HandleReport) {
	re.handles = append(re.handles, fn)
}

func (re *reportHandleChain) channel() chan map[string]*StatsReport {
	return re.ch
}

func (re *reportHandleChain) listening() {
	for msg := range re.ch {
		atomic.AddInt64(&re.children, 1)
		go func(s map[string]*StatsReport) {
			defer func() { atomic.AddInt64(&re.children, -1) }()
			for _, h := range re.handles {
				h(s)
			}
		}(msg)
	}
}

func (re *reportHandleChain) busy() bool {
	if atomic.LoadInt64(&re.children) <= 0 {
		return false
	}
	return true
}

func printReportToConsole(report map[string]*StatsReport) {
	data, err := json.Marshal(report)
	if err == nil {
		fmt.Println(string(data))
	}
}

func init() {
	ResultHandleChain = &resultHandleChain{handles: []HandleResult{defaultStatsCollector.log}, ch: make(chan *QueryResult)}
	ReportHandleChain = &reportHandleChain{handles: []HandleReport{printReportToConsole}, ch: make(chan map[string]*StatsReport)}
}
