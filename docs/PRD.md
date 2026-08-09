# metricsx 产品需求(PRD)

> 版本:v1.0.0(正式版) · 状态:已发布

## 1. 背景与动机

底座 11 个库都定义了统一的 `Metrics` 接口
(`IncCounter(name, labels...)` / `ObserveDuration(name, seconds, labels...)`),
但都没有实现——每个项目接入 Prometheus 都要手写适配器,
指标名/标签风格各异,无法统一观测。

结论:**做一个 Prometheus 适配层**,一个实例喂饱全部库,
统一命名、统一标签、统一分桶。

## 2. 目标

1. 实现统一 Metrics 接口,懒创建 CounterVec / HistogramVec;
2. `Register(name, help, labelNames...)` 预注册,声明标签键名;
3. namespace 统一前缀,ObserveDuration 映射 Histogram(秒);
4. 默认注册表直接暴露 /metrics,自定义注册表隔离测试;
5. 依赖仅 prometheus/client_golang(接入层);
6. 100% 覆盖,与底座质量门槛一致。

## 3. 非目标(明确不做)

- 不自研 Prometheus 暴露协议 / 文本格式;
- 不绑定具体 Web 框架(webx 自行接入);
- 不做 Gauge / Summary 封装(当前接口只需 Counter + Histogram);
- 不做指标动态改名/聚合(标签维度由调用方负责)。

## 4. 能力需求

### 4.1 核心(v0.1.0)

- `New(opts...) (*Metrics, error)`;
- `IncCounter` / `ObserveDuration`(懒创建 Vec);
- `Register(name, help, labelNames...)` 预注册;
- `Registry()` 暴露注册表(自定义模式);
- 未注册指标自动按 `label0..N` 占位键创建。

### 4.2 配置(v0.2.0)

- `WithNamespace(ns)` / `WithRegistry(r)` / `WithBuckets(b)` / `WithHelp?`;
- 预注册与懒创建冲突策略(重复注册错误);
- 与真实 Prometheus gather 的集成断言。

### 4.3 示例与观测(v0.3.0)

- 接入 dbx / httpx / resiliencex / jobx 组合示例;
- `Gather()` 快照助手(测试与调试)。

## 5. 非功能需求

- **性能**:指标查找缓存无锁读,IncCounter 命中 < 1µs;
- **质量**:语句覆盖率 100%、race、staticcheck、vet、fuzz、三平台 CI;
- **依赖**:prometheus/client_golang(唯一第三方)。

## 6. 验收标准

v0.1.0 发布时:

1. 懒创建 / 预注册 / 占位标签全路径测试;
2. gather 输出命名与分桶断言;
3. 配置校验与错误码;
4. 100% 语句覆盖率,race / staticcheck / vet 全绿。
