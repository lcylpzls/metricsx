// Package clix 提供 clix.Observer 到 metricsx 的适配：
// 命令开始计数、失败计数与耗时直方图。
package clix

import (
	"context"
	"time"

	"github.com/lcylpzls/clix"
	"github.com/lcylpzls/metricsx"
)

// Observer 返回接入 metricsx 的 clix.Observer。
func Observer(m *metricsx.Metrics) clix.Observer {
	return clixObserver{m: m}
}

type clixObserver struct {
	m *metricsx.Metrics
}

func (o clixObserver) OnCommandStart(_ context.Context, command string, _ []string) {
	o.m.IncCounter("clix.commands_started", command)
}

func (o clixObserver) OnCommandFinish(_ context.Context, command string, _ []string, err error, duration time.Duration) {
	if err != nil {
		o.m.IncCounter("clix.commands_failed", command)
	}
	o.m.ObserveDuration("clix.command_duration", duration.Seconds(), command)
}
