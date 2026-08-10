package metricsx_test

import (
	"testing"

	"github.com/lcylpzls/metricsx"
)

// TestPublicAPI 黑盒冒烟测试：覆盖根包全部转发函数、类型别名与常量。
func TestPublicAPI(t *testing.T) {
	if metricsx.Version != "v1.6.0" {
		t.Fatalf("Version = %s", metricsx.Version)
	}

	m, err := metricsx.New()
	if err != nil || m == nil {
		t.Fatalf("New 失败：%v", err)
	}
	m2, err := metricsx.New(metricsx.WithSink(m.Sink()))
	if err != nil || m2 == nil {
		t.Fatalf("New(WithSink) 失败：%v", err)
	}
	m.IncCounter("smoke.counter", "a", "b")
	m.AddCounter("smoke.counter", 1, "a", "b")
	m.ObserveDuration("smoke.duration", 0.1, "a")
	m.AddGauge("smoke.gauge", 1)
	m.SetGauge("smoke.gauge", 2)
	_ = m.Register("smoke.metric", "说明", "label")
	_ = m.Sink()
	_, _ = m.Snapshot()

	var _ metricsx.Option
	var _ metricsx.Sink
	var _ metricsx.SnapshotMetric
	var _ metricsx.Snapshot
	var _ metricsx.Snapshotter
	_ = metricsx.CodeInvalidConfig
	_ = metricsx.CodeAlreadyRegistered
}
