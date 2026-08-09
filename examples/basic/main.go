// basic 示例:一个 metricsx 实例喂饱 cachex 与 resiliencex。
package main

import (
	"fmt"
	"time"

	"github.com/lcylpzls/cachex"
	"github.com/lcylpzls/metricsx"
	"github.com/lcylpzls/resiliencex"
)

func main() {
	m, err := metricsx.New(metricsx.WithNamespace("demo"))
	if err != nil {
		panic(err)
	}
	// 预注册声明标签键名(可选)
	_ = m.Register("cachex.hits", "缓存命中", "method")

	cache, err := cachex.New(cachex.WithMetrics(m))
	if err != nil {
		panic(err)
	}
	defer cache.Close()

	limiter, err := resiliencex.NewTokenBucket(10, 5, resiliencex.WithMetrics(m))
	if err != nil {
		panic(err)
	}

	cache.Set("k", 1)
	_, _ = cache.Get("k")
	_, _ = cache.Get("miss")
	_ = limiter.Allow()

	families, err := m.Gather()
	if err != nil {
		panic(err)
	}
	fmt.Printf("指标族数量:%d\n", len(families))
	for _, f := range families {
		fmt.Printf("  %s\n", f.GetName())
	}
	_ = time.Second
}
