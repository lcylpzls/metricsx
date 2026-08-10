package core

import (
	"github.com/lcylpzls/errx"
	"github.com/lcylpzls/validx"
)

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

// init 注册配置校验规则到 validx 全局规则表，错误码保持 metricsx 语义。
func init() {
	_ = validx.RegisterRule("metricsx_config", func(value any, param, path string) error {
		// 内部调用保证 value 为 config。
		cfg := value.(config)
		if cfg.sinkSet && cfg.sink == nil {
			return errx.NewCode(CodeInvalidConfig, "指标后端不能为空")
		}
		return nil
	})
}

// validateConfig 校验核心配置（统一走 validx 规则）。
func validateConfig(cfg config) error {
	return validx.ValidateField(cfg, "metricsx_config")
}
