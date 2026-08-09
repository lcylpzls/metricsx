# 质量保障

## 门槛(每个版本强制)

- 语句覆盖率 100%;
- `go vet` / `staticcheck` 零告警;
- `go test -race` 全绿;
- fuzz 目标至少 1 个(名称/标签);
- 三平台 CI(ubuntu / windows / macos)× Go 1.26;
- govulncheck 零告警;go.mod tidy 无漂移;apidiff 对比上一 tag;
- 示例全部可构建并通过 vet。

## 测试策略

- 懒创建 / 预注册 / 占位标签 / 命名规范化全路径;
- gather 集成断言(真实 Prometheus 注册表);
- 标签不匹配 / 非法 UTF-8 / 注册失败静默忽略;
- 并发首建(race + barrier);默认注册表写入验证;
- fuzz 回归种子(非法 UTF-8 输入)。

## 性能

见 [performance.md](performance.md)。热路径要求 0 分配,
指标查找无锁读。

## API 兼容性

- <1.0.0 允许有意的破坏性变更,须在 CHANGELOG 说明;
- v1.0.0 起 API 冻结,破坏性变更仅随大版本。
