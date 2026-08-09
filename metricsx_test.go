package metricsx

import (
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
	if err != nil {
		t.Fatalf("New 失败:%v", err)
	}
	return m, reg
}

// gatherFamily 从注册表抓取指定指标族。
func gatherFamily(t *testing.T, reg *prometheus.Registry, name string) *dto.MetricFamily {
	t.Helper()
	families, err := reg.Gather()
	if err != nil {
		t.Fatalf("Gather 失败:%v", err)
	}
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
	if err != nil {
		t.Fatal(err)
	}
	if m.Registry() != prometheus.DefaultRegisterer {
		t.Error("默认应为 DefaultRegisterer")
	}
	// 实际写入默认注册表
	m.IncCounter("default_test", "v")
	families, err := prometheus.DefaultGatherer.Gather()
	if err != nil {
		t.Fatal(err)
	}
	found := false
	for _, f := range families {
		if f.GetName() == "default_test_total" {
			found = true
		}
	}
	if !found {
		t.Error("默认注册表应包含写入的指标")
	}
}

func TestIncCounterLazy(t *testing.T) {
	m, reg := newTestMetrics(t, WithNamespace("myapp"))
	m.IncCounter("dbx.queries", "select")
	m.IncCounter("dbx.queries", "select")
	m.IncCounter("dbx.queries", "insert")

	f := gatherFamily(t, reg, "myapp_dbx_queries_total")
	if f == nil {
		t.Fatal("指标族不存在")
	}
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
	if f == nil {
		t.Fatal("直方图指标族不存在")
	}
	hist := f.GetMetric()[0].GetHistogram()
	if hist.GetSampleCount() != 2 {
		t.Errorf("样本数 = %d,want 2", hist.GetSampleCount())
	}
}

func TestRegisterWithLabels(t *testing.T) {
	m, reg := newTestMetrics(t)
	if err := m.Register("dbx.queries", "数据库查询次数", "op"); err != nil {
		t.Fatalf("注册失败:%v", err)
	}
	m.IncCounter("dbx.queries", "select")
	f := gatherFamily(t, reg, "dbx_queries_total")
	if f == nil {
		t.Fatal("指标族不存在")
	}
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
	if err == nil {
		t.Fatal("重复注册应报错")
	}
	if code, _ := errx.CodeOf(err); code != CodeAlreadyRegistered {
		t.Errorf("错误码 = %s,want %s", code, CodeAlreadyRegistered)
	}
}

func TestRegisterEmptyName(t *testing.T) {
	m, _ := newTestMetrics(t)
	err := m.Register("", "帮助")
	if err == nil {
		t.Fatal("空指标名应报错")
	}
	if code, _ := errx.CodeOf(err); code != CodeInvalidConfig {
		t.Errorf("错误码 = %s,want %s", code, CodeInvalidConfig)
	}
}

func TestRegisterAfterLazyCreated(t *testing.T) {
	m, _ := newTestMetrics(t)
	m.IncCounter("already", "v")
	err := m.Register("already", "帮助", "op")
	if err == nil {
		t.Fatal("已懒创建的指标再注册应报错")
	}
	if code, _ := errx.CodeOf(err); code != CodeAlreadyRegistered {
		t.Errorf("错误码 = %s,want %s", code, CodeAlreadyRegistered)
	}
	m.ObserveDuration("already_dur", 1.0, "v")
	err = m.Register("already_dur", "帮助", "op")
	if err == nil {
		t.Fatal("已懒创建的直方图再注册应报错")
	}
	if code, _ := errx.CodeOf(err); code != CodeAlreadyRegistered {
		t.Errorf("错误码 = %s,want %s", code, CodeAlreadyRegistered)
	}
}

func TestVersion(t *testing.T) {
	if Version != "v0.3.0" {
		t.Errorf("Version = %s,want v0.3.0", Version)
	}
}

func TestGather(t *testing.T) {
	m, _ := newTestMetrics(t)
	m.IncCounter("gather_test", "v")
	families, err := m.Gather()
	if err != nil {
		t.Fatalf("Gather 失败:%v", err)
	}
	found := false
	for _, f := range families {
		if f.GetName() == "gather_test_total" {
			found = true
		}
	}
	if !found {
		t.Error("Gather 应包含写入的指标")
	}
}

func TestGatherDefaultRegistry(t *testing.T) {
	m, err := New()
	if err != nil {
		t.Fatal(err)
	}
	m.IncCounter("gather_default", "v")
	families, err := m.Gather()
	if err != nil {
		t.Fatalf("默认注册表 Gather 失败:%v", err)
	}
	found := false
	for _, f := range families {
		if f.GetName() == "gather_default_total" {
			found = true
		}
	}
	if !found {
		t.Error("默认注册表 Gather 应包含写入的指标")
	}
}

// registererOnly 只实现 Registerer 不实现 Gatherer。
type registererOnly struct{}

func (registererOnly) Register(prometheus.Collector) error  { return nil }
func (registererOnly) MustRegister(...prometheus.Collector) {}
func (registererOnly) Unregister(prometheus.Collector) bool { return true }

func TestGatherUnsupportedRegistry(t *testing.T) {
	m, err := New(WithRegistry(registererOnly{}))
	if err != nil {
		t.Fatal(err)
	}
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
	if f == nil {
		t.Fatal("占位标签指标应存在")
	}
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
	if f := gatherFamily(t, reg, "bad_total"); f != nil {
		t.Error("非法标签不应产生指标")
	}
	if f := gatherFamily(t, reg, "bad_dur_seconds"); f != nil {
		t.Error("非法标签不应产生直方图")
	}
	// 合法标签正常
	m.IncCounter("ok", "v")
	if f := gatherFamily(t, reg, "ok_total"); f == nil {
		t.Error("合法标签应产生指标")
	}
}

// FuzzMetrics 保证任意名称与标签组合不 panic。
func FuzzMetrics(f *testing.F) {
	f.Add("dbx.queries", "a", "b")
	f.Add("", "x", "y")
	f.Fuzz(func(t *testing.T, name, l1, l2 string) {
		m, err := New(WithRegistry(prometheus.NewRegistry()))
		if err != nil {
			t.Fatal(err)
		}
		m.IncCounter(name, l1, l2)
		m.ObserveDuration(name, 1.0, l1, l2)
		_ = m.Register(name, "帮助", "a", "b")
		_ = m.Register(name, "帮助2", "x")
	})
}
