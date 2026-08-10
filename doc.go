// Package metricsx 是家族指标基座核心：
// 定义统一指标协议（Sink）与内置内存后端，零第三方依赖；
// Prometheus 后端在 github.com/lcylpzls/metricsx/prometheus 子模块。
//
// 典型用法：
//
//	m, _ := metricsx.New()
//	m.IncCounter("orders.created", "region", "cn")
//	m.ObserveDuration("http.request", 0.12, "method", "GET")
package metricsx
