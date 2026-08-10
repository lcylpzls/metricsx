package metricsx

// Sink 是指标后端协议：核心只定义协议，不绑定具体实现。
// Prometheus 后端见 metricsx/prometheus 子模块。
type Sink interface {
	// IncCounter 增加一个计数指标。
	IncCounter(name string, labels []string)
	// AddCounter 按增量累加计数指标（可为负数）。
	AddCounter(name string, delta float64, labels []string)
	// ObserveDuration 记录一次耗时观测（秒）。
	ObserveDuration(name string, seconds float64, labels []string)
	// AddGauge 按增量调整瞬时量指标。
	AddGauge(name string, delta float64, labels []string)
	// SetGauge 将瞬时量指标设置为指定值。
	SetGauge(name string, value float64, labels []string)
	// RegisterMetric 预注册指标元信息（帮助文本与标签键名）。
	// 名称非法返回 MTRX_INVALID_CONFIG，重复注册返回 MTRX_ALREADY_REGISTERED。
	RegisterMetric(name, help string, labelNames []string) error
}
