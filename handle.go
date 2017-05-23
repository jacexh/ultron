package ultron

import (
	"encoding/json"
	"fmt"
	"sort"
	"strings"
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

	cutLine = strings.Repeat("-", 126)
)

func (rc *resultHandleChain) AddHandles(fn ...HandleResult) {
	rc.handles = append(rc.handles, fn...)
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

func (re *reportHandleChain) AddHandles(fn ...HandleReport) {
	re.handles = append(re.handles, fn...)
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
		var keys []string
		for k := range report {
			keys = append(keys, k)
		}
		sort.Strings(keys)

		s := fmt.Sprintf("|%-48s|%12s|%12s|%12s|%8s|%9s|%8s|%8s|\n", "Name", "Requests", "Failures", "QPS", "Min", "Max", "Avg", "Median")
		d := fmt.Sprintf("\nPercentage of the requests completed within given times: \n\n|%-48s|%12s|%8s|%8s|%8s|%8s|%8s|%8s|%8s|\n", "Name", "Requests", "60%", "70%", "80%", "90%", "95%", "98%", "99%")
		for _, key := range keys {
			r := report[key]
			s += fmt.Sprintf("|%-48s|%12d|%12d|%12d|%8d|%9d|%8d|%8d|\n", r.Name, r.Requests, r.Failures, r.QPS, r.Min, r.Max, r.Average, r.Median)
			d += fmt.Sprintf("|%-48s|%12d|%8d|%8d|%8d|%8d|%8d|%8d|%8d|\n", r.Name, r.Requests, r.Distributions["0.60"], r.Distributions["0.70"], r.Distributions["0.80"], r.Distributions["0.90"], r.Distributions["0.95"], r.Distributions["0.98"], r.Distributions["0.99"])
		}
		fmt.Println(cutLine + "\n" + s + d + cutLine + "\n")
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
