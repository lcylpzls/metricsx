package updatex

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
	cb.CheckTotal(2)
	cb.CheckFailures(errors.New("network"))
	cb.CheckFailures(errx.NewCode("CHECK_FAILED", "检查失败"))
	cb.UpdateSuccess("v1.0.0")
	cb.UpdateFailures(errors.New("verify"))
	cb.UpdateFailures(errx.NewCode("UPDATE_FAILED", "更新失败"))

	families, err := prometheusx.Gather(m)
	testx.RequireNoError(t, err)

	names := map[string]bool{}
	for _, f := range families {
		names[f.GetName()] = true
	}
	for _, want := range []string{
		"updatex_checks_total",
		"updatex_check_failures_total",
		"updatex_updates_total",
	} {
		if !names[want] {
			t.Errorf("缺少指标族：%s", want)
		}
	}
}
