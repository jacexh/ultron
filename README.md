# Ultron
a http load testing tool in go

[![Build Status](https://travis-ci.org/jacexh/ultron.svg?branch=master)](https://travis-ci.org/jacexh/ultron)

## Example:

### **Script**

file path: `example/fasthttp/main.go`

```go
benchmark := ultron.NewFastHTTPRequest("fasthttp-benchmark")
benchmark.Prepare = func() *fasthttp.Request {
		req := fasthttp.AcquireRequest()
		req.SetRequestURI("http://192.168.1.33/benchmark")
		return req
}

task := ultron.NewTaskSet()
task.MinWait = ultron.ZeroDuration
task.MaxWait = ultron.ZeroDuration
task.Add(benchmark, 1)

ultron.CoreRunner.WithTaskSet(task).SetConcurrence(200).SetHatchRate(30).Run()
```

### Report

```json
{
  "fasthttp-benchmark": {
    "name": "fasthttp-benchmark",
    "requests": 1917994,
    "failures": 0,
    "min": 0,
    "max": 23,
    "median": 2,
    "average": 2,
    "qps": 50211,
    "distributions": {
      "0.50": 2,
      "0.60": 2,
      "0.70": 2,
      "0.80": 2,
      "0.90": 2,
      "0.95": 2,
      "0.97": 2,
      "0.98": 3,
      "0.99": 4,
      "1.00": 23
    },
    "failure_details": {},
    "full_history": false
  }
}
```
