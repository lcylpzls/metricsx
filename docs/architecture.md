# metricsx 架构设计

> 状态:已实现(v1.6.1),本文描述当前架构;公开 API 以 `go doc` 与 README 为准。

## 1. 总体分层

```text
业务库(dbx/httpx/cachex/resiliencex/...)
└── Metrics 接口(统一形态)
    └── metricsx.Metrics(适配器)
        ├── 指标缓存(CounterVec / HistogramVec)
        ├── 预注册表(名称 → 标签键名/帮助)
        └── Prometheus Registry(默认或自定义)
```

## 2. 核心模块职责

| 模块 | 职责 |
| --- | --- |
| `metrics.go` | `Metrics`、`New`、`IncCounter`、`ObserveDuration` |
| `register.go` | `Register` 预注册、冲突策略 |
| `options.go` | `WithNamespace` / `WithRegistry` / `WithBuckets`、校验 |
| `errors.go` | `MTRX_*` 错误码 |

## 3. 指标创建

- CounterVec:名称拼接 `namespace_name_total`(client_golang 自动后缀);
- HistogramVec:名称拼接 `namespace_name_seconds`,默认分桶
  (1ms → 10s 指数序列);
- 懒创建:首次调用按 label 数量建 Vec,之后 map 查找复用;
- 未注册指标 label 键名占位 `label0..N`,预注册后使用声明键名。

## 4. 注册表

- 默认 `prometheus.DefaultRegisterer`(暴露到应用 /metrics 端点);
- `WithRegistry(r)` 注入自定义 Registry(测试隔离);
- `Registry()` 返回当前注册表。

## 5. 错误模型

| 场景 | 错误码 |
| --- | --- |
| 配置非法(namespace/buckets/注册表) | MTRX_INVALID_CONFIG |
| 重复注册(名称已存在) | MTRX_ALREADY_REGISTERED |

## 6. 依赖策略

- prometheus/client_golang(官方客户端,唯一第三方);
- errx(错误码),与底座生态一致。
