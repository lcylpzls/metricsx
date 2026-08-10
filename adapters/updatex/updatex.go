// Package updatex 提供自动升级 updatex 指标回调到 metricsx 的适配。
package updatex

import (
	"github.com/lcylpzls/errx"
	"github.com/lcylpzls/metricsx"
	"github.com/lcylpzls/updatex"
)

// New 返回接入 metricsx 的 updatex.Metrics 回调集合，
// 赋值给 updatex.Config.Metrics 使用。
func New(m *metricsx.Metrics) updatex.Metrics {
	return updatex.Metrics{
		CheckTotal: func(delta int) {
			m.AddCounter("updatex.checks", float64(delta))
		},
		CheckFailures: func(err error) {
			m.IncCounter("updatex.check_failures", codeOf(err))
		},
		UpdateSuccess: func(version string) {
			m.IncCounter("updatex.updates", "success", version)
		},
		UpdateFailures: func(err error) {
			m.IncCounter("updatex.updates", "failure", codeOf(err))
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
