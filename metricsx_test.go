package metricsx

import (
	testx "github.com/lcylpzls/testx"
	"sync"
	"testing"

	"github.com/lcylpzls/errx"
	"github.com/prometheus/client_golang/prometheus"
	dto "github.com/prometheus/client_model/go"
)

func newTestMetrics(t *testing.T, opts ...Option) (*Metrics, *prometheus.Registry) {
	t.Helper()
	reg := prometheus.NewRegistry()
	all := append([]Option{WithRegistry(reg)}, opts...)
	m, err := New(all...)
	testx.RequireNoError(t, err)

	return m, reg
}

// gatherFamily 从注册表抓取指定指标族。
func gatherFamily(t *testing.T, reg *prometheus.Registry, name string) *dto.MetricFamily {
	t.Helper()
	families, err := reg.Gather()
	testx.RequireNoError(t, err)

	for _, f := range families {
		if f.GetName() == name {
			return f
		}
	}
	return nil
}

func TestNewInvalidConfig(t *testing.T) {
	cases := []Option{
		WithNamespace("bad.ns"),
		WithNamespace("bad ns"),
		WithRegistry(nil),
		WithBuckets([]float64{1, 0, 2}),
		WithBuckets([]float64{-1}),
	}
	for _, opt := range cases {
		if _, err := New(opt); err == nil {
			t.Errorf("配置应非法:%v", opt)
		} else if code, _ := errx.CodeOf(err); code != CodeInvalidConfig {
			t.Errorf("错误码 = %s,want %s", code, CodeInvalidConfig)
		}
	}
	// 默认配置合法
	if _, err := New(); err != nil {
		t.Fatalf("默认配置应合法:%v", err)
	}
	// nil 选项忽略
	if _, err := New(nil); err != nil {
		t.Fatalf("nil 选项应忽略:%v", err)
	}
}

func TestDefaultRegistry(t *testing.T) {
	m, err := New()
	testx.RequireNoError(t, err)

	if m.Registry() != prometheus.DefaultRegisterer {
		t.Error("默认应为 DefaultRegisterer")
	}
	// 实际写入默认注册表
	m.IncCounter("default_test", "v")
	families, err := prometheus.DefaultGatherer.Gather()
	testx.RequireNoError(t, err)

	found := false
	for _, f := range families {
		if f.GetName() == "default_test_total" {
			found = true
		}
	}
	testx.True(t, found)

}

func TestIncCounterLazy(t *testing.T) {
	m, reg := newTestMetrics(t, WithNamespace("myapp"))
	m.IncCounter("dbx.queries", "select")
	m.IncCounter("dbx.queries", "select")
	m.IncCounter("dbx.queries", "insert")

	f := gatherFamily(t, reg, "myapp_dbx_queries_total")
	testx.RequireNotNil(t, f)

	if len(f.GetMetric()) != 2 {
		t.Fatalf("序列数 = %d,want 2", len(f.GetMetric()))
	}
	values := map[string]float64{}
	for _, metric := range f.GetMetric() {
		label := metric.GetLabel()[0].GetValue()
		values[label] = metric.GetCounter().GetValue()
	}
	if values["select"] != 2 || values["insert"] != 1 {
		t.Errorf("计数不符:%v", values)
	}
}

func TestObserveDurationLazy(t *testing.T) {
	m, reg := newTestMetrics(t, WithNamespace("myapp"))
	m.ObserveDuration("httpx.duration", 0.5, "GET")
	m.ObserveDuration("httpx.duration", 1.5, "GET")

	f := gatherFamily(t, reg, "myapp_httpx_duration_seconds")
	testx.RequireNotNil(t, f)

	hist := f.GetMetric()[0].GetHistogram()
	if hist.GetSampleCount() != 2 {
		t.Errorf("样本数 = %d,want 2", hist.GetSampleCount())
	}
}

func TestAddCounterDelta(t *testing.T) {
	m, reg := newTestMetrics(t, WithNamespace("myapp"))
	m.AddCounter("filex.bytes", 1024, "bucket", "put")
	m.AddCounter("filex.bytes", 512, "bucket", "put")
	f := gatherFamily(t, reg, "myapp_filex_bytes_total")
	testx.RequireNotNil(t, f)

	if got := f.GetMetric()[0].GetCounter().GetValue(); got != 1536 {
		t.Errorf("累加值 = %v,want 1536", got)
	}
	if got := f.GetMetric()[0].GetLabel()[1].GetValue(); got != "put" {
		t.Errorf("标签值 = %q,want put", got)
	}
}

