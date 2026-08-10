# 性能基准

## 方法

```powershell
go test -run '^$' -bench . -benchmem -benchtime=1s .
```

CI 的 bench job 记录每次 main 推送的基准日志(artifact),不设硬性门禁。

## 参考数据(v1.1.0,Windows / AMD Ryzen 5 7600)

| Benchmark | ns/op | B/op | allocs/op |
| --- | --- | --- | --- |
| IncCounter 命中 | 33.1 | 0 | 0 |
| ObserveDuration 命中 | 41.4 | 0 | 0 |
| AddCounter 命中 | 36.5 | 0 | 0 |
| AddGauge 命中 | 35.3 | 0 | 0 |
| SetGauge 命中 | 34.8 | 0 | 0 |
| DirectClient(对照组) | 20.6 | 0 | 0 |

## 解读

- 适配层总成本(查找 + 标签校验 + client_golang 写入)约 33~41ns,
  约为直接使用 client_golang Vec(21ns)的 2 倍——额外成本
  换来统一接口、懒创建、命名规范化与标签校验;
- 指标缓存为 sync.Map:读路径无锁,
  并发懒创建由互斥锁 + 注册失败静默忽略保证正确;
- 热路径 0 分配,远低于 1µs 目标。

## 优化原则

- 热路径无分配;指标查找无锁读;
- 懒创建仅首次开销(注册一次);
- 命名规范化与标签校验不引入热路径成本。
