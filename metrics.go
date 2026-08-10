package metricsx

import (
	"unicode/utf8"

	"github.com/lcylpzls/errx"
	"github.com/lcylpzls/validx"
)

// Metrics 是指标入口：所有操作委托后端（Sink），并发安全。
// 默认后端为内置内存实现；Prometheus 后端见 metricsx/prometheus。
type Metrics struct {
	sink Sink
}

// New 创建指标入口。配置非法返回 MTRX_INVALID_CONFIG。
func New(opts ...Option) (*Metrics, error) {
	cfg := defaultConfig()
	for _, opt := range opts {
		if opt != nil {
			opt(&cfg)
		}
	}
	if err := validateConfig(cfg); err != nil {
		return nil, err
	}
	if cfg.sink == nil {
		cfg.sink = newMemorySink()
	}
	return &Metrics{sink: cfg.sink}, nil
}

// Sink 返回当前指标后端。
func (m *Metrics) Sink() Sink {
	return m.sink
}

// Snapshot 返回后端快照；后端不支持快照时返回 false。
func (m *Metrics) Snapshot() (Snapshot, bool) {
	if s, ok := m.sink.(Snapshotter); ok {
		return s.Snapshot(), true
	}
	return nil, false
}

// IncCounter 增加一个计数指标。
func (m *Metrics) IncCounter(name string, labels ...string) {
	if !validLabels(labels) {
		return
	}
	m.sink.IncCounter(name, labels)
}

// ObserveDuration 记录一次耗时观测（秒）。
func (m *Metrics) ObserveDuration(name string, seconds float64, labels ...string) {
	if !validLabels(labels) {
		return
	}
	m.sink.ObserveDuration(name, seconds, labels)
}

// AddCounter 按增量累加计数指标。
func (m *Metrics) AddCounter(name string, delta float64, labels ...string) {
	if !validLabels(labels) {
		return
	}
	m.sink.AddCounter(name, delta, labels)
}

// AddGauge 按增量调整瞬时量指标。
func (m *Metrics) AddGauge(name string, delta float64, labels ...string) {
	if !validLabels(labels) {
		return
	}
	m.sink.AddGauge(name, delta, labels)
}

// SetGauge 将瞬时量指标设置为指定值。
func (m *Metrics) SetGauge(name string, value float64, labels ...string) {
	if !validLabels(labels) {
		return
	}
	m.sink.SetGauge(name, value, labels)
}

// Register 预注册指标，声明帮助文本与标签键名。
// 名称非法或重复注册返回 errx 错误。
func (m *Metrics) Register(name, help string, labelNames ...string) error {
	return m.sink.RegisterMetric(name, help, labelNames)
}

// init 注册标签校验规则到 validx 全局规则表。
func init() {
	_ = validx.RegisterRule("metricsx_valid_labels", func(value any, param, path string) error {
		// 内部调用保证 value 为 []string。
		labels := value.([]string)
		for _, l := range labels {
			if !utf8.ValidString(l) {
				return errx.NewCode(CodeInvalidConfig, "标签值必须是合法 UTF-8")
			}
		}
		return nil
	})
}

// validLabels 校验标签值均为合法 UTF-8（非法标签静默忽略，统一走 validx 规则）。
func validLabels(labels []string) bool {
	return validx.ValidateField(labels, "metricsx_valid_labels") == nil
}
