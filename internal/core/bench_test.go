package core

import (
	"testing"

	testx "github.com/lcylpzls/testx"
)

// BenchmarkIncCounterHit 基准：内存后端计数器命中。
func BenchmarkIncCounterHit(b *testing.B) {
	m, err := New()
	testx.RequireNoError(b, err)
	m.IncCounter("bench", "v")
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		m.IncCounter("bench", "v")
	}
}

// BenchmarkObserveDurationHit 基准：内存后端直方图命中。
func BenchmarkObserveDurationHit(b *testing.B) {
	m, err := New()
	testx.RequireNoError(b, err)
	m.ObserveDuration("bench_dur", 0.1, "v")
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		m.ObserveDuration("bench_dur", 0.1, "v")
	}
}

// BenchmarkAddCounterHit 基准：内存后端增量计数命中。
func BenchmarkAddCounterHit(b *testing.B) {
	m, err := New()
	testx.RequireNoError(b, err)
	m.AddCounter("bench_add", 1, "v")
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		m.AddCounter("bench_add", 1, "v")
	}
}

// BenchmarkSetGaugeHit 基准：内存后端瞬时量命中。
func BenchmarkSetGaugeHit(b *testing.B) {
	m, err := New()
	testx.RequireNoError(b, err)
	m.SetGauge("bench_gauge", 1, "v")
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		m.SetGauge("bench_gauge", 1, "v")
	}
}
