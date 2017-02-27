package main

import (
	"net/http"

	"github.com/jacexh/ultron"
)

func main() {
	baidu := ultron.NewHTTPRequest("GET: baidu")
	baidu.Prepare = func() *http.Request {
		req, _ := http.NewRequest(http.MethodGet, "http://www.baidu.com/", nil)
		return req
	}

	taskSet := ultron.NewTaskSet()
	taskSet.Concurrency = 100
	taskSet.MinWait = ultron.ZeroDuration
	taskSet.MaxWait = ultron.ZeroDuration
	taskSet.Add(baidu, 1)
	ultron.CoreRunner.WithTaskSet(taskSet).Run()
}
