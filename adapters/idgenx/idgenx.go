// Package idgenx 提供分布式 ID 生成器 idgenx 指标回调到 metricsx 的适配。
package idgenx

import (
	"strconv"

	"github.com/lcylpzls/errx"
	"github.com/lcylpzls/idgenx"
	"github.com/lcylpzls/metricsx"
)

// New 返回接入 metricsx 的 idgenx.Metrics 回调集合，传给 idgenx.WithMetrics。
func New(m *metricsx.Metrics) idgenx.Metrics {
	return idgenx.Metrics{
		Generated: func(node int64, delta int) {
			m.AddCounter("idgenx.generated", float64(delta), strconv.FormatInt(node, 10))
		},
		Rejected: func(node int64, err error) {
			m.IncCounter("idgenx.rejected", strconv.FormatInt(node, 10), codeOf(err))
		},
		WaitMS: func(node int64, ms int64) {
			m.ObserveDuration("idgenx.wait_duration", float64(ms)/1000, strconv.FormatInt(node, 10))
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
