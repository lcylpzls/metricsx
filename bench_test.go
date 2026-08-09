package metricsx

import (
	"testing"

	"github.com/prometheus/client_golang/prometheus"
)

// BenchmarkIncCounterHit 基准:已懒创建的计数器命中。
func BenchmarkIncCounterHit(b *testing.B) {
	m, err := New(WithRegistry(prometheus.NewRegistry()))
	if err != nil {
		b.Fatal(err)
	}
	m.IncCounter("bench", "v")
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		m.IncCounter("bench", "v")
	}
}

// BenchmarkObserveDurationHit 基准:已懒创建的直方图命中。
func BenchmarkObserveDurationHit(b *testing.B) {
	m, err := New(WithRegistry(prometheus.NewRegistry()))
	if err != nil {
		b.Fatal(err)
	}
	m.ObserveDuration("bench_dur", 0.1, "v")
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		m.ObserveDuration("bench_dur", 0.1, "v")
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
