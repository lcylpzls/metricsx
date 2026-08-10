package jobx

import (
	"errors"
	testx "github.com/lcylpzls/testx"
	"testing"
	"time"

	"github.com/lcylpzls/errx"
	"github.com/lcylpzls/metricsx"
	prometheusx "github.com/lcylpzls/metricsx/prometheus"
	"github.com/prometheus/client_golang/prometheus"
)

func TestNewCallbacks(t *testing.T) {
	m, err := metricsx.New(prometheusx.WithPrometheus(prometheusx.WithRegistry(prometheus.NewRegistry())))
	testx.RequireNoError(t, err)

	cb := New(m)
	cb.Queued("task", 1)
	cb.Running("task", -1)
	cb.Completed("task", 250*time.Millisecond)
	cb.Failed("task", errors.New("boom"))
	cb.Failed("task", errx.NewCode("JOB_FAILED", "结构化失败"))
	cb.Retried("task", 2)
	cb.Dropped("task")
	cb.Skipped("task")
	cb.Replaced("task")

	families, err := prometheusx.Gather(m)
	testx.RequireNoError(t, err)

	names := map[string]bool{}
	for _, f := range families {
		names[f.GetName()] = true
	}
	for _, want := range []string{
		"jobx_queued",
		"jobx_running",
		"jobx_completed_total",
		"jobx_completed_duration_seconds",
		"jobx_failed_total",
		"jobx_retried_total",
		"jobx_dropped_total",
		"jobx_skipped_total",
		"jobx_replaced_total",
	} {
		if !names[want] {
			t.Errorf("缺少指标族：%s", want)
		}
	}
}
