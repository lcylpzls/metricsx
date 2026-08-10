package filex

import (
	testx "github.com/lcylpzls/testx"
	"testing"

	"github.com/lcylpzls/metricsx"
	"github.com/prometheus/client_golang/prometheus"
)

func TestHookForward(t *testing.T) {
	m, err := metricsx.New(metricsx.WithRegistry(prometheus.NewRegistry()))
	testx.RequireNoError(t, err)

	h := New(m)
	h.Add("bucket-a", "put", 1024)
	h.IncError("bucket-a", "STORAGE_FAILED")

	families, err := m.Gather()
	testx.RequireNoError(t, err)

	got := map[string]float64{}
	for _, f := range families {
		switch f.GetName() {
		case "filex_operations_total", "filex_operation_bytes_total", "filex_errors_total":
			metric := f.GetMetric()[0]
			switch f.GetName() {
			case "filex_operations_total", "filex_errors_total":
				got[f.GetName()] = metric.GetCounter().GetValue()
			case "filex_operation_bytes_total":
				got[f.GetName()] = metric.GetCounter().GetValue()
			}
		}
	}
	if got["filex_operations_total"] != 1 ||
		got["filex_operation_bytes_total"] != 1024 ||
		got["filex_errors_total"] != 1 {
		t.Errorf("转发不符：%v", got)
	}
}
