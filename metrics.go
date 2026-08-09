package metricsx

import (
	"fmt"
	"strings"
	"sync"
	"unicode/utf8"

	"github.com/lcylpzls/errx"
	"github.com/prometheus/client_golang/prometheus"
)

// registeredMetric 是预注册的指标元信息。
type registeredMetric struct {
	help       string
	labelNames []string
}

// Metrics 是 Prometheus 指标适配器,实现底座各库统一形态的
// Metrics 接口(IncCounter / ObserveDuration),并发安全。
type Metrics struct {
	cfg config

	mu         sync.RWMutex
	registered map[string]registeredMetric
	counters   map[string]*prometheus.CounterVec
	histograms map[string]*prometheus.HistogramVec
}

// New 创建指标适配器。配置非法返回 MTRX_INVALID_CONFIG。
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
	return &Metrics{
		cfg:        cfg,
		registered: make(map[string]registeredMetric),
		counters:   make(map[string]*prometheus.CounterVec),
		histograms: make(map[string]*prometheus.HistogramVec),
	}, nil
}

// IncCounter 增加一个计数指标。
// 指标未预注册时按 label 数量使用占位键 label0..N。
// 标签数量与注册键名不一致时静默忽略(不 panic)。
func (m *Metrics) IncCounter(name string, labels ...string) {
	if !validLabels(labels) {
		return
	}
	vec := m.counterVec(name, len(labels))
	if vec == nil {
		return
	}
	vec.WithLabelValues(labels...).Inc()
}

// ObserveDuration 记录一次耗时观测(秒)。
func (m *Metrics) ObserveDuration(name string, seconds float64, labels ...string) {
	if !validLabels(labels) {
		return
	}
	vec := m.histogramVec(name, len(labels))
	if vec == nil {
		return
	}
	vec.WithLabelValues(labels...).Observe(seconds)
}

// validLabels 校验标签值均为合法 UTF-8
// (client_golang 对非法 UTF-8 标签值会 panic,此处静默忽略)。
func validLabels(labels []string) bool {
	for _, l := range labels {
		if !utf8.ValidString(l) {
			return false
		}
	}
	return true
}

// Register 预注册指标,声明帮助文本与标签键名。
// 重复注册返回 MTRX_ALREADY_REGISTERED。
func (m *Metrics) Register(name, help string, labelNames ...string) error {
	if name == "" {
		return errx.New(errx.KindInvalid, CodeInvalidConfig, "指标名不能为空")
	}
	m.mu.Lock()
	defer m.mu.Unlock()
	if _, ok := m.registered[name]; ok {
		return errx.Newf(errx.KindInvalid, CodeAlreadyRegistered, "指标 %q 已注册", name)
	}
	if _, ok := m.counters[name]; ok {
		return errx.Newf(errx.KindInvalid, CodeAlreadyRegistered,
			"指标 %q 已被懒创建为计数器", name)
	}
	if _, ok := m.histograms[name]; ok {
		return errx.Newf(errx.KindInvalid, CodeAlreadyRegistered,
			"指标 %q 已被懒创建为直方图", name)
	}
	m.registered[name] = registeredMetric{help: help, labelNames: append([]string(nil), labelNames...)}
	return nil
}

// Registry 返回当前注册表(默认或自定义)。
func (m *Metrics) Registry() prometheus.Registerer {
	return m.cfg.registry
}

// counterVec 获取或懒创建计数器向量。
// 标签数量与注册键名不一致时返回 nil。
func (m *Metrics) counterVec(name string, labelCount int) *prometheus.CounterVec {
	m.mu.Lock()
	defer m.mu.Unlock()
	if vec, ok := m.counters[name]; ok {
		return vec
	}
	var vec *prometheus.CounterVec
	reg, ok := m.metricMeta(name, labelCount)
	if !ok {
		return nil
	}
	vec = prometheus.NewCounterVec(prometheus.CounterOpts{
		Namespace: m.cfg.namespace,
		Name:      sanitizeName(name) + "_total",
		Help:      reg.help,
	}, reg.labelNames)
	if err := m.cfg.registry.Register(vec); err != nil {
		// 注册失败(名称冲突等)不缓存,静默忽略本次。
		return nil
	}
	m.counters[name] = vec
	return vec
}

// histogramVec 获取或懒创建直方图向量。
func (m *Metrics) histogramVec(name string, labelCount int) *prometheus.HistogramVec {
	m.mu.Lock()
	defer m.mu.Unlock()
	if vec, ok := m.histograms[name]; ok {
		return vec
	}
	var vec *prometheus.HistogramVec
	reg, ok := m.metricMeta(name, labelCount)
	if !ok {
		return nil
	}
	vec = prometheus.NewHistogramVec(prometheus.HistogramOpts{
		Namespace: m.cfg.namespace,
		Name:      sanitizeName(name) + "_seconds",
		Help:      reg.help,
		Buckets:   m.cfg.buckets,
	}, reg.labelNames)
	if err := m.cfg.registry.Register(vec); err != nil {
		return nil
	}
	m.histograms[name] = vec
	return vec
}

// metricMeta 返回指标的帮助文本与标签键名。
// 标签数量与预注册键名不一致时返回 false。
func (m *Metrics) metricMeta(name string, labelCount int) (registeredMetric, bool) {
	if reg, ok := m.registered[name]; ok {
		if len(reg.labelNames) != labelCount {
			return registeredMetric{}, false
		}
		return reg, true
	}
	labels := make([]string, labelCount)
	for i := range labels {
		labels[i] = fmt.Sprintf("label%d", i)
	}
	return registeredMetric{help: name, labelNames: labels}, true
}

// sanitizeName 将指标名规范化为 Prometheus 合法名称:
// 非字母数字下划线冒号字符替换为下划线(如 "dbx.queries" → "dbx_queries")。
func sanitizeName(name string) string {
	if metricNamePattern.MatchString(name) {
		return name
	}
	var b strings.Builder
	b.Grow(len(name))
	for i := 0; i < len(name); i++ {
		c := name[i]
		if (c >= 'a' && c <= 'z') || (c >= 'A' && c <= 'Z') ||
			(c >= '0' && c <= '9') || c == '_' || c == ':' {
			b.WriteByte(c)
		} else {
			b.WriteByte('_')
		}
	}
	if b.Len() == 0 {
		return "metric"
	}
	return b.String()
}
