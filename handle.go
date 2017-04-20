package ultron

import (
	"encoding/json"
	"fmt"
	"sync"
	"time"
)

type (
	// HandleResult type of result handle function
	HandleResult func(*AttackResult)
	// HandleReport type of report handle function
	HandleReport func(map[string]*StatsReport)

	resultHandleChain struct {
		handles []HandleResult
		ch      chan *AttackResult
		wg      sync.WaitGroup
	}

	reportHandleChain struct {
		handles []HandleReport
		ch      chan map[string]*StatsReport
		wg      sync.WaitGroup
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

func (rc *resultHandleChain) channel() chan *AttackResult {
	return rc.ch
}

func (rc *resultHandleChain) listening() {
	for msg := range rc.ch {
		rc.wg.Add(1)
		go func(ret *AttackResult) {
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
	var full bool
	for _, r := range report {
		if r.FullHistory {
			full = true
		}
		break
	}

	if !full {
		s := fmt.Sprintf("|%-24s|%6s|%10s|%10s|%8s|%8s|%8s|%8s|%8s|%8s|%8s|%8s|%8s|\n", "Name", "QPS", "Requests", "Failures", "Min", "Max", "Avg", "Median", "70%", "80%", "90%", "95%", "99%")
		for _, r := range report {
			s += fmt.Sprintf("|%-24s|%6d|%10d|%10d|%8d|%8d|%8d|%8d|%8d|%8d|%8d|%8d|%8d|\n", r.Name, r.QPS, r.Requests, r.Failures, r.Min, r.Max, r.Average, r.Median, r.Distributions["0.70"], r.Distributions["0.80"], r.Distributions["0.90"], r.Distributions["0.95"], r.Distributions["0.99"])
		}
		fmt.Println(s)
	} else {
		data, err := json.Marshal(report)
		if err == nil {
			fmt.Println(string(data))
		}
	}
}

func init() {
	ResultHandleChain = &resultHandleChain{
		handles: []HandleResult{defaultStatsCollector.log},
		ch:      make(chan *AttackResult),
	}
	ReportHandleChain = &reportHandleChain{
		handles: []HandleReport{printReportToConsole},
		ch:      make(chan map[string]*StatsReport),
	}
}
