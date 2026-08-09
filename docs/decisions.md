# metricsx 架构决策记录(ADR)

> 状态说明:**已接受**(按全自动迭代授权,无需逐项确认)。

## ADR-001:单一适配器服务全部库

- **决策**:接口签名统一,一个 Metrics 实例可喂饱 dbx/httpx/
  cachex/resiliencex/jobx 等;
- **影响**:一次注入全栈可观测。

## ADR-002:懒创建 + 预注册

- **决策**:Vec 按首次调用懒创建;Register 声明帮助文本与标签键名;
- **影响**:零配置可用,需要可读标签时显式注册。

## ADR-003:命名规范对齐 Prometheus

- **决策**:计数器 `ns_name_total`、直方图 `ns_name_seconds`,
  namespace 可配;
- **影响**:Grafana 面板与告警规则可统一配置。

## ADR-004:注册表可注入

- **决策**:默认 DefaultRegisterer,WithRegistry 可隔离测试;
- **影响**:生产零配置,测试无污染。

## ADR-005:依赖仅 client_golang

- **决策**:接入层依赖官方客户端,不自研暴露协议;
- **影响**:与 dbx 依赖驱动同哲学,协议交给官方。
