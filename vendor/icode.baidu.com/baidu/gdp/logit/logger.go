// Copyright(C) 2020 Baidu Inc. All Rights Reserved.
// Author: Chen Xin (chenxin@baidu.com)
// Date: 2020/04/19

package logit

import (
	"context"
	"errors"
	"fmt"
	"io"
	"strings"
)

// Logger 接口定义
type Logger interface {
	Debug(ctx context.Context, message string, fields ...Field)
	Trace(ctx context.Context, message string, fields ...Field)
	Notice(ctx context.Context, message string, fields ...Field)
	Warning(ctx context.Context, message string, fields ...Field)
	Error(ctx context.Context, message string, fields ...Field)
	Fatal(ctx context.Context, message string, fields ...Field)

	Output(ctx context.Context, level Level, callDepth int, message string, fields ...Field)
}

// Binder 可以与 Logger 绑定的组件
type Binder interface {
	// SetLogger 设置打日志的Logger
	SetLogger(Logger)

	// Logger 获取Logger
	Logger() Logger
}

// WithLogger 默认的Binder实现
type WithLogger struct {
	logger Logger
}

// SetLogger 设置logger
func (b *WithLogger) SetLogger(logger Logger) {
	b.logger = logger
}

// Logger 返回logger
func (b *WithLogger) Logger() Logger {
	return b.logger
}

// AutoLogger 自动获取logger，若未设置，会返回DefaultLogger
func (b *WithLogger) AutoLogger() Logger {
	if b.logger != nil {
		return b.logger
	}
	return DefaultLogger
}

// MultiLogger 将多个logger转换为一个，实现日志多目标输出；类似io.MultiWriter
func MultiLogger(loggers ...Logger) Logger {
	allLoggers := make([]Logger, 0, len(loggers))
	for _, l := range loggers {
		if ml, ok := l.(*multiLogger); ok {
			allLoggers = append(allLoggers, ml.loggers...)
		} else {
			allLoggers = append(allLoggers, l)
		}
	}
	return &multiLogger{
		loggers: allLoggers,
	}
}

type multiLogger struct {
	loggers []Logger
}

func (m *multiLogger) Debug(ctx context.Context, message string, fields ...Field) {
	m.Output(ctx, DebugLevel, 1, message, fields...)
}

func (m *multiLogger) Trace(ctx context.Context, message string, fields ...Field) {
	m.Output(ctx, TraceLevel, 1, message, fields...)
}

func (m *multiLogger) Notice(ctx context.Context, message string, fields ...Field) {
	m.Output(ctx, NoticeLevel, 1, message, fields...)
}

func (m *multiLogger) Warning(ctx context.Context, message string, fields ...Field) {
	m.Output(ctx, WarningLevel, 1, message, fields...)
}

func (m *multiLogger) Error(ctx context.Context, message string, fields ...Field) {
	m.Output(ctx, ErrorLevel, 1, message, fields...)
}

func (m *multiLogger) Fatal(ctx context.Context, message string, fields ...Field) {
	m.Output(ctx, FatalLevel, 1, message, fields...)
}

func (m *multiLogger) Output(ctx context.Context, level Level, callDepth int, message string, fields ...Field) {
	for _, l := range m.loggers {
		l.Output(ctx, level, callDepth+1, message, fields...)
	}
}

func (m *multiLogger) Close() error {
	if len(m.loggers) == 0 {
		return nil
	}
	var b strings.Builder

	for idx, lg := range m.loggers {
		if lc, ok := lg.(io.Closer); ok {
			if e := lc.Close(); e != nil {
				b.WriteString(fmt.Sprintf("logger[idx=%d].Close() with error: %s;", idx, e))
			}
		}
	}
	if b.Len() == 0 {
		return nil
	}
	return errors.New(b.String())
}

var _ Logger = (*multiLogger)(nil)

// TryCloseLogger 尝试关闭 logger
func TryCloseLogger(l Logger) error {
	if c, ok := l.(io.Closer); ok {
		return c.Close()
	}
	return nil
}

// NewLoggerWithOutput 使用 output 函数快速创建一个 logger
func NewLoggerWithOutput(fn func(ctx context.Context, level Level, callDepth int, message string, fields ...Field)) Logger {
	return &skinnyLogger{
		output: fn,
	}
}

var _ Logger = (*skinnyLogger)(nil)

type skinnyLogger struct {
	output func(ctx context.Context, level Level, callDepth int, message string, fields ...Field)
}

func (sl *skinnyLogger) Debug(ctx context.Context, message string, fields ...Field) {
	sl.Output(ctx, DebugLevel, 1, message, fields...)
}

func (sl *skinnyLogger) Trace(ctx context.Context, message string, fields ...Field) {
	sl.Output(ctx, TraceLevel, 1, message, fields...)
}

func (sl *skinnyLogger) Notice(ctx context.Context, message string, fields ...Field) {
	sl.Output(ctx, NoticeLevel, 1, message, fields...)
}

func (sl *skinnyLogger) Warning(ctx context.Context, message string, fields ...Field) {
	sl.Output(ctx, WarningLevel, 1, message, fields...)
}

func (sl *skinnyLogger) Error(ctx context.Context, message string, fields ...Field) {
	sl.Output(ctx, ErrorLevel, 1, message, fields...)
}

func (sl *skinnyLogger) Fatal(ctx context.Context, message string, fields ...Field) {
	sl.Output(ctx, FatalLevel, 1, message, fields...)
}

func (sl *skinnyLogger) Output(ctx context.Context, level Level, callDepth int, message string, fields ...Field) {
	sl.output(ctx, level, callDepth+1, message, fields...)
}
