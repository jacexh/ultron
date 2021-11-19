# ChangeLog

## v2.2.0

- 修改InfluxDBV1Handler的接口设计
- 修改统计对象中的部分字段名称
## v2.1.2

- 修复了写入influxdb数据异常的问题
## v2.1.1

- 修复写入InflxuDB时锁未释放的问题
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