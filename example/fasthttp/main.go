package main

import "github.com/jacexh/ultron"
import "github.com/valyala/fasthttp"

func main() {
	task := ultron.NewTaskSet()
	benchmark := ultron.NewFastHTTPRequest("fasthttp-benchmark")
	benchmark.Prepare = func() *fasthttp.Request {
		req := fasthttp.AcquireRequest()
		req.SetRequestURI("http://192.168.1.33/benchmark")
		return req
	}
	ultron.CoreRunner.WithTaskSet(task.Add(benchmark, 1)).Run()
}
