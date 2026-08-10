// Package jobx 提供任务调度 jobx 指标回调到 metricsx 的适配。
package jobx

import (
	"strconv"
	"time"

	"github.com/lcylpzls/errx"
	"github.com/lcylpzls/jobx"
	"github.com/lcylpzls/metricsx"
)

// New 返回接入 metricsx 的 jobx.Metrics 回调集合，传给 jobx.WithMetrics。
func New(m *metricsx.Metrics) jobx.Metrics {
	return jobx.Metrics{
		Queued: func(name string, delta int) {
			m.AddGauge("jobx.queued", float64(delta), name)
		},
		Running: func(name string, delta int) {
			m.AddGauge("jobx.running", float64(delta), name)
		},
		Completed: func(name string, duration time.Duration) {
			m.IncCounter("jobx.completed", name)
			m.ObserveDuration("jobx.completed_duration", duration.Seconds(), name)
		},
		Failed: func(name string, err error) {
			m.IncCounter("jobx.failed", name, codeOf(err))
		},
		Retried: func(name string, attempt int) {
			m.IncCounter("jobx.retried", name, strconv.Itoa(attempt))
		},
		Dropped: func(name string) {
			m.IncCounter("jobx.dropped", name)
		},
		Skipped: func(name string) {
			m.IncCounter("jobx.skipped", name)
		},
		Replaced: func(name string) {
			m.IncCounter("jobx.replaced", name)
		},
	}
}

// codeOf 返回错误码标签，无结构化码时使用 unknown。
func codeOf(err error) string {
	if code, ok := errx.CodeOf(err); ok {
		return string(code)
	}
	return "unknown"
}
