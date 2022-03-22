// Copyright(C) 2020 Baidu Inc. All Rights Reserved.
// Author: Chen Xin (chenxin@baidu.com)
// Date: 2020/04/19

package logit

import (
	"runtime"
	"strconv"
	"strings"
	"sync"

	"icode.baidu.com/baidu/gdp/extension/pool"
)

const (
	callerKey = "caller"
	stackKey  = "stack"
)

var (
	pcsPool = sync.Pool{
		New: func() interface{} {
			return &stackPtr{
				pcs: make([]uintptr, 64),
			}
		},
	}
)

type stackPtr struct {
	pcs []uintptr
}

// Stack retrieve call stack
func Stack() Field {
	return StackWithSkip(3)
}

var swsBP = pool.NewBytesPool()

// StackWithSkip 返回调用栈的Field
func StackWithSkip(skip int) Field {
	buf := swsBP.Get()
	stack := pcsPool.Get().(*stackPtr)

	callStackSize := runtime.Callers(skip, stack.pcs)
	frames := runtime.CallersFrames(stack.pcs[:callStackSize])
	for frame, more := frames.Next(); more; frame, more = frames.Next() {
		buf.WriteString(frame.File)
		buf.WriteByte(':')
		buf.WriteString(strconv.Itoa(frame.Line))
		buf.WriteByte(';')
	}
	val := buf.String()

	pcsPool.Put(stack)
	swsBP.Put(buf)
	return String(stackKey, val)
}

// CallerField 默认的获取调用栈的Field
func CallerField() Field {
	return CallerFieldWithSkip(2)
}

// CallerFieldWithSkip 获取调用栈
func CallerFieldWithSkip(skip int) Field {
	return String(callerKey, callerWithSkip(skip+1))
}

// callerWithSkip 获取调用栈的路径
// 如  xxx/xxx/xxx.go:80
func callerWithSkip(skip int) string {
	_, file, line, ok := runtime.Caller(skip)
	if !ok {
		return "unknown"
	}
	return strings.Join([]string{
		CallerPathClean(file),
		strconv.Itoa(line),
	}, ":")
}

// CallerPathClean 对 caller 的文件路径进行精简
// 原始的是完整的路径，比较长，该方法可以将路径变短
var CallerPathClean = callerPathClean

var callerSearchPrefixPaths = []string{
	"icode.baidu.com/baidu/",
	"task_workspace/baidu/",
}

func callerPathClean(file string) string {
	var pos int
	for _, p := range callerSearchPrefixPaths {
		index := strings.Index(file, p)
		if index >= 0 {
			pos = index + len(p)
			break
		}
	}
	if pos <= 0 {
		return file
	}
	return file[pos:]
}
