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
