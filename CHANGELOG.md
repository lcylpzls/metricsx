# 更新日志

本项目遵循[语义化版本](https://semver.org/lang/zh-CN/)。

## [v1.4.0] - 2026-08-10

### 破坏性重构

- 核心拆分为协议 + 内存后端：`metricsx` 仅保留 `Sink` 接口、
  `WithSink` 与零依赖内存实现，第三方依赖清零；
- Prometheus 后端迁移至新子模块 `metricsx/prometheus`
  （`WithPrometheus` / `WithNamespace` / `WithRegistry` /
  `WithBuckets` / `Gather` / `Registry`）；
- `metricsx.Gather` / `metricsx.Registry` 移出核心；
- `Register` 委托后端 `RegisterMetric`，冲突检测由后端执行；
- adapters 适配 `v1.4.0` 核心 API，测试改用 Prometheus 子模块。

### 质量

- 核心与 Prometheus 子模块覆盖率均 100%；
- race / vet / staticcheck / fuzz / govulncheck 全绿。

## [v1.3.3] - 2026-08-10

### 变更

- go 指令与 CI/Release 工作流统一为 Go 1.26.5；
- README Go 版本徽章同步更新。

## [v1.3.2] - 2026-08-10

### 变更

- 家族统一 Go 1.21：全部 go.mod 与 CI/Release 工作流版本号对齐 1.21；
- testx 依赖升级 v1.2.1。

## [Unreleased]

### 规划

- 完成调研、PRD、架构、API 草案、ADR 与迭代计划。

## [v1.3.1] - 2026-08-10

### 修复

- examples/basic 示例模块 go.mod 与最新依赖对齐（go mod tidy），
  修复 main CI 示例构建失败。

## [v1.3.0] - 2026-08-10

### 变更

- 家族测试底座接入：根包与 adapters 全部测试改用 testx 断言
  （语义等价改写，含 Require* 致命断言）；
- 测试依赖新增 `testx v1.2.0`，errx 同步升级 v1.4.0。

### 质量

- 根包与全部适配包语句覆盖率 100%；race / vet / staticcheck 全绿。

## [v1.2.0] - 2026-08-10

### 新增

- 官方适配层子模块 `github.com/lcylpzls/metricsx/adapters`：
  filex / jobx / updatex / idgenx / errx 五个底座桥接；
- CI 与 Release 增加 adapters 测试与 `adapters/vX.Y.Z` 发布支持；
- 家族指标统一接入规范文档（docs/observability-metrics.md）。

### 质量

- 适配层各包覆盖率 100%；race / vet / staticcheck 全绿。

## [v1.1.0] - 2026-08-10

### 新增

- AddCounter：按增量累加计数（字节量、批量事件）；
- AddGauge / SetGauge：瞬时量指标（活跃请求、连接数、任务水位）；
- GaugeVec 懒创建、预注册冲突检测与并发安全，与 Counter/Histogram 一致；
- 文档与 PRD 同步扩展（Gauge 场景由家族统一标准化接入驱动）。

### 质量

- 覆盖率保持 100%；race / vet / staticcheck / fuzz 全绿。

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

## [v0.5.0] - 2026-08-09

### 治理与文档

- SECURITY.md、CODEOWNERS、CONTRIBUTING、issue/PR 模板;
- operations(接入/命名/场景)、quality(质量门槛)、release(发布流程)、
  comparison(与 client_golang 对比)文档;
- 覆盖率 100%,race / vet / staticcheck / fuzz / vuln 全绿。

## [v1.0.0] - 2026-08-09

### 正式版

- 公开 API 冻结,遵循语义化版本;
- Version 常量更新为 v1.0.0;
- README 稳定性承诺正式生效;
- docs/api-design.md 升级为正式 API 参考;
- 全量回归:100% 覆盖率、race、staticcheck、fuzz、govulncheck、
  apidiff 对比 v0.5.0、三平台 CI。

### 版本历程

- v0.1.0 – v0.5.0:适配核心、冲突策略、Gather 与示例、
  性能优化、工业级打磨共 5 个迭代版本;
- v1.0.0:正式版,API 冻结。
