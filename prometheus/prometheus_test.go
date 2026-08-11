package prometheus

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/lcylpzls/metricsx"
	testx "github.com/lcylpzls/testx"
	"github.com/prometheus/client_golang/prometheus"
)

func TestPrometheusSinkEndToEnd(t *testing.T) {
	reg := prometheus.NewRegistry()
	m, err := metricsx.New(WithPrometheus(WithNamespace("demo"), WithRegistry(reg)))
	testx.RequireNoError(t, err)
	m.IncCounter("requests", "a")
	m.AddCounter("bytes", 10, "a")
	m.ObserveDuration("latency", 0.1, "GET")
	m.AddGauge("active", 1)
	m.SetGauge("active", 2)
	families, err := Gather(m)
	testx.RequireNoError(t, err)
	names := map[string]bool{}
	for _, f := range families {
		names[f.GetName()] = true
	}
	for _, want := range []string{
		"demo_requests_total", "demo_bytes_total",
		"demo_latency_seconds", "demo_active",
	} {
		testx.True(t, names[want])
	}
}

// TestHTTPHandler 覆盖 Prometheus 后端文本导出。
func TestHTTPHandler(t *testing.T) {
	reg := prometheus.NewRegistry()
	m, err := metricsx.New(WithPrometheus(WithNamespace("demo"), WithRegistry(reg)))
	testx.RequireNoError(t, err)
	m.IncCounter("requests", "a")

	rec := httptest.NewRecorder()
	HTTPHandler(m).ServeHTTP(rec, httptest.NewRequest(http.MethodGet, "/metrics", nil))
	testx.RequireEqual(t, rec.Code, http.StatusOK)
	testx.RequireTrue(t, strings.Contains(rec.Body.String(), "demo_requests_total"))
}

// TestHTTPHandlerNotPrometheus 覆盖非 Prometheus 后端。
func TestHTTPHandlerNotPrometheus(t *testing.T) {
	m, err := metricsx.New()
	testx.RequireNoError(t, err)
	rec := httptest.NewRecorder()
	HTTPHandler(m).ServeHTTP(rec, httptest.NewRequest(http.MethodGet, "/metrics", nil))
	testx.RequireEqual(t, rec.Code, http.StatusInternalServerError)
}

// TestHTTPHandlerNotGatherer 覆盖注册表不支持 Gather 的分支。
func TestHTTPHandlerNotGatherer(t *testing.T) {
	m, err := metricsx.New(WithPrometheus(WithRegistry(fakeRegisterer{})))
	testx.RequireNoError(t, err)
	rec := httptest.NewRecorder()
	HTTPHandler(m).ServeHTTP(rec, httptest.NewRequest(http.MethodGet, "/metrics", nil))
	testx.RequireEqual(t, rec.Code, http.StatusInternalServerError)
}

func TestRegistry(t *testing.T) {
	reg := prometheus.NewRegistry()
	m, _ := metricsx.New(WithPrometheus(WithRegistry(reg)))
	got, ok := Registry(m)
	testx.True(t, ok)
	testx.Equal(t, got, reg)

	mm, _ := metricsx.New()
	_, ok = Registry(mm)
	testx.False(t, ok)
}

func TestGatherErrors(t *testing.T) {
	mm, _ := metricsx.New()
	_, err := Gather(mm)
	testx.ErrCode(t, err, metricsx.CodeInvalidConfig)

	m, _ := metricsx.New(WithPrometheus(WithRegistry(fakeRegisterer{})))
	_, err = Gather(m)
	testx.ErrCode(t, err, metricsx.CodeInvalidConfig)
}

func TestRegisterMetric(t *testing.T) {
	reg := prometheus.NewRegistry()
	m, _ := metricsx.New(WithPrometheus(WithRegistry(reg)))
	testx.RequireNoError(t, m.Register("req", "请求", "a"))
	err := m.Register("req", "请求", "a")
	testx.ErrCode(t, err, metricsx.CodeAlreadyRegistered)
	err = m.Register("", "x")
	testx.ErrCode(t, err, metricsx.CodeInvalidConfig)

	m.IncCounter("lazy", "v")
	err = m.Register("lazy", "冲突")
	testx.ErrCode(t, err, metricsx.CodeAlreadyRegistered)

	m.ObserveDuration("lazy_h", 1, "v")
	err = m.Register("lazy_h", "冲突")
	testx.ErrCode(t, err, metricsx.CodeAlreadyRegistered)

	m.AddGauge("lazy_g", 1, "v")
	err = m.Register("lazy_g", "冲突")
	testx.ErrCode(t, err, metricsx.CodeAlreadyRegistered)

	m.IncCounter("req", "x")
}

func TestLabelMismatchIgnored(t *testing.T) {
	reg := prometheus.NewRegistry()
	m, _ := metricsx.New(WithPrometheus(WithRegistry(reg)))
	_ = m.Register("req", "请求", "a")
	m.IncCounter("req", "x", "y")
	families, err := Gather(m)
	testx.RequireNoError(t, err)
	testx.Len(t, families, 0)
}

