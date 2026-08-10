package errx

import (
	testx "github.com/lcylpzls/testx"
	"testing"

	"github.com/lcylpzls/errx"
	"github.com/lcylpzls/metricsx"
	prometheusx "github.com/lcylpzls/metricsx/prometheus"
	"github.com/prometheus/client_golang/prometheus"
	dto "github.com/prometheus/client_model/go"
)

func TestInstallUninstall(t *testing.T) {
	m, err := metricsx.New(prometheusx.WithPrometheus(prometheusx.WithRegistry(prometheus.NewRegistry())))
	testx.RequireNoError(t, err)

	Install(m)
	defer Uninstall()

	_ = errx.NewCode("ADAPTER_A", "首次构造")
	families, err := prometheusx.Gather(m)
	testx.RequireNoError(t, err)

	count := metricCount(families, "errx_constructed_total")
	if count < 1 {
		t.Fatalf("安装后应捕获构造事件：%v", count)
	}

	Uninstall()
	_ = errx.NewCode("ADAPTER_B", "卸载后构造")
	families, _ = prometheusx.Gather(m)
	if got := metricCount(families, "errx_constructed_total"); got != count {
		t.Errorf("卸载后不应再计数：%v -> %v", count, got)
	}
}

func metricCount(families []*dto.MetricFamily, name string) float64 {
	for _, f := range families {
		if f.GetName() == name {
			return f.GetMetric()[0].GetCounter().GetValue()
		}
	}
	return 0
}
