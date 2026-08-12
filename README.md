# metricsx

家族指标基座核心：定义统一指标协议（Sink）与内置内存后端，
核心零第三方依赖；Prometheus 后端在 `metricsx/prometheus` 子模块。

> 当前状态：**v1.6.3**（Prometheus 后端位于 `metricsx/prometheus` 子模块）。

## 快速上手

### 默认内存后端（零依赖）

```go
m, err := metricsx.New()
m.IncCounter("orders.created", "region", "cn")
m.ObserveDuration("http.request", 0.12, "method", "GET")
snap, ok := m.Snapshot() // 测试与调试
```

### Prometheus 后端（子模块）

```go
import (
	metricsx "github.com/lcylpzls/metricsx"
	prometheusx "github.com/lcylpzls/metricsx/prometheus"
)

m, err := metricsx.New(prometheusx.WithPrometheus(
	prometheusx.WithNamespace("myapp"),
	prometheusx.WithRegistry(reg), // 默认 prometheus.DefaultRegisterer
))
```

指标自动以 `myapp_dbx_queries_total` 等命名暴露到 Prometheus /metrics，
`prometheusx.Gather(m)` 可抓取快照。

## 架构

- 核心 `metricsx`：`Sink` 协议 + `WithSink` 注入 + 内存后端，仅依赖 errx；
- 子模块 `metricsx/prometheus`：`WithPrometheus` / `WithNamespace` /
  `WithRegistry` / `WithBuckets` / `Gather` / `Registry`；
- 家族接入：各库经 `WithMetrics(m)` / `Config.Metrics` /
  `errx.SetMetricsHook` 直接注入，适配桥在项目层内联。

## 性能

- 内存后端热路径零分配；Prometheus 后端指标缓存 sync.Map 无锁读；
- Prometheus 文本端点可直接用 `metricsx/prometheus.HTTPHandler(m)`
  挂载到任意 mux / webx 路由；
- 详见 [docs/performance.md](docs/performance.md)。

## 文档与治理

- [docs/README.md](docs/README.md) — 文档索引
- [prometheus](prometheus) — Prometheus 后端子模块
- [examples/basic](examples/basic) — 接入示例
- [CONTRIBUTING.md](CONTRIBUTING.md) — 开发流程
- [SECURITY.md](SECURITY.md) — 安全说明

## 版本说明

- v1.4.0 为破坏性重构：`WithNamespace` / `WithRegistry` /
  `WithBuckets` / `Gather` 迁移至 `metricsx/prometheus`；
- 本库遵循语义化版本；当前按维护者授权进行未投产破坏性演进。
