// Copyright(C) 2020 Baidu Inc. All Rights Reserved.
// Author: Wei Du (duwei04@baidu.com)
// Date: 2020/7/24

package logit

import (
	"context"
	"errors"
	"fmt"
	"io"
	"os"
	"strings"
	"time"

	"icode.baidu.com/baidu/gdp/conf"
)

// DefaultLogger 默认logger:内容输出到黑洞
var DefaultLogger Logger = &nopLogger{}

// Option FileLogger 的配置选项
type Option interface {
	apply(*Config)
}

type funcOption struct {
	f func(*Config)
}

func (fdo *funcOption) apply(do *Config) {
	fdo.f(do)
}

func newFuncOption(f func(*Config)) *funcOption {
	return &funcOption{
		f: f,
	}
}

func errOption(err error) Option {
	return newFuncOption(func(config *Config) {
		if err != nil {
			config.err = err
		}
	})
}

// OptConfig 配置选项-整体配置
func OptConfig(c *Config) Option {
	return newFuncOption(func(config *Config) {
		*config = *c
	})
}

// OptConfigFile 配置选项-从文件读取配置
func OptConfigFile(confName string) Option {
	c, err := LoadConfig(confName)
	if err != nil {
		return errOption(err)
	}
	return OptConfig(c)
}

// OptConfigFileOrLogFileName 当配置文件存在，使用配置文件，若不存在，则直接使用日志文件名
func OptConfigFileOrLogFileName(confName string, logFileName string) Option {
	if conf.Exists(confName) {
		return OptConfigFile(confName)
	}
	return OptLogFileName(logFileName)
}

// OptSetConfigFn 配置选型-直接对Config进行修改
func OptSetConfigFn(fn func(c *Config)) Option {
	return newFuncOption(func(config *Config) {
		fn(config)
	})
}

// OptLogFileName 配置选项-设置日志文件名
//
// 如 OptLogFileName("ral/ral-worker.log")
func OptLogFileName(name string) Option {
	if name == "" {
		return errOption(errors.New("optLogFileName with empty name"))
	}
	return newFuncOption(func(config *Config) {
		config.FileName = name
	})
}

// OptRotateRule 配置选项-设置日志切分规则
//
// 	如 1hour,1day,no,默认为1hour
// 	更多规则详见：http://icode.baidu.com/repos/baidu/gdp/extension/blob/master:writer/rotate_producer.go
func OptRotateRule(rule string) Option {
	if rule == "" {
		return errOption(errors.New("optRotateRule with empty rule"))
	}
	return newFuncOption(func(config *Config) {
		config.RotateRule = rule
	})
}

// OptMaxFileNum 配置选项-设置日志文件保留数
//
// 保留最多日志文件数，默认(当值为0时)为48，若<0,则不会清理
func OptMaxFileNum(num int) Option {
	return newFuncOption(func(config *Config) {
		config.MaxFileNum = num
	})
}

// OptBufferSize 配置选项-设置writer 的 BufferSize
func OptBufferSize(size int) Option {
	return newFuncOption(func(config *Config) {
		config.BufferSize = size
	})
}

// OptWriterTimeout 配置选项-设置writer 的 超时时间
func OptWriterTimeout(timeout time.Duration) Option {
	return newFuncOption(func(config *Config) {
		config.WriterTimeout = int(timeout / time.Millisecond)
	})
}

// OptWriter 配置选项-将writer替换掉
//
// 注意，如此替换writer 后，其他的writer配置参数将不生效
// 如BufferSize、WriterTimeout、RotateRule、MaxFileNum
func OptWriter(w io.WriteCloser) Option {
	return newFuncOption(func(config *Config) {
		config.writer = w
	})
}

// OptFlushDuration 配置选项-writer的刷新间隔(毫秒)
func OptFlushDuration(dur uint) Option {
	if dur == 0 {
		return errOption(fmt.Errorf("flushDuration=%d expect>0", dur))
	}
	return newFuncOption(func(config *Config) {
		config.FlushDuration = int(dur)
	})
}

