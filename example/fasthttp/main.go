package main

import (
	"github.com/jacexh/ultron"
	"github.com/valyala/fasthttp"
)

func main() {
	benchmark := ultron.NewFastHTTPAttacker("fasthttp-benchmark")
	benchmark.Prepare = func() *fasthttp.Request {
		req := fasthttp.AcquireRequest()
		req.SetRequestURI("http://192.168.1.30")
		return req
	}

	task := ultron.NewTaskSet()
	task.MinWait = ultron.ZeroDuration
	task.MaxWait = ultron.ZeroDuration
	task.Add(benchmark, 1)

	ultron.Runner.Config.Concurrence = 200
	ultron.Runner.Config.Requests = 100000
	ultron.Runner.Run(task)
}
