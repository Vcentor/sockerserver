// Copyright(C) 2020 Baidu Inc. All Rights Reserved.
// Author: Chen Xin (chenxin@baidu.com)
// Date: 2020/04/19

package logit

import "context"

var (
	// NopLogger 黑洞，调用这个logger，什么也不做
	NopLogger Logger = &nopLogger{}
)

// IsNopLogger 判断当前 logger 是否是 NopLogger
func IsNopLogger(lg Logger) bool {
	l, ok := lg.(isNopLogger)
	return ok && l.NopLogger()
}

// isNopLogger 是否是 noopLogger
type isNopLogger interface {
	NopLogger() bool
}

// nopLogger do nothing, just implement the interface
type nopLogger struct{}

func (nl *nopLogger) Debug(ctx context.Context, message string, fields ...Field)   {}
func (nl *nopLogger) Trace(ctx context.Context, message string, fields ...Field)   {}
func (nl *nopLogger) Notice(ctx context.Context, message string, fields ...Field)  {}
func (nl *nopLogger) Warning(ctx context.Context, message string, fields ...Field) {}
func (nl *nopLogger) Error(ctx context.Context, message string, fields ...Field)   {}
func (nl *nopLogger) Fatal(ctx context.Context, message string, fields ...Field)   {}
func (nl *nopLogger) Output(ctx context.Context, level Level, callDepth int, message string, fields ...Field) {
}
func (nl *nopLogger) NopLogger() bool {
	return true
}

var _ Logger = (*nopLogger)(nil)
var _ isNopLogger = (*nopLogger)(nil)
