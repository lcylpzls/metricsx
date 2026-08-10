package idgenx

import (
	"errors"
	testx "github.com/lcylpzls/testx"
	"testing"

	"github.com/lcylpzls/errx"
	"github.com/lcylpzls/metricsx"
	prometheusx "github.com/lcylpzls/metricsx/prometheus"
	"github.com/prometheus/client_golang/prometheus"
)

func TestNewCallbacks(t *testing.T) {
	m, err := metricsx.New(prometheusx.WithPrometheus(prometheusx.WithRegistry(prometheus.NewRegistry())))
	testx.RequireNoError(t, err)

	cb := New(m)
	cb.Generated(7, 1)
	cb.Rejected(7, errors.New("backward"))
	cb.Rejected(8, errx.NewCode("CLOCK_BACKWARD", "时钟回拨"))
	cb.WaitMS(7, 500)

	families, err := prometheusx.Gather(m)
	testx.RequireNoError(t, err)

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