func TestAddCounterLabelMismatchIgnored(t *testing.T) {
	m, reg := newTestMetrics(t)
	if err := m.Register("ac", "帮助", "a", "b"); err != nil {
		t.Fatal(err)
	}
	m.AddCounter("ac", 1.0, "only-one")
	if f := gatherFamily(t, reg, "ac_total"); f != nil {
		t.Error("标签不匹配不应产生增量计数")
	}
}

func TestAddCounterRegisterFailureIgnored(t *testing.T) {
	m, reg := newTestMetrics(t)
	existing := prometheus.NewCounterVec(prometheus.CounterOpts{Name: "manual_ac_total"}, nil)
	reg.MustRegister(existing)
	existing.WithLabelValues().Add(5)
	m.AddCounter("manual_ac", 1.0)
	if f := gatherFamily(t, reg, "manual_ac_total"); f == nil ||
		f.GetMetric()[0].GetCounter().GetValue() != 5 {
		t.Error("手动注册的计数器应保留且不被覆盖")
	}
}

func TestAddGaugeDelta(t *testing.T) {
	m, reg := newTestMetrics(t, WithNamespace("myapp"))
	m.AddGauge("webx.inflight", 1)
	m.AddGauge("webx.inflight", 1)
	m.AddGauge("webx.inflight", -1)
	f := gatherFamily(t, reg, "myapp_webx_inflight")
	testx.RequireNotNil(t, f)

	if got := f.GetMetric()[0].GetGauge().GetValue(); got != 1 {
		t.Errorf("瞬时值 = %v,want 1", got)
	}
}

func TestLazyCreateInnerCheck(t *testing.T) {
	old := lockHook
	defer func() { lockHook = old }()

	t.Run("counter", func(t *testing.T) {
		m, _ := newTestMetrics(t)
		sentinel := prometheus.NewCounterVec(prometheus.CounterOpts{Name: "inner_c"}, nil)
		lockHook = func(x *Metrics) { x.counters.Store("inner_c", sentinel) }
		if got := m.counterVec("inner_c", 0); got != sentinel {
			t.Error("counter 二次命中分支未返回哨兵向量")
		}
	})
	t.Run("histogram", func(t *testing.T) {
		m, _ := newTestMetrics(t)
		sentinel := prometheus.NewHistogramVec(prometheus.HistogramOpts{Name: "inner_h"}, nil)
		lockHook = func(x *Metrics) { x.histograms.Store("inner_h", sentinel) }
		if got := m.histogramVec("inner_h", 0); got != sentinel {
			t.Error("histogram 二次命中分支未返回哨兵向量")
		}
	})
	t.Run("gauge", func(t *testing.T) {
		m, _ := newTestMetrics(t)
		sentinel := prometheus.NewGaugeVec(prometheus.GaugeOpts{Name: "inner_g"}, nil)
		lockHook = func(x *Metrics) { x.gauges.Store("inner_g", sentinel) }
		if got := m.gaugeVec("inner_g", 0); got != sentinel {
			t.Error("gauge 二次命中分支未返回哨兵向量")
		}
	})
}

func TestSetGaugeValue(t *testing.T) {
	m, reg := newTestMetrics(t, WithNamespace("myapp"))
	m.SetGauge("webx.connections", 7)
	m.SetGauge("webx.connections", 3)
	f := gatherFamily(t, reg, "myapp_webx_connections")
	testx.RequireNotNil(t, f)

	if got := f.GetMetric()[0].GetGauge().GetValue(); got != 3 {
		t.Errorf("瞬时值 = %v,want 3", got)
	}
}

