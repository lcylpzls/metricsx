package core

import "github.com/lcylpzls/errx"

// 错误码定义:metricsx 各失败场景的错误码。
const (
	// CodeInvalidConfig 配置非法。
	CodeInvalidConfig errx.Code = "MTRX_INVALID_CONFIG"
	// CodeAlreadyRegistered 指标已注册。
	CodeAlreadyRegistered errx.Code = "MTRX_ALREADY_REGISTERED"
)

func init() {
	errx.RegisterCode(CodeInvalidConfig, "配置非法")
	errx.RegisterCodeKind(CodeInvalidConfig, errx.KindInvalid)
	errx.RegisterCode(CodeAlreadyRegistered, "指标已注册")
	errx.RegisterCodeKind(CodeAlreadyRegistered, errx.KindInvalid)
}
