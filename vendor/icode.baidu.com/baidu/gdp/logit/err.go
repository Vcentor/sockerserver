// Copyright(C) 2021 Baidu Inc. All Rights Reserved.
// Author: Wei Du (duwei04@baidu.com)
// Date: 2021/12/2

package logit

import (
	"context"
	"fmt"
	"strconv"
)

// SetErr 给 Ctx 里添加上 err 日志字段，该字段一般用于标记当前请求是否成功处理。
// ctx 应提前初始化，否则将 panic
// 	具体含义如下：
// 	code=0: 处理成功了(还包括上游传入异常参数的情况)
// 	code=1: 处理失败了(服务已接收到正确的参数，但是还是处理失败了)
// 	code=2: 降级处理了
// 	各种 server 的日志里一般会添加该字段，应尽量保证在不同的 server 含义相同
// 	该字段和日志中的 errno 字段表达的含义不一样，errno 表示的是更具体的错误码
// 	比如对于 HTTP Server，errno=400 表示的是请求参数异常，errno=200 是成功,
// 	这个时候，errno 等于 400 和 200 的时候，都可以设置 err=0
func SetErr(ctx context.Context, code int) {
	var f Field
	// 对常用的几个值进行优化
	switch code {
	case 0:
		f = errField0
	case 1:
		f = errField1
	case 2:
		f = errField2
	default:
		var has bool
		if f, has = errFieldOther[code]; !has {
			f = Int("err", code)
		}
	}
	ReplaceFields(ctx, f)
}

var (
	errField0 = Int("err", 0)
	errField1 = Int("err", 1)
	errField2 = Int("err", 2)

	errFieldOther = map[int]Field{}
)

func init() {
	errField0.SetLevel(AllLevels)
	errField1.SetLevel(AllLevels)
	errField2.SetLevel(AllLevels)

	for i := 3; i < 10; i++ {
		f := Int("err", i)
		f.SetLevel(AllLevels)
		errFieldOther[i] = f
	}
}

// GetErr 读取 ctx 中的 err 字段的值
// 	若不存在,将返回 -1
// 	若解析失败，将返回 1
// 	该字段一般是通过 SetErr 方法写入
func GetErr(ctx context.Context) int {
	f := FindField(ctx, "err")
	if f == nil {
		return -1
	}

	var str string

	switch f.Type() {
	case IntType:
		return f.Value().(int)
	case StringType:
		str = f.Value().(string)
		switch str {
		case "-1":
			return -1
		case "0":
			return 0
		case "1":
			return 1
		case "2":
			return 2
		}
	default:
		str = fmt.Sprint(f.Value())
	}
	code, err := strconv.Atoi(str)
	if err == nil {
		return code
	}
	// 解析失败的情况
	return 1
}
