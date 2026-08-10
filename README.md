# metricsx

自研 Prometheus 指标适配层:一个实例喂饱底座全部库的统一
Metrics 接口,统一命名、统一标签、统一分桶。

> 当前状态:**v1.3.0 正式版,API 已冻结**。

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

瞬时量（活跃请求、连接数等）使用 `AddGauge` / `SetGauge`，字节等
批量累计使用 `AddCounter`，自动以 `myapp_webx_inflight` 等命名暴露。

## 性能

- IncCounter 命中 ~42ns、ObserveDuration ~43ns,热路径 0 分配;
- 指标缓存 sync.Map 无锁读,并发懒创建安全;
- 详见 [docs/performance.md](docs/performance.md)。

## 文档与治理

- [docs/README.md](docs/README.md) — 文档索引
- [docs/operations.md](docs/operations.md) — 运行手册
- [adapters](adapters) — 家族官方适配层（filex/jobx/updatex/idgenx/errx）
- [examples/basic](examples/basic) — 接入示例
- [CONTRIBUTING.md](CONTRIBUTING.md) — 开发流程
- [SECURITY.md](SECURITY.md) — 安全说明

## 稳定性承诺

- 本库遵循[语义化版本](https://semver.org/lang/zh-CN/);
- v1.0.0 起公开 API 冻结:新增能力以次版本发布,
  破坏性变更仅随主版本;
- 每个版本发布前执行:100% 覆盖率、race、staticcheck、fuzz、
  govulncheck、apidiff 对比与三平台 CI。

## 定位

metricsx 不是指标系统,不实现 Prometheus 协议;它解决每个项目
接 Prometheus 都要重复的部分:

- 实现 dbx / httpx / webx / cachex / resiliencex / jobx 等
  统一形态的 Metrics 接口(IncCounter / ObserveDuration);
- 扩展 Gauge 与增量计数(AddGauge / SetGauge / AddCounter),
  覆盖活跃水位、连接数与字节量场景;
- 官方适配层 metricsx/adapters：桥接 filex/jobx/updatex/idgenx/errx；
- 懒创建 CounterVec / HistogramVec,统一 namespace 与命名规范;
- 预注册声明帮助文本与标签键名,未注册自动占位。

## 文档

- [docs/README.md](docs/README.md) — 文档索引
- [docs/metrics-research.md](docs/metrics-research.md) — 指标适配调研手册

## License

MIT © [lcylpzls](https://github.com/lcylpzls)
