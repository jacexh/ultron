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

	ultron.CoreRunner.WithTaskSet(task).SetConcurrence(200).SetHatchRate(30).Run()
}
