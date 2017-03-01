package main

import (
	"net/http"
	"time"

	"github.com/jacexh/ultron"
)

func main() {
	baidu := ultron.NewHTTPRequest("GET: baidu")
	baidu.Prepare = func() *http.Request {
		req, _ := http.NewRequest(http.MethodGet, "http://192.168.1.33/benchmark", nil)
		return req
	}
	index := ultron.NewHTTPRequest("INDEX")
	index.Prepare = func() *http.Request {
		req, _ := http.NewRequest(http.MethodGet, "http://192.168.1.33/", nil)
		return req
	}

	taskSet := ultron.NewTaskSet()
	taskSet.MinWait = ultron.ZeroDuration
	taskSet.MaxWait = ultron.ZeroDuration
	taskSet.Add(baidu, 2)
	taskSet.Add(index, 1)
	ultron.CoreRunner.WithTaskSet(taskSet).SetDuration(time.Second * 20).Run()
}
