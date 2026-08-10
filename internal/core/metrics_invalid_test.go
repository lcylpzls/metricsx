package core

import "testing"

// TestInvalidLabels 覆盖各指标方法的无效标签分支。
func TestInvalidLabels(t *testing.T) {
	m, err := New()
	if err != nil {
		t.Fatalf("New 失败：%v", err)
	}
	// 非法 UTF-8 标签应被拒绝（validLabels 校验失败）。
	bad := string([]byte{0xff})
	m.ObserveDuration("x", 1, bad)
	m.AddCounter("x", 1, bad)
	m.AddGauge("x", 1, bad)
	m.SetGauge("x", 1, bad)
	m.IncCounter("x", bad)
}
