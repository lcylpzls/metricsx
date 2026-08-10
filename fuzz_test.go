package metricsx

import (
	"testing"

	testx "github.com/lcylpzls/testx"
)

// FuzzMetrics 验证任意指标名/标签输入下核心操作不 panic。
func FuzzMetrics(f *testing.F) {
	f.Add("dbx.queries", "a", "b")
	f.Add("", "x", "y")
	f.Fuzz(func(t *testing.T, name, l1, l2 string) {
		m, err := New()
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
