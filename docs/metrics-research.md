# 指标适配设计调研手册

> 调研日期:2026-08-09 · 范围:Prometheus 客户端 / 自家库 Metrics 接口 / 常见适配器
> 目的:只取设计思想,代码自研(仅依赖官方客户端)。

## 1. prometheus/client_golang(官方客户端)

**设计思想**

- Counter / Gauge / Histogram / Summary 四类指标,Vec 支持标签维度;
- 注册表(Registry):全局 DefaultRegisterer 与自定义 Registry 隔离;
- 命名规范:计数器 `*_total`、单位 `*_seconds`、namespace 前缀;
- Histogram 分桶:预定义边界,记录观测分布。

**取其精华**:Vec 懒创建与注册表隔离、命名规范、分桶模型。

**去其糟粕**:直接使用 API 冗长(CounterVec 的 WithLabelValues),
各业务库重复拼装。

## 2. 自家生态 Metrics 接口(11 个库统一形态)

**设计思想**

- `IncCounter(name, labels...)` / `ObserveDuration(name, seconds, labels...)`;
- 所有库(dbx/httpx/webx/cachex/resiliencex/jobx/authx)签名一致,
  指标名统一 `库.主题` 风格;
- 默认 no-op,注入才生效。

**取其精华**:接口统一是最大红利——单一适配器可服务全部库。

**去其糟粕**:接口只传 label 值不传键名,适配层需预注册声明键名。

## 3. 常见 Web 适配器(gin-prometheus / echo-prometheus)

**设计思想**:中间件把 HTTP 请求/耗时转成 Prometheus 指标,
统一 namespace 与标签。

**取其精华**:namespace 前缀、标签顺序固定、分桶默认值。

**去其糟粕**:与具体框架绑定,不可复用。

## 4. 设计思想汇总

- 单一适配器实现统一 Metrics 接口,懒创建 CounterVec / HistogramVec;
- 预注册(Register)声明指标帮助文本与标签键名,未注册自动占位键;
- namespace 统一前缀,ObserveDuration 映射 Histogram(秒单位);
- 默认注册表可直接暴露 /metrics,自定义注册表隔离测试;
- 依赖仅 client_golang(接入层,同 dbx 依赖驱动)。

## 5. 明确不采纳

- 自研 Prometheus 暴露协议(gather/text 格式);
- 与具体框架绑定的中间件适配;
- 全局单例注册表(显式 New + 可注入)。

> 本手册为 metricsx 设计输入,不构成对外承诺;实现时按需取舍。