func TestRegisterWithLabels(t *testing.T) {
	m, reg := newTestMetrics(t)
	if err := m.Register("dbx.queries", "数据库查询次数", "op"); err != nil {
		t.Fatalf("注册失败:%v", err)
	}
	m.IncCounter("dbx.queries", "select")
	f := gatherFamily(t, reg, "dbx_queries_total")
	testx.RequireNotNil(t, f)

	if got := f.GetMetric()[0].GetLabel()[0].GetName(); got != "op" {
		t.Errorf("标签键名 = %q,want op", got)
	}
	if got := f.GetHelp(); got != "数据库查询次数" {
		t.Errorf("帮助文本 = %q", got)
	}
}

func TestRegisterDuplicate(t *testing.T) {
	m, _ := newTestMetrics(t)
	if err := m.Register("name", "帮助"); err != nil {
		t.Fatal(err)
	}
	err := m.Register("name", "帮助2")
	testx.RequireError(t, err)

	if code, _ := errx.CodeOf(err); code != CodeAlreadyRegistered {
		t.Errorf("错误码 = %s,want %s", code, CodeAlreadyRegistered)
	}
}

func TestRegisterEmptyName(t *testing.T) {
	m, _ := newTestMetrics(t)
	err := m.Register("", "帮助")
	testx.RequireError(t, err)

	if code, _ := errx.CodeOf(err); code != CodeInvalidConfig {
		t.Errorf("错误码 = %s,want %s", code, CodeInvalidConfig)
	}
}

func TestRegisterAfterLazyCreated(t *testing.T) {
	m, _ := newTestMetrics(t)
	m.IncCounter("already", "v")
	err := m.Register("already", "帮助", "op")
	testx.RequireError(t, err)

	if code, _ := errx.CodeOf(err); code != CodeAlreadyRegistered {
		t.Errorf("错误码 = %s,want %s", code, CodeAlreadyRegistered)
	}
	m.ObserveDuration("already_dur", 1.0, "v")
	err = m.Register("already_dur", "帮助", "op")
	testx.RequireError(t, err)

	if code, _ := errx.CodeOf(err); code != CodeAlreadyRegistered {
		t.Errorf("错误码 = %s,want %s", code, CodeAlreadyRegistered)
	}
	m.AddGauge("already_gauge", 1.0, "v")
	err = m.Register("already_gauge", "帮助", "op")
	testx.RequireError(t, err)

	if code, _ := errx.CodeOf(err); code != CodeAlreadyRegistered {
		t.Errorf("错误码 = %s,want %s", code, CodeAlreadyRegistered)
	}
}

func TestVersion(t *testing.T) {
	testx.Equal(t, Version, "v1.3.0")

}

func TestGather(t *testing.T) {
	m, _ := newTestMetrics(t)
	m.IncCounter("gather_test", "v")
	families, err := m.Gather()
	testx.RequireNoError(t, err)

	found := false
	for _, f := range families {
		if f.GetName() == "gather_test_total" {
			found = true
		}
	}
	testx.True(t, found)

}

func TestGatherDefaultRegistry(t *testing.T) {
	m, err := New()
	testx.RequireNoError(t, err)

	m.IncCounter("gather_default", "v")
	families, err := m.Gather()
	testx.RequireNoError(t, err)

	found := false
	for _, f := range families {
		if f.GetName() == "gather_default_total" {
			found = true
		}
	}
	testx.True(t, found)

}

// registererOnly 只实现 Registerer 不实现 Gatherer。
type registererOnly struct{}

func (registererOnly) Register(prometheus.Collector) error  { return nil }
func (registererOnly) MustRegister(...prometheus.Collector) {}
func (registererOnly) Unregister(prometheus.Collector) bool { return true }

func TestGatherUnsupportedRegistry(t *testing.T) {
	m, err := New(WithRegistry(registererOnly{}))
	testx.RequireNoError(t, err)

	if _, err := m.Gather(); err == nil {
		t.Fatal("不支持的注册表 Gather 应报错")
	}
}

func TestLabelMismatchIgnored(t *testing.T) {
	m, reg := newTestMetrics(t)
	if err := m.Register("two", "帮助", "a", "b"); err != nil {
		t.Fatal(err)
	}
	// 注册 2 键但只传 1 个值:静默忽略,不 panic。
	m.IncCounter("two", "only-one")
	if f := gatherFamily(t, reg, "two_total"); f != nil {
		t.Error("标签不匹配不应产生指标")
	}
	// 未注册指标的占位标签:任意数量都可用
	m.IncCounter("auto", "a", "b", "c")
	f := gatherFamily(t, reg, "auto_total")
	testx.RequireNotNil(t, f)

	if len(f.GetMetric()[0].GetLabel()) != 3 {
		t.Errorf("占位标签数 = %d,want 3", len(f.GetMetric()[0].GetLabel()))
	}
}

