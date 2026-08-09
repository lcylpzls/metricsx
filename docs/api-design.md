# metricsx API 参考

> 状态:**v1.0.0 API 已冻结**。新增能力以次版本发布,
> 破坏性变更仅随主版本;任何修改须经 apidiff 对比并记录 CHANGELOG。

## 1. 快速上手

```go
m, err := metricsx.New(metricsx.WithNamespace("myapp"))
if err != nil {
	panic(err)
}

// 喂饱全部库
dbx.OpenConfig(ctx, dbx.Config{DSN: dsn, Metrics: m})
httpx.New(httpx.WithMetrics(m))
cachex.New(cachex.WithMetrics(m))
resiliencex.NewTokenBucket(100, 10, resiliencex.WithMetrics(m))

// 预注册声明标签键名(可选)
m.Register("dbx.queries", "数据库查询次数", "op")
```

## 2. 核心类型

```go
type Metrics struct { /* 未导出 */ }

func New(opts ...Option) (*Metrics, error)
func (m *Metrics) IncCounter(name string, labels ...string)
func (m *Metrics) ObserveDuration(name string, seconds float64, labels ...string)
func (m *Metrics) Register(name, help string, labelNames ...string) error
func (m *Metrics) Registry() prometheus.Registerer
func (m *Metrics) Gather() ([]*dto.MetricFamily, error) // v0.3.0
```

## 3. 配置

```go
func WithNamespace(ns string) Option         // 默认空
func WithRegistry(r prometheus.Registerer) Option // 默认 DefaultRegisterer
func WithBuckets(buckets []float64) Option  // Histogram 分桶
```

## 4. 语义

- `IncCounter("dbx.queries", "select")` →
  CounterVec `myapp_dbx_queries_total{label0="select"}`;
- `ObserveDuration("httpx.duration", 0.5, "GET")` →
  HistogramVec `myapp_httpx_duration_seconds{label0="GET"}`;
- `Register` 后可读标签:如 `Register("dbx.queries", "查询次数", "op")`;
- 重复注册返回 MTRX_ALREADY_REGISTERED。

## 5. 错误码

| 错误码 | 含义 |
| --- | --- |
| MTRX_INVALID_CONFIG | 配置非法 |
| MTRX_ALREADY_REGISTERED | 指标已注册 |

## 6. 迭代范围

| 版本 | 内容 |
| --- | --- |
| v0.1.0 | New / IncCounter / ObserveDuration / Register / Registry,懒创建 |
| v0.2.0 | 配置校验完善、集成断言(gather)、冲突策略 |
| v0.3.0 | Gather 助手、接入示例、CI examples |
| v0.4.0 | 性能优化与对比基准 |
| v0.5.0 | 工业级打磨与自我审查 |
