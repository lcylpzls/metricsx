// Package prometheus 提供 metricsx 的 Prometheus 后端：
// 通过 WithPrometheus 注入 metricsx.New，懒创建 CounterVec /
// HistogramVec / GaugeVec，统一命名与标签。
package prometheus

import (
	"fmt"
	"net/http"
	"regexp"
	"strings"
	"sync"

	"github.com/lcylpzls/errx"
	"github.com/lcylpzls/metricsx"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	dto "github.com/prometheus/client_model/go"
)

// metricNamePattern 是 Prometheus 指标名合法字符（字母数字下划线冒号）。
var metricNamePattern = regexp.MustCompile(`^[a-zA-Z_:][a-zA-Z0-9_:]*$`)

// defaultBuckets 是 ObserveDuration 的默认直方图分桶（1ms → 10s 指数）。
var defaultBuckets = []float64{
	0.001, 0.002, 0.005, 0.01, 0.02, 0.05,
	0.1, 0.2, 0.5, 1, 2, 5, 10,
}

// Option 修改 Prometheus 后端配置。
type Option func(*sinkConfig)

// sinkConfig 是 Prometheus 后端配置。
type sinkConfig struct {
	namespace string
	registry  prometheus.Registerer
	buckets   []float64
}

func defaultSinkConfig() sinkConfig {
	return sinkConfig{
		registry: prometheus.DefaultRegisterer,
		buckets:  defaultBuckets,
	}
}

// WithNamespace 设置指标命名空间前缀（如 "myapp"），空串表示无前缀。
func WithNamespace(ns string) Option {
	return func(c *sinkConfig) { c.namespace = ns }
}

// WithRegistry 注入注册表，默认 prometheus.DefaultRegisterer。
func WithRegistry(r prometheus.Registerer) Option {
	return func(c *sinkConfig) { c.registry = r }
}

// WithBuckets 设置 ObserveDuration 的直方图分桶，空串表示使用默认。
func WithBuckets(buckets []float64) Option {
	return func(c *sinkConfig) { c.buckets = buckets }
}

// WithPrometheus 返回注入 Prometheus 后端的 metricsx 选项。
func WithPrometheus(opts ...Option) metricsx.Option {
	cfg := defaultSinkConfig()
	for _, opt := range opts {
		if opt != nil {
			opt(&cfg)
		}
	}
	return metricsx.WithSink(newSink(cfg))
}

// promSink 是 metricsx.Sink 的 Prometheus 实现。
type promSink struct {
	cfg sinkConfig

	mu         sync.RWMutex
	registered map[string]registeredMetric
	counters   sync.Map // string -> *prometheus.CounterVec
	histograms sync.Map // string -> *prometheus.HistogramVec
	gauges     sync.Map // string -> *prometheus.GaugeVec
	// lockHook 是测试专用并发分支钩子，生产恒为 nil。
	lockHook func(*promSink)
}

// registeredMetric 是预注册的指标元信息。
type registeredMetric struct {
	help       string
	labelNames []string
}

// newSink 创建 Prometheus 后端。
func newSink(cfg sinkConfig) *promSink {
	return &promSink{
		cfg:        cfg,
		registered: make(map[string]registeredMetric),
	}
}

// IncCounter 增加一个计数指标。
func (s *promSink) IncCounter(name string, labels []string) {
	vec := s.counterVec(name, len(labels))
	if vec == nil {
		return
	}
	vec.WithLabelValues(labels...).Inc()
}

// AddCounter 按增量累加计数指标。
func (s *promSink) AddCounter(name string, delta float64, labels []string) {
	vec := s.counterVec(name, len(labels))
	if vec == nil {
		return
	}
	vec.WithLabelValues(labels...).Add(delta)
}

// ObserveDuration 记录一次耗时观测（秒）。
func (s *promSink) ObserveDuration(name string, seconds float64, labels []string) {
	vec := s.histogramVec(name, len(labels))
	if vec == nil {
		return
	}
	vec.WithLabelValues(labels...).Observe(seconds)
}

// AddGauge 按增量调整瞬时量指标。
func (s *promSink) AddGauge(name string, delta float64, labels []string) {
	vec := s.gaugeVec(name, len(labels))
	if vec == nil {
		return
	}
	vec.WithLabelValues(labels...).Add(delta)
}

// SetGauge 将瞬时量指标设置为指定值。
func (s *promSink) SetGauge(name string, value float64, labels []string) {
	vec := s.gaugeVec(name, len(labels))
	if vec == nil {
		return
	}
	vec.WithLabelValues(labels...).Set(value)
}

// RegisterMetric 预注册指标元信息。
func (s *promSink) RegisterMetric(name, help string, labelNames []string) error {
	if name == "" {
		return errx.NewCode(metricsx.CodeInvalidConfig, "指标名不能为空")
	}
	s.mu.Lock()
	defer s.mu.Unlock()
	if _, ok := s.registered[name]; ok {
		return errx.NewCodef(metricsx.CodeAlreadyRegistered, "指标 %q 已注册", name)
	}
	if _, ok := s.counters.Load(name); ok {
		return errx.NewCodef(metricsx.CodeAlreadyRegistered,
			"指标 %q 已被懒创建为计数器", name)
	}
	if _, ok := s.histograms.Load(name); ok {
		return errx.NewCodef(metricsx.CodeAlreadyRegistered,
			"指标 %q 已被懒创建为直方图", name)
	}
	if _, ok := s.gauges.Load(name); ok {
		return errx.NewCodef(metricsx.CodeAlreadyRegistered,
			"指标 %q 已被懒创建为瞬时量", name)
	}
	s.registered[name] = registeredMetric{help: help, labelNames: append([]string(nil), labelNames...)}
	return nil
}

