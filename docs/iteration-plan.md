# metricsx 迭代计划与质量门槛

## 1. 迭代阶段

### P0 项目骨架

- go.mod(module github.com/lcylpzls/metricsx,go 1.26)、目录、CI
  (三平台 + staticcheck + govulncheck + tidy + apidiff + bench)。

### P1 v0.1.0 适配核心

- New / IncCounter / ObserveDuration / Register / Registry;
- 懒创建 CounterVec / HistogramVec,占位标签,默认分桶。

### P2 v0.2.0 配置与集成

- WithNamespace / WithRegistry / WithBuckets;
- gather 集成断言、重复注册冲突、校验。

### P3 v0.3.0 示例与观测

- Gather 快照助手;接入 dbx / httpx / resiliencex 组合示例;
- CI examples。

### P4 v0.4.0 性能

- 指标缓存无锁读;IncCounter 命中基准;对比手写适配器。

### P5 v0.5.0+ 工业级打磨与自我审查

- 治理文件、运行/质量/发布文档;
- 持续自审:命名规范、标签一致性、并发边界、性能;
- 成熟后停下征询用户是否发布 1.0.0。

## 2. 质量门槛(每阶段强制)

- 语句覆盖率 100%;`go vet` / `staticcheck` 零告警;
- `go test -race` 全绿;fuzz 至少 1 个目标(配置/名称);
- 三平台 CI × Go 1.26;govulncheck 零告警;
- go.mod tidy 漂移检查;apidiff 对比上一 tag;
- 所有日志、注释、文档使用简体中文。

## 3. 性能目标

| 场景 | 目标 |
| --- | --- |
| IncCounter 命中 | < 1µs,无额外分配 |
| ObserveDuration 命中 | < 1µs |
| 懒创建首次 | 仅一次注册开销 |

## 4. 风险与对策

| 风险 | 对策 |
| --- | --- |
| 名称冲突(不同库同名指标) | 库前缀天然区分;重复注册显式报错 |
| 标签键名缺失 | 未注册占位 label0..N,预注册声明键名 |
| 测试污染全局注册表 | WithRegistry 隔离 + 测试清理 |
