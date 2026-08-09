// Package metricsx 提供 Prometheus 指标适配层:
// 一个实例实现底座各库统一形态的 Metrics 接口,
// 懒创建 CounterVec / HistogramVec,统一命名与标签。
package metricsx
