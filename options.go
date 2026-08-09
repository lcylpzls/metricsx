package metricsx

import (
	"regexp"

	"github.com/lcylpzls/errx"
	"github.com/prometheus/client_golang/prometheus"
)

// metricNamePattern 是 Prometheus 指标名合法字符(字母数字下划线冒号)。
var metricNamePattern = regexp.MustCompile(`^[a-zA-Z_:][a-zA-Z0-9_:]*$`)

// defaultBuckets 是 ObserveDuration 的默认直方图分桶(1ms → 10s 指数)。
var defaultBuckets = []float64{
	0.001, 0.002, 0.005, 0.01, 0.02, 0.05,
	0.1, 0.2, 0.5, 1, 2, 5, 10,
}

// config 是适配器配置。
type config struct {
	namespace string
	registry  prometheus.Registerer
	buckets   []float64
}

func defaultConfig() config {
	return config{
		registry: prometheus.DefaultRegisterer,
		buckets:  defaultBuckets,
	}
}

// Option 修改适配器配置,在 New 时按顺序应用。
type Option func(*config)

// WithNamespace 设置指标命名空间前缀(如 "myapp"),空串表示无前缀。
// 命名空间须匹配 Prometheus 命名规则。
func WithNamespace(ns string) Option {
	return func(c *config) { c.namespace = ns }
}

// WithRegistry 注入注册表,默认 prometheus.DefaultRegisterer。
// 测试可传入独立注册表避免污染全局。
func WithRegistry(r prometheus.Registerer) Option {
	return func(c *config) { c.registry = r }
}

// WithBuckets 设置 ObserveDuration 的直方图分桶,空串表示使用默认。
func WithBuckets(buckets []float64) Option {
	return func(c *config) { c.buckets = buckets }
}

// validateConfig 校验配置参数。
func validateConfig(cfg config) error {
	if cfg.namespace != "" && !metricNamePattern.MatchString(cfg.namespace) {
		return errx.NewCodef(CodeInvalidConfig,
			"命名空间 %q 非法,仅允许字母数字下划线冒号", cfg.namespace)
	}
	if cfg.registry == nil {
		return errx.NewCode(CodeInvalidConfig, "注册表不能为空")
	}
	for _, b := range cfg.buckets {
		if b <= 0 {
			return errx.NewCode(CodeInvalidConfig, "分桶必须为正数")
		}
	}
	return nil
}