func TestSanitizeName(t *testing.T) {
	cases := []struct {
		in, want string
	}{
		{"dbx.queries", "dbx_queries"},
		{"resx.limiter.accepted", "resx_limiter_accepted"},
		{"valid", "valid"},
		{"a-b", "a_b"},
		{"!!!", "___"},
		{"a:b", "a:b"},
	}
	for _, c := range cases {
		if got := sanitizeName(c.in); got != c.want {
			t.Errorf("sanitizeName(%q) = %q,want %q", c.in, got, c.want)
		}
	}
}

func TestRegisterFailureIgnored(t *testing.T) {
	m, _ := newTestMetrics(t)
	// 注册一个与预注册指标名冲突的 collector,验证注册失败被静默忽略。
	m.Register("conflict", "帮助")
	// 同 Prometheus 名下已有 Vec 时,再次懒创建会注册失败 → 静默忽略。
	_ = m
	// 直接构造冲突:手动注册同名 Vec 到 registry,再触发懒创建。
	m2, reg := newTestMetrics(t)
	existing := prometheus.NewCounterVec(prometheus.CounterOpts{Name: "manual_total"}, nil)
	reg.MustRegister(existing)
	existing.WithLabelValues().Inc()
	m2.IncCounter("manual") // 懒创建注册失败,静默忽略
	if f := gatherFamily(t, reg, "manual_total"); f == nil ||
		f.GetMetric()[0].GetCounter().GetValue() != 1 {
		t.Error("手动注册的指标应保留且不被覆盖")
	}
}

func TestConcurrentIncCounter(t *testing.T) {
	m, reg := newTestMetrics(t)
	start := make(chan struct{})
	var wg sync.WaitGroup
	for i := 0; i < 8; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			<-start
			for j := 0; j < 500; j++ {
				m.IncCounter("hot", "v")
				m.ObserveDuration("hot_dur", 0.1, "v")
			}
		}()
	}
	close(start)
	wg.Wait()
	f := gatherFamily(t, reg, "hot_total")
	if f == nil || f.GetMetric()[0].GetCounter().GetValue() != 4000 {
		t.Errorf("并发计数不符:%v", f)
	}
}

func TestConcurrentObserveDuration(t *testing.T) {
	m, reg := newTestMetrics(t)
	start := make(chan struct{})
	var wg sync.WaitGroup
	for i := 0; i < 8; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			<-start
			for j := 0; j < 500; j++ {
				m.ObserveDuration("dur_only", 0.1, "v")
			}
		}()
	}
	close(start)
	wg.Wait()
	f := gatherFamily(t, reg, "dur_only_seconds")
	if f == nil || f.GetMetric()[0].GetHistogram().GetSampleCount() != 4000 {
		t.Errorf("并发直方图计数不符")
	}
}

func TestObserveDurationLabelMismatch(t *testing.T) {
	m, reg := newTestMetrics(t)
	if err := m.Register("dur", "帮助", "a", "b"); err != nil {
		t.Fatal(err)
	}
	m.ObserveDuration("dur", 0.5, "only-one")
	if f := gatherFamily(t, reg, "dur_seconds"); f != nil {
		t.Error("标签不匹配不应产生直方图")
	}
}

func TestHistogramRegisterFailureIgnored(t *testing.T) {
	m, reg := newTestMetrics(t)
	existing := prometheus.NewHistogramVec(
		prometheus.HistogramOpts{Name: "manual_dur_seconds"}, nil)
	reg.MustRegister(existing)
	existing.WithLabelValues().Observe(1)
	m.ObserveDuration("manual_dur", 0.5)
	if f := gatherFamily(t, reg, "manual_dur_seconds"); f == nil ||
		f.GetMetric()[0].GetHistogram().GetSampleCount() != 1 {
		t.Error("手动注册的直方图应保留")
	}
}

