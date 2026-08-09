# metricsx

自研 Prometheus 指标适配层:一个实例喂饱底座全部库的统一
Metrics 接口,统一命名、统一标签、统一分桶。

> 当前状态:**v0.5.0 实现完成,待 CI 验证与发布**。

## 快速上手

```go
m, err := metricsx.New(metricsx.WithNamespace("myapp"))
if err != nil {
	panic(err)
}

// 喂饱全部库
dbx.OpenConfig(ctx, dbx.Config{DSN: dsn, Metrics: m})
httpx.New(httpx.WithMetrics(m))
cachex.New(cachex.WithMetrics(m))

// 预注册声明标签键名(可选)
m.Register("dbx.queries", "数据库查询次数", "op")
```

指标自动以 `myapp_dbx_queries_total` 等命名暴露到 Prometheus /metrics。
`m.Gather()` 可抓取快照用于测试与调试。

## 性能

- IncCounter 命中 ~42ns、ObserveDuration ~43ns,热路径 0 分配;
- 指标缓存 sync.Map 无锁读,并发懒创建安全;
- 详见 [docs/performance.md](docs/performance.md)。

## 文档与治理

- [docs/README.md](docs/README.md) — 文档索引
- [docs/operations.md](docs/operations.md) — 运行手册
- [examples/basic](examples/basic) — 接入示例
- [CONTRIBUTING.md](CONTRIBUTING.md) — 开发流程
- [SECURITY.md](SECURITY.md) — 安全说明

## 定位

metricsx 不是指标系统,不实现 Prometheus 协议;它解决每个项目
接 Prometheus 都要重复的部分:

- 实现 dbx / httpx / webx / cachex / resiliencex / jobx 等
  统一形态的 Metrics 接口(IncCounter / ObserveDuration);
- 懒创建 CounterVec / HistogramVec,统一 namespace 与命名规范;
- 预注册声明帮助文本与标签键名,未注册自动占位。

## 文档

- [docs/README.md](docs/README.md) — 文档索引
- [docs/metrics-research.md](docs/metrics-research.md) — 指标适配调研手册

## License

MIT © [lcylpzls](https://github.com/lcylpzls)