// Registry 返回后端注册表；后端不是 Prometheus 时返回 false。
func Registry(m *metricsx.Metrics) (prometheus.Registerer, bool) {
	s, ok := m.Sink().(*promSink)
	if !ok {
		return nil, false
	}
	return s.cfg.registry, true
}

// Gather 返回后端注册表全部指标快照。
func Gather(m *metricsx.Metrics) ([]*dto.MetricFamily, error) {
	reg, ok := Registry(m)
	if !ok {
		return nil, errx.NewCode(metricsx.CodeInvalidConfig, "指标后端不是 Prometheus")
	}
	g, ok := reg.(prometheus.Gatherer)
	if !ok {
		return nil, errx.NewCode(metricsx.CodeInvalidConfig, "注册表不支持 Gather")
	}
	return g.Gather()
}

// HTTPHandler 返回导出 Prometheus 文本格式的 http.Handler。
// 后端不是 Prometheus 时返回 500 错误处理器，避免业务侧直接依赖
// prometheus/client_golang。
func HTTPHandler(m *metricsx.Metrics) http.Handler {
	reg, ok := Registry(m)
	if !ok {
		return http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
			http.Error(w, "metrics backend is not prometheus", http.StatusInternalServerError)
		})
	}
	g, ok := reg.(prometheus.Gatherer)
	if !ok {
		return http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
			http.Error(w, "prometheus registry does not support gather", http.StatusInternalServerError)
		})
	}
	return promhttp.HandlerFor(g, promhttp.HandlerOpts{})
}

// counterVec 获取或懒创建计数器向量。
func (s *promSink) counterVec(name string, labelCount int) *prometheus.CounterVec {
	if v, ok := s.counters.Load(name); ok {
		return v.(*prometheus.CounterVec)
	}
	s.mu.Lock()
	defer s.mu.Unlock()
	if s.lockHook != nil {
		s.lockHook(s)
	}
	if v, ok := s.counters.Load(name); ok {
		return v.(*prometheus.CounterVec)
	}
	reg, ok := s.metricMeta(name, labelCount)
	if !ok {
		return nil
	}
	vec := prometheus.NewCounterVec(prometheus.CounterOpts{
		Namespace: s.cfg.namespace,
		Name:      sanitizeName(name) + "_total",
		Help:      reg.help,
	}, reg.labelNames)
	if err := s.cfg.registry.Register(vec); err != nil {
		return nil
	}
	s.counters.Store(name, vec)
	return vec
}

// histogramVec 获取或懒创建直方图向量。
func (s *promSink) histogramVec(name string, labelCount int) *prometheus.HistogramVec {
	if v, ok := s.histograms.Load(name); ok {
		return v.(*prometheus.HistogramVec)
	}
	s.mu.Lock()
	defer s.mu.Unlock()
	if s.lockHook != nil {
		s.lockHook(s)
	}
	if v, ok := s.histograms.Load(name); ok {
		return v.(*prometheus.HistogramVec)
	}
	reg, ok := s.metricMeta(name, labelCount)
	if !ok {
		return nil
	}
	vec := prometheus.NewHistogramVec(prometheus.HistogramOpts{
		Namespace: s.cfg.namespace,
		Name:      sanitizeName(name) + "_seconds",
		Help:      reg.help,
		Buckets:   s.cfg.buckets,
	}, reg.labelNames)
	if err := s.cfg.registry.Register(vec); err != nil {
		return nil
	}
	s.histograms.Store(name, vec)
	return vec
}

// gaugeVec 获取或懒创建瞬时量向量。
func (s *promSink) gaugeVec(name string, labelCount int) *prometheus.GaugeVec {
	if v, ok := s.gauges.Load(name); ok {
		return v.(*prometheus.GaugeVec)
	}
	s.mu.Lock()
	defer s.mu.Unlock()
	if s.lockHook != nil {
		s.lockHook(s)
	}
	if v, ok := s.gauges.Load(name); ok {
		return v.(*prometheus.GaugeVec)
	}
	reg, ok := s.metricMeta(name, labelCount)
	if !ok {
		return nil
	}
	vec := prometheus.NewGaugeVec(prometheus.GaugeOpts{
		Namespace: s.cfg.namespace,
		Name:      sanitizeName(name),
		Help:      reg.help,
	}, reg.labelNames)
	if err := s.cfg.registry.Register(vec); err != nil {
		return nil
	}
	s.gauges.Store(name, vec)
	return vec
}

// metricMeta 返回指标的帮助文本与标签键名。
func (s *promSink) metricMeta(name string, labelCount int) (registeredMetric, bool) {
	if reg, ok := s.registered[name]; ok {
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

// sanitizeName 将指标名规范化为 Prometheus 合法名称。
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
