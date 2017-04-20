package main

import (
	"net/http"
	"time"

	"github.com/jacexh/ultron"
)

func main() {
	baidu := ultron.NewHTTPAttacker("GET: baidu")
	baidu.Prepare = func() *http.Request {
		req, _ := http.NewRequest(http.MethodGet, "http://192.168.1.33/benchmark", nil)
		return req
	}
	index := ultron.NewHTTPAttacker("INDEX")
	index.Prepare = func() *http.Request {
		req, _ := http.NewRequest(http.MethodGet, "http://192.168.1.33/", nil)
		return req
	}

	taskSet := ultron.NewTaskSet()
	taskSet.MinWait = ultron.ZeroDuration
	taskSet.MaxWait = ultron.ZeroDuration
	taskSet.Add(baidu, 2)
	taskSet.Add(index, 1)
	ultron.Runner.Config = &ultron.RunnerConfig{
		Concurrence: 100,
		Duration:    time.Minute * 1,
		HatchRate:   10,
	}

	ultron.Runner.Run(taskSet)
}
