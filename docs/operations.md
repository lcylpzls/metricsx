# 运行手册

## 快速接入

```go
m, err := metricsx.New(metricsx.WithNamespace("myapp"))
if err != nil {
	panic(err)
}

// 预注册可读标签(可选,推荐)
m.Register("dbx.queries", "数据库查询次数", "op")
m.Register("httpx.duration", "HTTP 请求耗时", "method")

// 注入各库
dbx.OpenConfig(ctx, dbx.Config{DSN: dsn, Metrics: m})
httpx.New(httpx.WithMetrics(m))
cachex.New(cachex.WithMetrics(m))
resiliencex.NewTokenBucket(100, 10, resiliencex.WithMetrics(m))
```

## 命名规范

| 调用 | Prometheus 指标名 |
| --- | --- |
| IncCounter("dbx.queries", "select") | myapp_dbx_queries_total{label0="select"} |
| Register 后 | myapp_dbx_queries_total{op="select"} |
| ObserveDuration("httpx.duration", 0.5, "GET") | myapp_httpx_duration_seconds{label0="GET"} |

未注册指标标签键名为 label0..N;建议对需要可读标签的指标预注册。

## 常见场景

### 暴露 /metrics 端点(webx)

默认使用 DefaultRegisterer,webx 或任意 HTTP 端点挂
`promhttp.Handler()` 即可暴露全部指标。

### 测试隔离

```go
reg := prometheus.NewRegistry()
m, _ := metricsx.New(metricsx.WithRegistry(reg))
// 断言:reg.Gather() 或 m.Gather()
```

## 注意事项

- 指标名中的点号等非法字符自动规范化为下划线;
- 非法 UTF-8 标签值静默忽略,不 panic;
- 标签数量与预注册键名不一致时静默忽略;
- 注册失败(名称冲突)静默忽略,部署前用 Gather 核对;
- 标签应来自低基数集合,避免高基数标签撑爆 Prometheus。
