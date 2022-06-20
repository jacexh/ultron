# ChangeLog

## v2.4.5

- Convert the type of `ultron_attacker_failures_total` and `ultron_attacker_requests_total` to [Counter](https://prometheus.io/docs/concepts/metric_types/#counter)

## v2.4.2

- now you can change name, level, etc. of log file via an external config file: `./config.yml`
## v2.4.1

- make `Influxdbv1Handler` as independent module, path: `github.com/wosai/ultron/handler/influxdbv1/v2`
## v2.4.0

- add `executorSharedContext` to carry fire-scoped values
- change method signature of `HTTPPrepareFunc`
- change method signature of `HTTPCheckFunc`

## v2.3.2

- `fastattacker`作为独立的module，路径为`github.com/wosai/ultron/attacker/fastattacker/v2`
- 添加`jsonrpc`的Attacker实现，module path: `github.com/wosai/ultron/attacker/jsonrpc/v2`
- change method signature of `NewHTTPAttacker`

## v2.2.1

- 修复在Slave停止Plan时当处于RampUp阶段会出现的数据竞争问题

## v2.2.0

- 修改InfluxDBV1Handler的接口设计
- 修改统计对象中的部分字段名称
- 修复前端各类安全漏洞
- 前端界面处于稳定可用状态
- 忽略在阶段切换过程中、进程被终止时产生的`context.Canceled`错误
- 调整`AttackerStrategy`接口定义

## v2.1.2

- 修复了写入influxdb数据异常的问题
## v2.1.1

- 修复写入InfluxDB时锁未释放的问题
- 后端以json格式暴露prometheus metric
- 修复无法正确生成加压策略的问题
- 前端优化一堆问题

## v2.1.0

- 添加Web Portal
- 接入Prometheus

## v2.0.2

- 放弃Nested Module

## v2.0.0

- 重写了整个项目，对各种可扩展对象进行了抽象
- 2.0中有且仅有分布式执行模式
- 新增`Plan`对象，这样来2.0就可以支持多`Plan`的串行执行
- 优化了终端下报告的展示效果

## v1.6.0

- 项目地址从`github.com/qastub/ultron` 迁移到 `github.com/wosai/ultron`