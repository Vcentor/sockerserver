// Copyright(C) 2021 Baidu Inc. All Rights Reserved.
// Author: Wei Du (duwei04@baidu.com)
// Date: 2021/7/20

package timer

import (
	"sync"
	"time"
)

// NewTimeMachine 创建一个用于测试的时光机
func NewTimeMachine(cst string) *TimeMachine {
	tm := &TimeMachine{}
	tm.MustTravelByCST(cst)
	return tm
}

// NewTimeMachineDefault 创建一个默认的时光机
// 默认会穿梭到 2020-09-24 10:30:00.123456789
func NewTimeMachineDefault() *TimeMachine {
	return NewTimeMachine("2020-09-24 10:30:00.123456789")
}

// TimeMachine 时光机，一般用于测试，
// 调用 Now 方法的时候， 总是可以返回一个预期中的时间
//
// 使用方式：
//
// 	1. 在正式代码中：
// 	var nowFunc = time.Now
//
// 	2. 在测试代码中：
// 	var tm=NewTimeMachine("2020-12-08 20:38:45")
// 	func init(){
// 		nowFunc=tm.Now
// 	}
type TimeMachine struct {
	begin time.Time
	now   time.Time
	mux   sync.Mutex

	running    bool
	continueAt time.Time
	disable    bool
}

// Now 返回当前时间
func (tm *TimeMachine) Now() time.Time {
	tm.mux.Lock()
	defer tm.mux.Unlock()

	if tm.disable {
		return time.Now()
	}
	now := tm.now
	if tm.running {
		now = now.Add(time.Since(tm.continueAt))
	}
	return now
}

// Travel 穿越到指定时间
func (tm *TimeMachine) Travel(now time.Time) {
	tm.mux.Lock()

	tm.now = now

	if tm.begin.IsZero() {
		tm.begin = now
	}

	tm.mux.Unlock()
}

// Jump 在当前时间基础上跳跃指定时长
//
// 建议不要在 Continue=true 的时候使用，否则获取的时间是不固定的
func (tm *TimeMachine) Jump(dur time.Duration) {
	now := tm.Now()
	tm.Travel(now.Add(dur))
}

// MustTravelByCST 穿越到指定时间
// 使用字符串CST格式的时间，如 2020-12-08 20:38:45
func (tm *TimeMachine) MustTravelByCST(now string) {
	t, err := time.ParseInLocation(LayoutCST, now, time.Local)
	if err != nil {
		panic(err)
	}
	tm.Travel(t)
}

// Continue 是否让时间继续流动
// 设置 running=false 后，Now 方法返回一个固定的时间(时间暂停)(这个是默认 TimeMachine 的行为)
// 设置 running=true 后，Now 方法返回一个继续变动的时间(时间继续流动向前)
func (tm *TimeMachine) Continue(running bool) {
	tm.mux.Lock()

	tm.running = running
	if running {
		tm.continueAt = time.Now()
	}

	tm.mux.Unlock()
}

// Enable 是否停止时光机
func (tm *TimeMachine) Enable(enable bool) {
	tm.mux.Lock()
	tm.disable = !enable
	tm.mux.Unlock()
}

// Reset 重置为初始化状态
func (tm *TimeMachine) Reset() {
	tm.mux.Lock()

	tm.disable = false
	tm.now = tm.begin
	tm.running = false

	tm.mux.Unlock()
}