// OptPrefixFunc 配置选项-设置日志前缀方法
func OptPrefixFunc(fn PrefixFunc) Option {
	return newFuncOption(func(config *Config) {
		config.PrefixFunc = fn
	})
}

// OptPrefixName 配置选项-设置日志前缀方法名称
// 	框架内置的可选值：
// 	default      : 默认，时间精确到秒
// 	default_nano : 时间精确到纳秒
// 	no           : 无前缀
// 	若上述不满足，还可以自定义，或者 OptPrefixFunc 直接定义函数
func OptPrefixName(name string) Option {
	return newFuncOption(func(config *Config) {
		// 赋值为nil，这样避免之前已经设置过 PrefixFunc
		config.PrefixFunc = nil
		config.Prefix = name
	})
}

// loggerCfg 通过加载 Option 初始化生成一个 Config
func loggerCfg(opts ...Option) (*Config, error) {
	cfg := &Config{}
	for _, opt := range opts {
		opt.apply(cfg)
		cfg.parsed = false
	}

	if cfg.err != nil {
		return nil, cfg.err
	}

	if err := cfg.parser(); err != nil {
		return nil, err
	}
	return cfg, nil
}

// NewLogger 创建一个新的logger
//
// 	ctx 用于控制 logger 的 writer 的生命周期
// 	程序在退出前将所有日志落盘，需要将 logger 关闭：
// 	 XXXLogger.(io.Closer).Close()
// 	 或者使用 TryCloseLogger( XXXLogger ) 方法
func NewLogger(ctx context.Context, opts ...Option) (l Logger, errResult error) {
	cfg, err := loggerCfg(opts...)
	if err != nil {
		return nil, err
	}
	nop := &nopLogger{}

	if cfg.nopLog() {
		return nop, nil
	}

	mapper := make(map[Level]Logger, 6) // 默认是6个日志等级
	closeFns := make([]func() error, 0, 6)

	var closeWritersFunc func() error

	if ctx != nil {
		closeWritersFunc = func() error {
			var builder strings.Builder
			for idx, fn := range closeFns {
				if e := fn(); e != nil {
					builder.WriteString(fmt.Sprintf("idx=%d error=%s;", idx, e))
				}
			}
			if builder.Len() == 0 {
				return nil
			}
			return fmt.Errorf("logger close with errors: %s", builder.String())
		}

		closeAll := func() {
			if e := closeWritersFunc(); e != nil {
				fmt.Fprintf(os.Stderr, "%s %v\n", time.Now(), e)
			}
		}

		defer func() {
			if errResult != nil {
				closeAll()
			}
		}()

		go func() {
			<-ctx.Done()
			closeAll()
		}()
	}

	for idx, item := range cfg.Dispatch {
		if len(item.Levels) == 0 {
			continue
		}

		itemOpt := *cfg
		itemOpt.FileName += item.FileSuffix

		awc, err := itemOpt.getWriter()

		if err != nil {
			return nil, fmt.Errorf("init logger (%d)(%q) failed: %w", idx, itemOpt.FileName, err)
		}
		lg := &SimpleLogger{
			PrefixFunc:       cfg.PrefixFunc,
			EncoderPool:      cfg.encoderPool,
			BeforeOutputFunc: cfg.BeforeOutputFunc,
			Writer:           awc,
		}
		closeFns = append(closeFns, awc.Close)

		for _, l := range item.Levels {
			if _, has := mapper[l]; !has {
				mapper[l] = MultiLogger(lg)
			} else {
				mapper[l] = MultiLogger(mapper[l], lg)
			}
		}
	}

	dl := newDispatcher(func(level Level) Logger {
		logger, has := mapper[level]
		if !has {
			return nop
		}
		return logger
	})

	dl.closeFunc = closeWritersFunc

	return dl, nil
}
