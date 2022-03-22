// Copyright(C) 2021 Baidu Inc. All Rights Reserved.
// Author: Wei Du (duwei04@baidu.com)
// Date: 2021/3/29

package pool

import (
	"context"
	"encoding/json"
	"errors"
	"time"
)

// ErrBadValue not active element
var ErrBadValue = errors.New("bad pool value")

// ErrOutOfMaxLife out of max life
var ErrOutOfMaxLife = errors.New("pool value out of max life")

// ErrOutOfMaxIdle out of max idle
var ErrOutOfMaxIdle = errors.New("pool value out of max idle")

// ErrOutOfMaxIdleTime out of max idle time
var ErrOutOfMaxIdleTime = errors.New("pool value out of max idle time")

// nowFunc returns the current time; it's overridden in tests.
var nowFunc = time.Now

// ErrClosed 对象池已关闭
var ErrClosed = errors.New("pool already closed")

// Pool 通用的 Pool 接口定义
type Pool interface {
	Get(ctx context.Context) (interface{}, error)
	Put(interface{}) error
	Stats() Stats
	Close() error
	Option() Option
}

// Option pool option
type Option struct {
	// MaxOpen max opening Element
	// <= 0 means unlimited
	MaxOpen int

	// MaxIdle
	// <=0 means disabled
	MaxIdle int

	// MaxLifeTime
	// maximum amount of time a Element may be reused
	MaxLifeTime time.Duration

	// MaxIdleTime
	// maximum amount of time a Element may be idle before being closed
	MaxIdleTime time.Duration
}

func (opt *Option) shortestIdleTime() time.Duration {
	if opt.MaxIdleTime <= 0 {
		return opt.MaxLifeTime
	}
	if opt.MaxLifeTime <= 0 {
		return opt.MaxIdleTime
	}

	min := opt.MaxIdleTime
	if min > opt.MaxLifeTime {
		min = opt.MaxLifeTime
	}
	return min
}

// Clone copy it
func (opt *Option) Clone() *Option {
	return &Option{
		MaxOpen:     opt.MaxOpen,
		MaxIdle:     opt.MaxIdle,
		MaxLifeTime: opt.MaxLifeTime,
		MaxIdleTime: opt.MaxIdleTime,
	}
}

// String 序列化，调试输出用
func (opt *Option) String() string {
	bf, _ := json.Marshal(opt)
	return string(bf)
}

// Stats Pool's Stats
type Stats struct {
	Open bool // pool opening status

	// simplePool Status
	NumOpen int // The number of established Elements both in use and idle.
	InUse   int // The number of Elements currently in use.
	Idle    int // The number of idle Elements.

	// Counters
	WaitCount         int64         // The total number of Elements waited for.
	WaitDuration      time.Duration // The total time blocked waiting for a new Element.
	MaxIdleClosed     int64         // The total number of Elements closed.
	MaxIdleTimeClosed int64         // The total number of Elements closed.
	MaxLifeTimeClosed int64         // The total number of Elements closed.
}

// String 序列化，调试用
func (s Stats) String() string {
	bf, _ := json.Marshal(s)
	return string(bf)
}

// GroupStats Group Pool stats
type GroupStats struct {
	// Groups 各个组的状态，使用[]类型的兼容性更好
	// 而 map[interface{}]类型是不能正常 json encode 的
	Groups []*GroupStatDetail
	All    Stats
}

// GroupStatDetail GroupStats 类型中使用，一个 group 的状态
type GroupStatDetail struct {
	Group interface{}
	Stats Stats
}

// NewElementNeed 创建新 Element 时所需要的
type NewElementNeed interface {
	Put(interface{}) error
	Option() Option
}
