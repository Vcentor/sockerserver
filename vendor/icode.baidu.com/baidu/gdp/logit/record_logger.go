// Copyright(C) 2020 Baidu Inc. All Rights Reserved.
// Author: Wei Du (duwei04@baidu.com)
// Date: 2020/8/19

package logit

import (
	"context"
	"fmt"
	"io"

	// nolint
	"github.com/golang/protobuf/proto"

	"icode.baidu.com/baidu/gdp/logit/b2log"
)

// RecordLogger 另外一个可用于打印二进制数据的logger
type RecordLogger interface {
	Record(msg interface{}) error
}

// OptRecordEncoder 专用于 RecordLogger，设置数据的Encoder func
func OptRecordEncoder(fn func(msg interface{}) ([]byte, error)) Option {
	return newFuncOption(func(config *Config) {
		config.binaryEncoder = fn
	})
}

// NewRecordLogger 创建一个 RecordLogger 实例
func NewRecordLogger(ctx context.Context, opts ...Option) (RecordLogger, error) {
	cfg, err := loggerCfg(opts...)
	if err != nil {
		return nil, err
	}

	if cfg.binaryEncoder == nil {
		return nil, fmt.Errorf("encodeFunc is nil")
	}

	writer, err := cfg.getWriter()
	if err != nil {
		return nil, err
	}
	l := &bLogger{
		writer:     writer,
		encodeFunc: cfg.binaryEncoder,
	}

	if ctx != nil {
		go func() {
			<-ctx.Done()
			l.Close()
		}()
	}
	return l, nil
}

type bLogger struct {
	encodeFunc func(obj interface{}) ([]byte, error)
	writer     io.WriteCloser
}

func (b *bLogger) Record(msg interface{}) error {
	logData, err := b.encodeFunc(msg)
	if err != nil {
		return err
	}

	if _, err := b.writer.Write(logData); err != nil {
		return err
	}
	return nil
}

func (b *bLogger) Close() error {
	return b.writer.Close()
}

var _ RecordLogger = (*bLogger)(nil)

// B2Logger 用于打印公司的pblog
type B2Logger interface {
	Record(msg proto.Message) error
}

// NewB2Logger 创建一个B2Logger 的实例
func NewB2Logger(ctx context.Context, opts ...Option) (B2Logger, error) {
	fn := func(msg interface{}) ([]byte, error) {
		return b2log.Encode(msg.(proto.Message))
	}

	opts = append(opts, OptRecordEncoder(fn))

	bLogger, err := NewRecordLogger(ctx, opts...)
	if err != nil {
		return nil, err
	}
	return &b2logIns{
		blogger: bLogger,
	}, nil
}

type b2logIns struct {
	blogger RecordLogger
}

func (b *b2logIns) Record(msg proto.Message) error {
	return b.blogger.Record(msg)
}

var _ B2Logger = (*b2logIns)(nil)