func TestSanitizeName(t *testing.T) {
	testx.Equal(t, sanitizeName("dbx.queries"), "dbx_queries")
	testx.Equal(t, sanitizeName("valid_name"), "valid_name")
	testx.Equal(t, sanitizeName("..."), "___")
}

func TestCustomBuckets(t *testing.T) {
	reg := prometheus.NewRegistry()
	m, err := metricsx.New(WithPrometheus(
		WithRegistry(reg),
		WithBuckets([]float64{0.01, 0.1, 1}),
	))
	testx.RequireNoError(t, err)
	m.ObserveDuration("latency", 0.05, "GET")
	families, err := Gather(m)
	testx.RequireNoError(t, err)
	testx.Len(t, families, 1)
}

func TestHistogramHit(t *testing.T) {
	reg := prometheus.NewRegistry()
	m, _ := metricsx.New(WithPrometheus(WithRegistry(reg)))
	m.ObserveDuration("h", 0.1, "v")
	m.ObserveDuration("h", 0.2, "v")
	families, err := Gather(m)
	testx.RequireNoError(t, err)
	testx.Len(t, families, 1)
}

func TestSinkNilVecBranches(t *testing.T) {
	reg := prometheus.NewRegistry()
	m, _ := metricsx.New(WithPrometheus(WithRegistry(reg)))
	_ = m.Register("req", "请求", "a")
	m.AddCounter("req", 1, "x", "y")
	m.ObserveDuration("req", 1, "x", "y")
	m.AddGauge("req", 1, "x", "y")
	m.SetGauge("req", 1, "x", "y")
	families, err := Gather(m)
	testx.RequireNoError(t, err)
	testx.Len(t, families, 0)
}

func TestSinkRegisterConflictBranches(t *testing.T) {
	reg := prometheus.NewRegistry()
	reg.MustRegister(prometheus.NewCounter(prometheus.CounterOpts{Name: "x_total"}))
	reg.MustRegister(prometheus.NewHistogram(prometheus.HistogramOpts{Name: "y_seconds", Buckets: defaultBuckets}))
	reg.MustRegister(prometheus.NewGauge(prometheus.GaugeOpts{Name: "z"}))
	m, _ := metricsx.New(WithPrometheus(WithRegistry(reg)))
	m.IncCounter("x", "v")
	m.ObserveDuration("y", 1, "v")
	m.AddGauge("z", 1, "v")
	families, err := Gather(m)
	testx.RequireNoError(t, err)
	testx.Len(t, families, 3)
}

func TestLazyCreateInnerCheck(t *testing.T) {
	reg := prometheus.NewRegistry()
	m, _ := metricsx.New(WithPrometheus(WithRegistry(reg)))
	s := m.Sink().(*promSink)

	s.lockHook = func(s *promSink) {
		s.lockHook = nil
		s.counters.Store("inner", prometheus.NewCounterVec(
			prometheus.CounterOpts{Name: "inner_total"}, []string{"v"}))
	}
	m.IncCounter("inner", "x")

	s.lockHook = func(s *promSink) {
		s.lockHook = nil
		s.histograms.Store("inner_h", prometheus.NewHistogramVec(
			prometheus.HistogramOpts{Name: "inner_h_seconds", Buckets: defaultBuckets}, []string{"v"}))
	}
	m.ObserveDuration("inner_h", 1, "x")

	s.lockHook = func(s *promSink) {
		s.lockHook = nil
		s.gauges.Store("inner_g", prometheus.NewGaugeVec(
			prometheus.GaugeOpts{Name: "inner_g"}, []string{"v"}))
	}
	m.SetGauge("inner_g", 1, "x")

	_, err := Gather(m)
	testx.RequireNoError(t, err)
}

// fakeRegisterer 是仅实现 prometheus.Registerer 的测试桩（不支持 Gather）。
type fakeRegisterer struct{}

func (fakeRegisterer) Register(prometheus.Collector) error  { return nil }
func (fakeRegisterer) MustRegister(...prometheus.Collector) {}
func (fakeRegisterer) Unregister(prometheus.Collector) bool { return true }

// fakeRegisterer 是只注册不 Gather 的注册表替身。
type fakeRegisterer struct{}

func (fakeRegisterer) Register(prometheus.Collector) error {
	return nil
}

func (fakeRegisterer) Unregister(prometheus.Collector) bool {
	return true
}

func (fakeRegisterer) MustRegister(...prometheus.Collector) {}

// FuzzMetrics 验证任意指标名/标签输入下 Prometheus 后端不 panic。
func FuzzMetrics(f *testing.F) {
	f.Add("dbx.queries", "a", "b")
	f.Add("", "x", "y")
	f.Fuzz(func(t *testing.T, name, l1, l2 string) {
		reg := prometheus.NewRegistry()
		m, err := metricsx.New(WithPrometheus(WithRegistry(reg)))
		testx.RequireNoError(t, err)
		m.IncCounter(name, l1, l2)
		m.ObserveDuration(name, 1.0, l1, l2)
		m.AddCounter(name, 1.0, l1, l2)
		m.AddGauge(name, 1.0, l1, l2)
		m.SetGauge(name, 1.0, l1, l2)
		_ = m.Register(name, "帮助", "a", "b")
	})
}
