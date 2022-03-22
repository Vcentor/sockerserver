// Copyright(C) 2021 Baidu Inc. All Rights Reserved.
// Author: Wei Du (duwei04@baidu.com)
// Date: 2021/12/6

package logit

import (
	"bytes"
	"context"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"sync"

	"icode.baidu.com/baidu/gdp/env"
)

var (
	// PanicLogger 用于打印 panic 日志的logger
	// 会在调用 ReportPanic 的时候延迟加载 logger
	PanicLogger Logger = &nopLogger{}
)

var panicLoadOnce sync.Once

// ReportPanic 打印 panic 信息到日志
// 	首次使用的时候，会尝试重新初始化 PanicLogger (若用户已初始化为非 nopLogger，将跳过)
// 	初始化时将尝试使用配置：logit/panic.toml，若配置不存在，将输出到 log/panic/panic.log
// 	若是处于 testing 状态，logger 会将内容输出到 stderr,而不是文件
// 	日志将输出成一行,在 shell 里可以使用：sed 's/\\n/\n/g' 将换行符还原
// 	如：cat panic.log|sed 's/\\n/\n/g'
// 	若输入 msg 是字符串 "ignore"，将只进行初始化，不打印日志
func ReportPanic(ctx context.Context, msg interface{}, fields ...Field) {
	panicLoadOnce.Do(loadPanicLogger)
	if str, ok := msg.(string); ok && str == "ignore" {
		return
	}

	trace := make([]byte, 4096)
	n := runtime.Stack(trace[:], false)
	fields = append(fields,
		CallerFieldWithSkip(2),
		String("panic", fmt.Sprint(msg)),
		ByteString("stack", bytes.ReplaceAll(trace[:n], []byte("\n"), []byte("\\n"))),
	)
	PanicLogger.Output(ctx, FatalLevel, 1, "panic", fields...)
}

// LoadPanicLogger 尝试使用默认的配置加载 PanicLogger
// 	将尝试使用配置：logit/panic.toml，若配置不存在，将输出到 log/panic/panic.log
func initPanicLogger() (Logger, error) {
	opts := []Option{
		OptConfigFileOrLogFileName("logit/panic.toml", filepath.Join(env.LogDir(), "panic", "panic.log")),
		OptPrefixName("default_nano"),
		OptSetConfigFn(func(c *Config) {
			c.Dispatch = []*ConfigDispatch{
				{
					FileSuffix: "",
					Levels:     []Level{DebugLevel, TraceLevel, NoticeLevel, WarningLevel, ErrorLevel, FatalLevel},
				},
			}
		}),
	}
	return NewLogger(context.Background(), opts...)
}

func loadPanicLogger() {
	// 已经是自定义的 logger，则不需要加载
	if !IsNopLogger(PanicLogger) {
		return
	}

	if isTesting() {
		PanicLogger = NewSimple(os.Stderr)
		return
	}

	pl, err := initPanicLogger()
	if err != nil {
		log.Println("loadPanicLogger failed:", err.Error())
		return
	}
	PanicLogger = pl
}

var isTestingFlag bool
var isTestingOnce sync.Once

func isTesting() bool {
	isTestingOnce.Do(func() {
		if len(os.Args) > 0 {
			isTestingFlag = strings.Contains(filepath.Base(os.Args[0]), ".test")
		}
	})
	return isTestingFlag
}
