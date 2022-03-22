// Copyright(C) 2020 Baidu Inc. All Rights Reserved.
// Author: Wei Du (duwei04@baidu.com)
// Date: 2020/12/31

package timer

import (
	"sync"
	"time"
)

// Now 获取当前时间
//
// 	在测试过程中,可以使用 UseTimeMachine 来改变当前函数的行为
// 	使用该函数后，要避免使用标准库的 time.Since，而应该使用 timer.Now().Sub(old)
// 	其他的和 time.Since 类似的和当前时间相关的函数也需要避免使用标准库的
func Now() time.Time {
	return nowFunc1()
}

var nowFunc = time.Now

// 为了先不改的当前 pkg 内的其他单测以及代码，
// 这里暂时先使用另外一个变量，后续再统一调整
var nowFunc1 = time.Now

var defaultTM = NewTimeMachineDefault()
var dumOnce sync.Once

// UseTimeMachine  启动或者停止时光机，只供在测试场景使用
//
// 	当 enable=true ，方法 Now() 返回固定的时间
// 	当 enable=false, 方法 Now() 返回正常的当前时间,并且会将
// 	基于对 Now 函数的性能考虑，当前函数和 Now() 并不是并发安全的，
// 	首次应该在测试的 init 过程中调用，首次调用之后是并发安全的
// 	每次调用都会将 TimeMachine 给 Reset
func UseTimeMachine(enable bool) *TimeMachine {
	dumOnce.Do(func() {
		nowFunc1 = defaultTM.Now
	})
	defaultTM.Reset()
	defaultTM.Enable(enable)
	return defaultTM
}

const (
	// LayoutCST 常用日期格式
	LayoutCST = "2006-01-02 15:04:05"

	// LayoutCSTMilli 常用日期格式，精确到毫秒(3位)
	LayoutCSTMilli = "2006-01-02 15:04:05.000"

	// LayoutCSTMicro 常用日期格式，精确到微妙(6位)
	LayoutCSTMicro = "2006-01-02 15:04:05.000000"

	// LayoutCSTNano 常用日期格式，精确到纳秒(9位)
	LayoutCSTNano = "2006-01-02 15:04:05.000000000"
)
