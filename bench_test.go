package metricsx

import (
	testx "github.com/lcylpzls/testx"
	"testing"

	"github.com/prometheus/client_golang/prometheus"
)

// BenchmarkIncCounterHit 基准:已懒创建的计数器命中。
func BenchmarkIncCounterHit(b *testing.B) {
	m, err := New(WithRegistry(prometheus.NewRegistry()))
	testx.RequireNoError(b, err)

	m.IncCounter("bench", "v")
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		m.IncCounter("bench", "v")
	}
}

// BenchmarkObserveDurationHit 基准:已懒创建的直方图命中。
func BenchmarkObserveDurationHit(b *testing.B) {
	m, err := New(WithRegistry(prometheus.NewRegistry()))
	testx.RequireNoError(b, err)

	m.ObserveDuration("bench_dur", 0.1, "v")
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		m.ObserveDuration("bench_dur", 0.1, "v")
	}
}

// BenchmarkAddCounterHit 基准:已懒创建的增量计数命中。
func BenchmarkAddCounterHit(b *testing.B) {
	m, err := New(WithRegistry(prometheus.NewRegistry()))
	testx.RequireNoError(b, err)

	m.AddCounter("bench_add", 1, "v")
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		m.AddCounter("bench_add", 1, "v")
	}
}

// BenchmarkAddGaugeHit 基准:已懒创建的瞬时量增量命中。
func BenchmarkAddGaugeHit(b *testing.B) {
	m, err := New(WithRegistry(prometheus.NewRegistry()))
	testx.RequireNoError(b, err)

	m.AddGauge("bench_gauge", 1, "v")
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		m.AddGauge("bench_gauge", 1, "v")
	}
}

// BenchmarkSetGaugeHit 基准:已懒创建的瞬时量设置命中。
func BenchmarkSetGaugeHit(b *testing.B) {
	m, err := New(WithRegistry(prometheus.NewRegistry()))
	testx.RequireNoError(b, err)

	m.SetGauge("bench_set", 1, "v")
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		m.SetGauge("bench_set", 1, "v")
	}
}

// BenchmarkDirectClient 基准:直接使用 client_golang 的 Vec(对照组)。
func BenchmarkDirectClient(b *testing.B) {
	reg := prometheus.NewRegistry()
	vec := prometheus.NewCounterVec(prometheus.CounterOpts{
		Name: "direct_total",
	}, []string{"label0"})
	reg.MustRegister(vec)
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		vec.WithLabelValues("v").Inc()
	}
}
