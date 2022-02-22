# Ultron

[![Go Report Card](https://goreportcard.com/badge/github.com/wosai/ultron)](https://goreportcard.com/report/github.com/wosai/ultron) 
[![codecov](https://codecov.io/gh/wosai/ultron/branch/master/graph/badge.svg)](https://codecov.io/gh/wosai/ultron) 
[![GoDoc](https://godoc.org/github.com/wosai/ultron?status.svg)](https://godoc.org/github.com/wosai/ultron)
[![Ultron CI](https://github.com/WoSai/ultron/actions/workflows/ci.yml/badge.svg)](https://github.com/WoSai/ultron/actions/workflows/ci.yml)
[![CodeQL](https://github.com/WoSai/ultron/actions/workflows/codeql-analysis.yml/badge.svg)](https://github.com/WoSai/ultron/actions/workflows/codeql-analysis.yml)

a http load testing tool in go

## Usage 
### Requirements

Go 1.16+

### Install

```bash
go get github.com/wosai/ultron/v2
```

### Example

#### LocalRunner

```go
package main

import (
	"net/http"
	"time"

	"github.com/wosai/ultron/v2"
)

func main() {
	task := ultron.NewTask()
	attacker := ultron.NewHTTPAttacker("google")
	attacker.Apply(
		ultron.WithPrepareFunc(func() (*http.Request, error) { // 压测事务逻辑实现
			return http.NewRequest(http.MethodGet, "https://www.google.com/ncr", nil)
		}),
		ultron.WithCheckFuncs(ultron.CheckHTTPStatusCode),
	)
	task.Add(attacker, 1)

	plan := ultron.NewPlan("google homepage")
	plan.AddStages(
		&ultron.V1StageConfig{
			Duration:        10 * time.Minute,
			ConcurrentUsers: 200,
			RampUpPeriod:    10,
		}
	)

	runner := ultron.NewLocalRunner()
	runner.Assign(task)   

	if err := runner.Launch(); err != nil {
		panic(err)
	}

	if err := runner.StartPlan(plan); err != nil {
		panic(err)
	}

	block := make(chan struct{}, 1)
	<-block
}
```

#### SlaveRunner

```go
package main

import (
	"net/http"

	"github.com/wosai/ultron/v2"
	"google.golang.org/grpc"
)

func main() {
	task := ultron.NewTask()
	attacker := ultron.NewHTTPAttacker("google")
	attacker.Apply(
		ultron.WithPrepareFunc(func() (*http.Request, error) { // 压测事务逻辑实现
			return http.NewRequest(http.MethodGet, "https://www.google.com/ncr", nil)
		}),
		ultron.WithCheckFuncs(ultron.CheckHTTPStatusCode),
	)
	task.Add(attacker, 1)

	// 启动runner
	runner := ultron.NewSlaveRunner()
	runner.Assign(task)
	runner.SubscribeResult(nil)                                          // 订阅单次压测结果
	if err := runner.Connect(":2021", grpc.WithInsecure()); err != nil { // 连接master的grpc服务
		panic(err)
	}

	// 阻塞当前goroutine，避免程序推出
	block := make(chan struct{}, 1)
	<-block
}
```

#### MasterRunner

```bash
ultron
```

![master](https://my-storage.oss-cn-shanghai.aliyuncs.com/picgo/20211102111633.png)

#### Web Portal

![plan](https://my-storage.oss-cn-shanghai.aliyuncs.com/picgo/20211118094334.png)

![stats](https://my-storage.oss-cn-shanghai.aliyuncs.com/picgo/20211118090312.png)

### Report

#### Terminal Table

![stats report](https://my-storage.oss-cn-shanghai.aliyuncs.com/picgo/20211102111021.png)

#### JSON Format
```json
{
    "first_attack": "2021-11-02T03:09:08.419359417Z",
    "last_attack": "2021-11-02T03:09:21.209236204Z",
    "total_requests": 139450,
    "total_tps": 10903.15429322517,
    "full_history": true,
    "reports": {
        "benchmark": {
            "name": "benchmark",
            "requests": 139450,
            "min": 10000253,
            "max": 40510983,
            "median": 10000000,
            "average": 11869156,
            "tps": 10903.15429322517,
            "distributions": {
                "0.50": 10000000,
                "0.60": 10000000,
                "0.70": 10000000,
                "0.80": 10000000,
                "0.90": 20000000,
                "0.95": 21000000,
                "0.97": 26000000,
                "0.98": 29000000,
                "0.99": 30000000,
                "1.00": 40510983
            },
            "full_history": true,
            "first_attack": "2021-11-02T03:09:08.419359417Z",
            "last_attack": "2021-11-02T03:09:21.209236204Z"
        }
    },
    "extras": {
        "plan": "benchmark test"
    }
}
```

#### Grafana Dashboard

```bash
scripts/grafana/dashboard.json
```

![](https://my-storage.oss-cn-shanghai.aliyuncs.com/picgo/20211119120144.png)

![](https://my-storage.oss-cn-shanghai.aliyuncs.com/picgo/20211119120154.png)

### Enhancements

Module | Type | Description 
:---: |  :----:  |  :---:
`github.com/wosai/ultron/attacker/fastattacker/v2` | Attacker | Another http attacker implemented by [fasthttp](https://github.com/valyala/fasthttp)
`github.com/wosai/ultron/attacker/jsonrpc/v2`  | Attacker | A attacker used for jsonrpc protocol
`github.com/wosai/ultron/handler/influxdbv1/v2` | Handler |  A handler that save attack result and report in InfluxDB v1

## Contributors
<a href="https://github.com/wosai/ultron/graphs/contributors">
  <img src="https://contrib.rocks/image?repo=wosai/ultron" />
</a>

Made with [contrib.rocks](https://contrib.rocks).
