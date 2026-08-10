package metricsx

import (
	"strings"
	"sync"

	"github.com/lcylpzls/errx"
)

// SnapshotMetric 是内存后端单条指标快照。
type SnapshotMetric struct {
	// Name 指标名。
	Name string
	// Labels 标签值列表。
	Labels []string
	// Kind 指标类型：counter / gauge / histogram。
	Kind string
	// Value 当前值（直方图为 Sum）。
	Value float64
	// Count 观测次数（仅直方图）。
	Count uint64
	// Sum 观测总和（仅直方图）。
	Sum float64
}

// Snapshot 是内存后端指标快照。
type Snapshot []SnapshotMetric

// Snapshotter 是支持快照的后端可选接口。
type Snapshotter interface {
	Snapshot() Snapshot
}

// metricKey 是内存后端的指标键。
type metricKey struct {
	name   string
	labels string
}

// histEntry 是直方图条目。
type histEntry struct {
	count uint64
	sum   float64
}

// registeredMetric 是预注册的指标元信息。
type registeredMetric struct {
	help       string
	labelNames []string
}

// memorySink 是零依赖内存指标后端，用于默认、测试与调试。
type memorySink struct {
	mu         sync.RWMutex
	counters   map[metricKey]float64
	gauges     map[metricKey]float64
	histogram  map[metricKey]*histEntry
	registered map[string]registeredMetric
}

// newMemorySink 创建内存指标后端。
func newMemorySink() *memorySink {
	return &memorySink{
		counters:   make(map[metricKey]float64),
		gauges:     make(map[metricKey]float64),
		histogram:  make(map[metricKey]*histEntry),
		registered: make(map[string]registeredMetric),
	}
}

// IncCounter 增加一个计数指标。
func (s *memorySink) IncCounter(name string, labels []string) {
	s.AddCounter(name, 1, labels)
}

// AddCounter 按增量累加计数指标。
func (s *memorySink) AddCounter(name string, delta float64, labels []string) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.counters[keyOf(name, labels)] += delta
}

// ObserveDuration 记录一次耗时观测（秒）。
func (s *memorySink) ObserveDuration(name string, seconds float64, labels []string) {
	s.mu.Lock()
	defer s.mu.Unlock()
	k := keyOf(name, labels)
	e := s.histogram[k]
	if e == nil {
		e = &histEntry{}
		s.histogram[k] = e
	}
	e.count++
	e.sum += seconds
}

// AddGauge 按增量调整瞬时量指标。
func (s *memorySink) AddGauge(name string, delta float64, labels []string) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.gauges[keyOf(name, labels)] += delta
}

// SetGauge 将瞬时量指标设置为指定值。
func (s *memorySink) SetGauge(name string, value float64, labels []string) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.gauges[keyOf(name, labels)] = value
}

// RegisterMetric 预注册指标元信息。
func (s *memorySink) RegisterMetric(name, help string, labelNames []string) error {
	if name == "" {
		return errx.NewCode(CodeInvalidConfig, "指标名不能为空")
	}
	s.mu.Lock()
	defer s.mu.Unlock()
	if _, ok := s.registered[name]; ok {
		return errx.NewCodef(CodeAlreadyRegistered, "指标 %q 已注册", name)
	}
	s.registered[name] = registeredMetric{help: help, labelNames: append([]string(nil), labelNames...)}
	return nil
}

// Snapshot 返回全部指标快照。
func (s *memorySink) Snapshot() Snapshot {
	s.mu.RLock()
	defer s.mu.RUnlock()
	out := make(Snapshot, 0, len(s.counters)+len(s.gauges)+len(s.histogram))
	for k, v := range s.counters {
		out = append(out, SnapshotMetric{
			Name: k.name, Labels: splitLabels(k.labels), Kind: "counter", Value: v,
		})
	}
	for k, v := range s.gauges {
		out = append(out, SnapshotMetric{
			Name: k.name, Labels: splitLabels(k.labels), Kind: "gauge", Value: v,
		})
	}
	for k, e := range s.histogram {
		out = append(out, SnapshotMetric{
			Name: k.name, Labels: splitLabels(k.labels), Kind: "histogram",
			Count: e.count, Sum: e.sum, Value: e.sum,
		})
	}
	return out
}

// keyOf 构造指标键。
func keyOf(name string, labels []string) metricKey {
	return metricKey{name: name, labels: strings.Join(labels, "\x00")}
}

// splitLabels 拆分标签值。
func splitLabels(s string) []string {
	if s == "" {
		return nil
	}
	return strings.Split(s, "\x00")
}
