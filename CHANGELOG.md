# 更新日志

本项目遵循[语义化版本](https://semver.org/lang/zh-CN/)。

## [Unreleased]

### 规划

- 完成调研、PRD、架构、API 草案、ADR 与迭代计划。

## [v0.1.0] - 2026-08-09

### 新增

- New / IncCounter / ObserveDuration / Register / Registry;
- 懒创建 CounterVec / HistogramVec,占位标签 label0..N;
- 命名规范化(点号等下划线),namespace 前缀,
  计数器 _total / 直方图 _seconds;
- 预注册声明帮助文本与标签键名,重复注册报错;
- 非法 UTF-8 标签值静默忽略(fuzz 发现,防 panic);
- 依赖仅 prometheus/client_golang;覆盖率 100%,
  race / vet / staticcheck / fuzz / vuln 全绿;
- IncCounter 命中 32ns,ObserveDuration 38ns,0 分配。

## [v0.2.0] - 2026-08-09

### 修复与完善

- Register 冲突策略:先懒创建后注册(计数器/直方图)明确报错,
  保证 help 与标签键名一致性;
- Register 空指标名校验(MTRX_INVALID_CONFIG);
- 覆盖率 100%,race / vet / staticcheck / fuzz / vuln 全绿。

## [v0.3.0] - 2026-08-09

### 新增

- Gather 快照助手:抓取注册表全部指标(测试与调试);
- examples/basic:一个实例喂饱 cachex 与 resiliencex 的组合示例;
- CI 新增 examples job;
- 覆盖率 100%,race / vet / staticcheck / fuzz / vuln 全绿。

## [v0.4.0] - 2026-08-09

### 性能

- 指标缓存改 sync.Map:读路径无锁,并发懒创建安全;
- IncCounter 命中 42ns / ObserveDuration 43ns,0 分配;
- 与直接 client_golang 对比基准(21ns,适配成本约 2 倍,
  换来统一接口/懒创建/命名规范化/标签校验);
- docs/performance.md;
- 覆盖率 100%,race / vet / staticcheck / fuzz / vuln 全绿。
