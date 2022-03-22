// Copyright(C) 2020 Baidu Inc. All Rights Reserved.
// Author: Wei Du (duwei04@baidu.com)
// Date: 2020/7/24

package logit

import (
	"context"
)

// newDispatcher 将日志按照(MinLevel)等级分发到不同的logger
func newDispatcher(dispatchFunc func(level Level) Logger) *dispatcher {
	return &dispatcher{
		dispatchFunc: dispatchFunc,
	}
}

type dispatcher struct {
	dispatchFunc func(level Level) Logger
	closeFunc    func() error
}

func (d *dispatcher) Debug(ctx context.Context, message string, fields ...Field) {
	d.Output(ctx, DebugLevel, 1, message, fields...)
}

func (d *dispatcher) Trace(ctx context.Context, message string, fields ...Field) {
	d.Output(ctx, TraceLevel, 1, message, fields...)
}

func (d *dispatcher) Notice(ctx context.Context, message string, fields ...Field) {
	d.Output(ctx, NoticeLevel, 1, message, fields...)
}

func (d *dispatcher) Warning(ctx context.Context, message string, fields ...Field) {
	d.Output(ctx, WarningLevel, 1, message, fields...)
}

func (d *dispatcher) Error(ctx context.Context, message string, fields ...Field) {
	d.Output(ctx, ErrorLevel, 1, message, fields...)
}

func (d *dispatcher) Fatal(ctx context.Context, message string, fields ...Field) {
	d.Output(ctx, FatalLevel, 1, message, fields...)
}

func (d *dispatcher) Output(ctx context.Context, level Level, callDepth int, message string, fields ...Field) {
	d.dispatchFunc(level).Output(ctx, level, callDepth+1, message, fields...)
}

func (d *dispatcher) Close() error {
	if d.closeFunc != nil {
		return d.closeFunc()
	}
	return nil
}

var _ Logger = (*dispatcher)(nil)
