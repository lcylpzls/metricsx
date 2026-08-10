// Package filex 提供对象存储 filex 指标接口到 metricsx 的适配。
package filex

import "github.com/lcylpzls/metricsx"

// Hook 实现 filex.Metrics，把操作/错误事件桥接到 metricsx。
type Hook struct {
	m *metricsx.Metrics
}

// New 创建 filex 指标适配器，传给 filex.Config.Metrics 使用。
func New(m *metricsx.Metrics) *Hook {
	return &Hook{m: m}
}

// Add 记录一次操作与字节量。
func (h *Hook) Add(bucket, operation string, bytes int64) {
	h.m.IncCounter("filex.operations", bucket, operation)
	h.m.AddCounter("filex.operation_bytes", float64(bytes), bucket, operation)
}

// IncError 记录一次错误。
func (h *Hook) IncError(bucket, code string) {
	h.m.IncCounter("filex.errors", bucket, code)
}
