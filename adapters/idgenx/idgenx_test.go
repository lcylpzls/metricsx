package idgenx

import (
	"errors"
	"testing"

	"github.com/lcylpzls/errx"
	"github.com/lcylpzls/metricsx"
	"github.com/prometheus/client_golang/prometheus"
)

func TestNewCallbacks(t *testing.T) {
	m, err := metricsx.New(metricsx.WithRegistry(prometheus.NewRegistry()))
	if err != nil {
		t.Fatal(err)
	}
	cb := New(m)
	cb.Generated(7, 1)
	cb.Rejected(7, errors.New("backward"))
	cb.Rejected(8, errx.NewCode("CLOCK_BACKWARD", "时钟回拨"))
	cb.WaitMS(7, 500)

	families, err := m.Gather()
	if err != nil {
		t.Fatal(err)
	}
	names := map[string]bool{}
	for _, f := range families {
		names[f.GetName()] = true
	}
	for _, want := range []string{
		"idgenx_generated_total",
		"idgenx_rejected_total",
		"idgenx_wait_duration_seconds",
	} {
		if !names[want] {
			t.Errorf("缺少指标族：%s", want)
		}
	}
}
