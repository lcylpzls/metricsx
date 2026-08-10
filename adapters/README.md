# metricsx/adapters

metricsx 与家族底座之间的官方适配层：把各库自定义的指标回调/接口
桥接到统一 metricsx 实例，实现“一个实例喂饱全家族”。

## 接入清单

| 底座 | 适配包 | 注入方式 |
| --- | --- | --- |
| filex | adapters/filex | `filex.Config.Metrics = fxadapter.New(m)` |
| jobx | adapters/jobx | `jobx.WithMetrics(jxadapter.New(m))` |
| updatex | adapters/updatex | `updatex.Config.Metrics = uxadapter.New(m)` |
| idgenx | adapters/idgenx | `idgenx.WithMetrics(ixadapter.New(m))` |
| errx | adapters/errx | `errxadapter.Install(m)`（全局钩子） |

## 示例

```go
m, err := metricsx.New(metricsx.WithNamespace("myapp"))
if err != nil {
	panic(err)
}

// 错误库（全局）
errxadapter.Install(m)

// 对象存储
store, err := filex.New(filex.Config{
	Dir:     "./data",
	Metrics: fxadapter.New(m),
})

// 任务调度
dispatcher, err := jobx.NewDispatcher(
	jobx.WithMetrics(jxadapter.New(m)),
)

// 自动升级
updater, err := updatex.New(updatex.Config{
	Source:        source,
	CurrentVersion: "v1.0.0",
	Metrics:        uxadapter.New(m),
})

// 分布式 ID
gen, err := idgenx.New(
	idgenx.WithNodeID(1),
	idgenx.WithMetrics(ixadapter.New(m)),
)
```

## 指标命名

- filex：`filex.operations` / `filex.operation_bytes` /
  `filex.errors`；
- jobx：`jobx.queued`（Gauge）/ `jobx.running`（Gauge）/
  `jobx.completed` / `jobx.completed_duration` / `jobx.failed` /
  `jobx.retried` / `jobx.dropped` / `jobx.skipped` /
  `jobx.replaced`；
- updatex：`updatex.checks` / `updatex.check_failures` /
  `updatex.updates`；
- idgenx：`idgenx.generated` / `idgenx.rejected` /
  `idgenx.wait_duration`；
- errx：`errx.constructed`（带 kind 标签）/ `errx.queried`。

命名统一走 metricsx 的 `库.主题` 规范，Prometheus 输出自动加
namespace 前缀与 `_total` / `_seconds` 后缀。

## 版本

当前适配层版本：**v0.1.0**（随 metricsx 仓库发布，
tag 为 `adapters/v0.1.0`）。
