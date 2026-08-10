// Package errx 提供 errx 全局指标钩子的 metricsx 适配。
package errx

import (
	"github.com/lcylpzls/errx"
	"github.com/lcylpzls/metricsx"
)

// hook 实现 errx.MetricsHook，转发给 metricsx。
type hook struct {
	m *metricsx.Metrics
}

func (h hook) IncCounter(name string, labels ...string) {
	h.m.IncCounter(name, labels...)
}

// Install 安装全局指标钩子，将 errx 错误构造/查询事件转发给 metricsx。
// 全局生效；同一进程只需安装一次。
func Install(m *metricsx.Metrics) {
	errx.SetMetricsHook(hook{m: m})
}

// Uninstall 卸载全局指标钩子，恢复零额外开销。
func Uninstall() {
	errx.ResetMetricsHook()
}
