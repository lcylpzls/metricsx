# metricsx 与直接使用 client_golang 对比

> 更新日期:2026-08-09 · metricsx v0.4.0

## 1. 性能(同机实测)

| 场景 | metricsx | 直接 client_golang | 备注 |
| --- | --- | --- | --- |
| IncCounter 命中 | 42 ns | 21 ns | 适配成本约 2 倍 |
| ObserveDuration 命中 | 43 ns | 21 ns | 同上 |

## 2. 能力对比

| 能力 | metricsx | 直接 client_golang | 常见 Web 适配器 |
| --- | --- | --- | --- |
| 统一接口服务全部库 | ✅ 一次注入全栈 | ❌ 每库手写 | ❌ 绑定单框架 |
| 懒创建 Vec | ✅ | ❌ 手动注册 | ✅ |
| 命名规范化(点号转下划线) | ✅ | ❌ 需手写 | ❌ |
| 预注册标签键名 | ✅ | ✅ 但手工 | ❌ |
| 占位标签(label0..N) | ✅ | ❌ | ❌ |
| 非法输入防护 | ✅ 静默忽略 | ❌ panic | ❌ |
| 自定义注册表隔离 | ✅ | ✅ | ❌ |

## 3. 取舍

- **metricsx 胜在**:一次注入喂饱全部库、命名/标签统一、
  非法输入不 panic、懒创建零配置;
- **直接 client_golang 胜在**:少一层适配,极致热路径;
- **常见 Web 适配器**:只解决单框架,不可复用。

## 4. 选型建议

- 底座生态(confx/logx/errx/httpx/cachex/resiliencex/jobx):metricsx;
- 仅单个指标且追求极致性能:直接 client_golang;
- 单框架 Web 应用:框架自带适配器即可。
