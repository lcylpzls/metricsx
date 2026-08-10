# 家族指标统一接入规范

metricsx 是家族唯一的指标采集底座。各基座库只暴露零依赖接口或
回调结构，不内置 Prometheus 实现；metricsx/adapters 提供官方桥接。

## 接入矩阵

| 基座 | 接口形态 | 接入方式 | 适配器 |
| --- | --- | --- | --- |
| cachex | `Metrics`（IncCounter/ObserveDuration） | `WithMetrics(m)` 直连 | 无需适配器 |
| dbx | `Metrics` | `WithMetrics(m)` 直连 | 无需适配器 |
| httpx | `Metrics` | `WithMetrics(m)` 直连 | 无需适配器 |
| resiliencex | `Metrics` | `WithMetrics(m)` 直连 | 无需适配器 |
| webx（v2） | `Metrics` + 可选 `GaugeMetrics` | `WithMetrics(m)` 直连 | 无需适配器 |
| logx | `MetricSink` + 可选 `CounterSink` | `Builder.WithMetrics(m)` 直连 | 无需适配器 |
| filex | `Metrics`（Add/IncError） | `Config.Metrics` | adapters/filex |
| jobx | 回调结构 | `WithMetrics` | adapters/jobx |
| updatex | 回调结构 | `Config.Metrics` | adapters/updatex |
| idgenx | 回调结构 | `WithMetrics` | adapters/idgenx |
| errx | 全局钩子 | `SetMetricsHook` | adapters/errx |

## 原则

1. 基座库零 Prometheus 依赖，只定义最小接口；
2. 同一进程只创建一个 metricsx 实例并注入全家族；
3. 指标名统一 `库.主题`，标签维度由适配层定义并保持稳定；
4. 热路径默认零开销（未注入时仅 nil 判断 / 原子加载）。

## 发布约定

- 适配层为独立 Go 模块 `github.com/lcylpzls/metricsx/adapters`，
  版本 tag 形如 `adapters/vX.Y.Z`；
- 主模块 metricsx 版本与适配层版本独立推进。
