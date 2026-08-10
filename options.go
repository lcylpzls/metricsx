package metricsx

import "github.com/lcylpzls/errx"

// config 是核心配置。
type config struct {
	sink    Sink
	sinkSet bool
}

func defaultConfig() config {
	return config{}
}

// Option 修改核心配置，在 New 时按顺序应用。
type Option func(*config)

// WithSink 注入自定义指标后端；未设置时使用内置内存后端。
// 显式传入 nil 视为配置错误。
func WithSink(sink Sink) Option {
	return func(c *config) {
		c.sink = sink
		c.sinkSet = true
	}
}

// validateConfig 校验核心配置。
func validateConfig(cfg config) error {
	if cfg.sinkSet && cfg.sink == nil {
		return errx.NewCode(CodeInvalidConfig, "指标后端不能为空")
	}
	return nil
}
