package clix

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/lcylpzls/metricsx"
	prometheusx "github.com/lcylpzls/metricsx/prometheus"
	testx "github.com/lcylpzls/testx"
	"github.com/prometheus/client_golang/prometheus"
)

func TestObserverMetrics(t *testing.T) {
	reg := prometheus.NewRegistry()
	m, err := metricsx.New(prometheusx.WithPrometheus(prometheusx.WithRegistry(reg)))
	testx.RequireNoError(t, err)
	obs := Observer(m)
	obs.OnCommandStart(context.Background(), "greet hello", []string{"a"})
	obs.OnCommandFinish(context.Background(), "greet hello", []string{"a"}, nil, time.Second)
	obs.OnCommandStart(context.Background(), "greet boom", nil)
	obs.OnCommandFinish(context.Background(), "greet boom", nil, errors.New("失败"), time.Second)

	families, err := prometheusx.Gather(m)
	testx.RequireNoError(t, err)
	names := map[string]bool{}
	for _, f := range families {
		names[f.GetName()] = true
	}
	for _, want := range []string{
		"clix_commands_started_total",
		"clix_commands_failed_total",
		"clix_command_duration_seconds",
	} {
		testx.True(t, names[want])
	}
}
