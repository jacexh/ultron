package ultron

import (
	"fmt"
	"sort"
	"strings"

	"go.uber.org/zap"
)

type (
	// ResultHandleFunc .
	ResultHandleFunc func(*Result)

	// ReportHandleFunc .
	ReportHandleFunc func(Report)

	eventHook struct {
		retFuncs    []ResultHandleFunc
		repFuncs    []ReportHandleFunc
		Concurrency int // 控制ResultHandleFunc执行时的最大并发数，避免造成竞争
		ch          chan struct{}
	}
)

var (
	// LocalEventHook 单机运行时使用的钩子
	LocalEventHook *eventHook
	// SlaveEventHook 分布式执行时，节点钩子
	SlaveEventHook *eventHook
	// MasterEventHook 分布式执行时，主控钩子
	MasterEventHook *eventHook

	// LocalEventHookConcurrency .
	LocalEventHookConcurrency = 200
	// SlaveEventHookConcurrency .
	SlaveEventHookConcurrency = 200
	// MasterEventHookConcurrency 默认不用控制
	MasterEventHookConcurrency = 0

	cutLine = strings.Repeat("-", 126)
)

func newEventHook(c int) *eventHook {
	return &eventHook{
		retFuncs:    []ResultHandleFunc{},
		repFuncs:    []ReportHandleFunc{},
		Concurrency: c,
	}
}

func (eh *eventHook) AddResultHandleFunc(retFunc ...ResultHandleFunc) {
	for _, f := range retFunc {
		eh.retFuncs = append(eh.retFuncs, f)
	}
}

func (eh *eventHook) AddReportHandleFunc(repFunc ...ReportHandleFunc) {
	for _, f := range repFunc {
		eh.repFuncs = append(eh.repFuncs, f)
	}
}

func (eh *eventHook) listen(retC resultPipeline, repC reportPipeline) {
	if eh.Concurrency > 0 {
		eh.ch = make(chan struct{}, eh.Concurrency)
	}

	// for {
	// 	select {
	// 	case rep := <-repC: // 频次低，不用阻塞
	// 		go func(r Report) {
	// 			for _, f := range eh.repFuncs {
	// 				f(r)
	// 			}
	// 		}(rep)

	// 	case ret := <-retC:
	// 		eh.ch <- struct{}{}
	// 		go func(r *Result) {
	// 			defer func() { <-eh.ch }()
	// 			for _, f := range eh.retFuncs {
	// 				f(r)
	// 			}
	// 		}(ret)

	// 	}
	// }
	if retC != nil {
		go func(c resultPipeline) {
			for r := range c {
				if eh.Concurrency > 0 {
					eh.ch <- struct{}{}
				}
				go func(r *Result) {
					defer func() {
						if eh.Concurrency > 0 {
							<-eh.ch
						}
					}()
					for _, f := range eh.retFuncs {
						f(r)
					}
				}(r)
			}
		}(retC)
	}

	if repC != nil {
		go func(c reportPipeline) {
			for rep := range c {
				go func(r Report) {
					for _, f := range eh.repFuncs {
						f(r)
					}
				}(rep)
			}
		}(repC)
	}
}

func printReportToConsole(report Report) {
	var full bool
	var keys []string

	for k, r := range report {
		keys = append(keys, k)
		if r.FullHistory {
			full = r.FullHistory
		}
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

	if full {
		data, err := json.MarshalIndent(report, "", "  ")
		if err == nil {
			fmt.Printf("============= Summary Report =============\n\n" + string(data) + "\n")
		} else {
			Logger.Error("marshel report object failed", zap.Error(err))
		}
	}
}
