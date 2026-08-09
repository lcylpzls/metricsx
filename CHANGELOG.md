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