func TestErrorCodesRegistered(t *testing.T) {
	if errx.Describe(CodeInvalidConfig) == "" || errx.Describe(CodeAlreadyRegistered) == "" {
		t.Error("错误码未注册")
	}
}

func TestInvalidUTF8LabelsIgnored(t *testing.T) {
	m, reg := newTestMetrics(t)
	m.IncCounter("bad", "\xa4") // 非法 UTF-8,应静默忽略不 panic
	m.ObserveDuration("bad_dur", 1.0, "\xa4")
	m.AddCounter("bad_add", 1.0, "\xa4")
	m.AddGauge("bad_gauge", 1.0, "\xa4")
	m.SetGauge("bad_set", 1.0, "\xa4")
	if f := gatherFamily(t, reg, "bad_total"); f != nil {
		t.Error("非法标签不应产生指标")
	}
	if f := gatherFamily(t, reg, "bad_dur_seconds"); f != nil {
		t.Error("非法标签不应产生直方图")
	}
	if f := gatherFamily(t, reg, "bad_add_total"); f != nil {
		t.Error("非法标签不应产生增量计数")
	}
	if f := gatherFamily(t, reg, "bad_gauge"); f != nil {
		t.Error("非法标签不应产生瞬时量")
	}
	// 合法标签正常
	m.IncCounter("ok", "v")
	if f := gatherFamily(t, reg, "ok_total"); f == nil {
		t.Error("合法标签应产生指标")
	}
}

func TestGaugeLabelMismatchIgnored(t *testing.T) {
	m, reg := newTestMetrics(t)
	if err := m.Register("g", "帮助", "a", "b"); err != nil {
		t.Fatal(err)
	}
	m.AddGauge("g", 1.0, "only-one")
	m.SetGauge("g", 1.0, "only-one")
	if f := gatherFamily(t, reg, "g"); f != nil {
		t.Error("标签不匹配不应产生瞬时量")
	}
}

func TestGaugeRegisterFailureIgnored(t *testing.T) {
	m, reg := newTestMetrics(t)
	existing := prometheus.NewGaugeVec(prometheus.GaugeOpts{Name: "manual_gauge"}, nil)
	reg.MustRegister(existing)
	existing.WithLabelValues().Set(9)
	m.SetGauge("manual_gauge", 1.0)
	if f := gatherFamily(t, reg, "manual_gauge"); f == nil ||
		f.GetMetric()[0].GetGauge().GetValue() != 9 {
		t.Error("手动注册的瞬时量应保留且不被覆盖")
	}
}

func TestConcurrentGauge(t *testing.T) {
	m, reg := newTestMetrics(t)
	start := make(chan struct{})
	var wg sync.WaitGroup
	for i := 0; i < 8; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			<-start
			for j := 0; j < 500; j++ {
				m.AddGauge("hot_gauge", 1, "v")
				m.SetGauge("hot_set", 1, "v")
			}
		}()
	}
	close(start)
	wg.Wait()
	if f := gatherFamily(t, reg, "hot_gauge"); f == nil ||
		f.GetMetric()[0].GetGauge().GetValue() != 4000 {
		t.Errorf("并发瞬时量计数不符:%v", f)
	}
	if f := gatherFamily(t, reg, "hot_set"); f == nil ||
		f.GetMetric()[0].GetGauge().GetValue() != 1 {
		t.Errorf("并发设置瞬时量不符:%v", f)
	}
}

// FuzzMetrics 保证任意名称与标签组合不 panic。
func FuzzMetrics(f *testing.F) {
	f.Add("dbx.queries", "a", "b")
	f.Add("", "x", "y")
	f.Fuzz(func(t *testing.T, name, l1, l2 string) {
		m, err := New(WithRegistry(prometheus.NewRegistry()))
		testx.RequireNoError(t, err)

		m.IncCounter(name, l1, l2)
		m.ObserveDuration(name, 1.0, l1, l2)
		m.AddCounter(name, 1.0, l1, l2)
		m.AddGauge(name, 1.0, l1, l2)
		m.SetGauge(name, 1.0, l1, l2)
		_ = m.Register(name, "帮助", "a", "b")
		_ = m.Register(name, "帮助2", "x")
	})
}
