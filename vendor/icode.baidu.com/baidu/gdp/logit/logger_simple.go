// Copyright(C) 2020 Baidu Inc. All Rights Reserved.
// Author: Wei Du (duwei04@baidu.com)
// Date: 2020/7/23

package logit

import (
	"context"
	"io"

	"icode.baidu.com/baidu/gdp/extension/pool"
)

// NewSimple 创建一个简单的logger
// 默认使用 text编码
func NewSimple(writer io.Writer) Logger {
	return &SimpleLogger{
		PrefixFunc:  DefaultPrefixFunc,
		Writer:      writer,
		EncoderPool: DefaultTextEncoderPool,

		// 此等级将打印所有的日志
		MinLevel: UnknownLevel,
	}
}

// SimpleLogger 默认的一个logger
type SimpleLogger struct {
	// 日志前缀
	PrefixFunc PrefixFunc

	// 在 Output 执行前执行
	BeforeOutputFunc BeforeOutputFunc

	// 由于Writer 是外部传入的，所以在SimpleLogger 内部，不能对其close
	Writer io.Writer

	EncoderPool EncoderPool

	// 最小日志等级，低于此等级的日志信息将不打印
	MinLevel Level
}

// Debug debug
func (sl *SimpleLogger) Debug(ctx context.Context, message string, fields ...Field) {
	sl.Output(ctx, DebugLevel, 1, message, fields...)
}

// Trace Trace
func (sl *SimpleLogger) Trace(ctx context.Context, message string, fields ...Field) {
	sl.Output(ctx, TraceLevel, 1, message, fields...)
}

// Notice Notice
func (sl *SimpleLogger) Notice(ctx context.Context, message string, fields ...Field) {
	sl.Output(ctx, NoticeLevel, 1, message, fields...)
}

// Warning Warning
func (sl *SimpleLogger) Warning(ctx context.Context, message string, fields ...Field) {
	sl.Output(ctx, WarningLevel, 1, message, fields...)
}

// Error Error
func (sl *SimpleLogger) Error(ctx context.Context, message string, fields ...Field) {
	sl.Output(ctx, ErrorLevel, 1, message, fields...)
}

// Fatal Fatal
func (sl *SimpleLogger) Fatal(ctx context.Context, message string, fields ...Field) {
	sl.Output(ctx, FatalLevel, 1, message, fields...)
}

// Output Output
func (sl *SimpleLogger) Output(ctx context.Context, level Level, callDepth int, message string, fields ...Field) {
	if level == UnknownLevel || level >= AllLevels {
		return
	}

	if sl.MinLevel > level || sl.MinLevel >= AllLevels {
		return
	}

	enc := sl.EncoderPool.Get()
	defer sl.EncoderPool.Put(enc)

	if sl.BeforeOutputFunc != nil {
		sl.BeforeOutputFunc(ctx, enc, level, callDepth+2)
	}

	// 字段的顺序
	// 1：ctx里存储的字段  2：全局meta fields  3：传入的临时补充字段

	// 10-只是一个经验值，meta fields 不应该太多
	fkv := make(map[string]Field, len(fields)+10)

	var metaFields []Field
	RangeMetaFields(ctx, func(f Field) error {
		metaFields = append(metaFields, f)
		fkv[f.Key()] = f
		return nil
	})

	for _, f := range fields {
		fkv[f.Key()] = f
	}

	// ctx 里存储的 fields 优先级 第一
	Range(ctx, func(f Field) error {
		if f.Level().Is(level) {
			// 若字段之前在ctx，后面又在fields 里出现，则使用后面传入的
			if fn, has := fkv[f.Key()]; has {
				fn.AddTo(enc)
				delete(fkv, f.Key()) // 不需要在打印了
			} else {
				f.AddTo(enc)
			}
		}
		return nil
	})

	if len(fkv) > 0 {
		if len(metaFields) > 0 {
			// meta fields 优先级第二
			for _, f := range metaFields {
				if lastField, has := fkv[f.Key()]; has {
					lastField.AddTo(enc)
					delete(fkv, f.Key())
				}
			}
		}

		// 补充字段 优先级第三
		for _, f := range fields {
			if lastField, has := fkv[f.Key()]; has {
				lastField.AddTo(enc)
			}
		}
	}

	enc.AddString("message", message)
	// 相当于 String("message", message).AddTo(enc)

	buf := bpSL.Get()

	if sl.PrefixFunc != nil {
		prefix := sl.PrefixFunc(ctx, level, callDepth+2)
		if len(prefix) > 0 {
			_, _ = buf.Write(prefix)
		}
	}
	_, _ = enc.WriteTo(buf)

	logLine := make([]byte, buf.Len())
	copy(logLine, buf.Bytes())
	_, _ = sl.Writer.Write(logLine)

	bpSL.Put(buf)
}

var bpSL = pool.NewBytesPool()

var _ Logger = (*SimpleLogger)(nil)
