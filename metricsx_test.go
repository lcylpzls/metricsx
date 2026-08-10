package metricsx

import (
	"strings"
	"sync"
	"testing"

	testx "github.com/lcylpzls/testx"
)

func TestNewDefaultMemory(t *testing.T) {
	m, err := New()
	testx.RequireNoError(t, err)
	if _, ok := m.Sink().(*memorySink); !ok {
		t.Fatalf("默认后端应为内存实现：%T", m.Sink())
	}
	snap, ok := m.Snapshot()
	testx.True(t, ok)
	testx.Len(t, snap, 0)
}

func TestNewInvalidSink(t *testing.T) {
	_, err := New(WithSink(nil))
	testx.ErrCode(t, err, CodeInvalidConfig)
}

func TestNewNilOption(t *testing.T) {
	m, err := New(nil)
	testx.RequireNoError(t, err)
	testx.NotNil(t, m)
}

func TestCounterOperations(t *testing.T) {
	m, _ := New()
	m.IncCounter("requests", "a")
	m.IncCounter("requests", "a")
	m.IncCounter("requests", "b")
	m.AddCounter("bytes", 10, "a")
	m.AddCounter("bytes", -2, "a")
	snap, _ := m.Snapshot()
	testx.Approx(t, findValue(t, snap, "requests", "a"), 2, 0)
	testx.Approx(t, findValue(t, snap, "requests", "b"), 1, 0)
	testx.Approx(t, findValue(t, snap, "bytes", "a"), 8, 0)
}

func TestGaugeOperations(t *testing.T) {
	m, _ := New()
	m.AddGauge("active", 1)
	m.AddGauge("active", -1)
	m.SetGauge("active", 5)
	snap, _ := m.Snapshot()
	testx.Approx(t, findValue(t, snap, "active"), 5, 0)
}

func TestObserveDuration(t *testing.T) {
	m, _ := New()
	m.ObserveDuration("latency", 0.1, "GET")
	m.ObserveDuration("latency", 0.2, "GET")
	snap, _ := m.Snapshot()
	count, sum := findHistogram(t, snap, "latency", "GET")
	testx.Equal(t, count, uint64(2))
	testx.Approx(t, sum, 0.3, 1e-9)
}

func TestInvalidLabelsIgnored(t *testing.T) {
	m, _ := New()
	m.IncCounter("t", string([]byte{0xff}))
	snap, _ := m.Snapshot()
	testx.Len(t, snap, 0)
}

func TestRegister(t *testing.T) {
	m, _ := New()
	testx.RequireNoError(t, m.Register("req", "请求", "a"))
	err := m.Register("req", "请求", "a")
	testx.ErrCode(t, err, CodeAlreadyRegistered)
	err = m.Register("", "x")
	testx.ErrCode(t, err, CodeInvalidConfig)
}

func TestCustomSinkDelegation(t *testing.T) {
	sink := &fakeSink{}
	m, err := New(WithSink(sink))
	testx.RequireNoError(t, err)
	m.IncCounter("a", "x")
	m.AddCounter("b", 2, "x")
	m.ObserveDuration("c", 0.1, "x")
	m.AddGauge("d", 1, "x")
	m.SetGauge("e", 3, "x")
	_ = m.Register("r", "帮助", "x")
	sink.mu.Lock()
	defer sink.mu.Unlock()
	testx.Equal(t, strings.Join(sink.ops, ","), "inc,add_counter,observe,gauge_add,gauge_set,register")
	_, ok := m.Snapshot()
	testx.False(t, ok)
}

func TestConcurrentMemory(t *testing.T) {
	m, _ := New()
	var wg sync.WaitGroup
	for i := 0; i < 8; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			for j := 0; j < 100; j++ {
				m.IncCounter("t", "v")
			}
			_, _ = m.Snapshot()
		}()
	}
	wg.Wait()
	snap, _ := m.Snapshot()
	testx.Approx(t, findValue(t, snap, "t", "v"), 800, 0)
}

func findValue(t *testing.T, snap Snapshot, name string, labels ...string) float64 {
	t.Helper()
	for _, sm := range snap {
		if sm.Name == name && strings.Join(sm.Labels, ",") == strings.Join(labels, ",") {
			return sm.Value
		}
	}
	t.Fatalf("未找到指标 %s %v：%+v", name, labels, snap)
	return 0
}

func findHistogram(t *testing.T, snap Snapshot, name string, labels ...string) (uint64, float64) {
	t.Helper()
	for _, sm := range snap {
		if sm.Name == name && sm.Kind == "histogram" &&
			strings.Join(sm.Labels, ",") == strings.Join(labels, ",") {
			return sm.Count, sm.Sum
		}
	}
	t.Fatalf("未找到直方图 %s %v：%+v", name, labels, snap)
	return 0, 0
}

// fakeSink 是 Sink 的测试替身，记录操作顺序。
type fakeSink struct {
	mu  sync.Mutex
	ops []string
}

func (f *fakeSink) IncCounter(name string, labels []string) {
	f.mu.Lock()
	f.ops = append(f.ops, "inc")
	f.mu.Unlock()
}

func (f *fakeSink) AddCounter(name string, delta float64, labels []string) {
	f.mu.Lock()
	f.ops = append(f.ops, "add_counter")
	f.mu.Unlock()
}

func (f *fakeSink) ObserveDuration(name string, seconds float64, labels []string) {
	f.mu.Lock()
	f.ops = append(f.ops, "observe")
	f.mu.Unlock()
}

func (f *fakeSink) AddGauge(name string, delta float64, labels []string) {
	f.mu.Lock()
	f.ops = append(f.ops, "gauge_add")
	f.mu.Unlock()
}

func (f *fakeSink) SetGauge(name string, value float64, labels []string) {
	f.mu.Lock()
	f.ops = append(f.ops, "gauge_set")
	f.mu.Unlock()
}

func (f *fakeSink) RegisterMetric(name, help string, labelNames []string) error {
	f.mu.Lock()
	f.ops = append(f.ops, "register")
	f.mu.Unlock()
	return nil
}
