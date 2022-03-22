// Copyright(C) 2020 Baidu Inc. All Rights Reserved.
// Author: Chen Xin (chenxin@baidu.com)
// Date: 2020/04/19

package logit

import (
	"time"
)

// TimeCost 计时器
type TimeCost struct {
	key string
	t   time.Time
}

// NewTimeCost 计时 field
func NewTimeCost(key string) func() Field {
	tc := TimeCost{
		key: key,
		t:   time.Now(),
	}
	return tc.Stop
}

// Stop 停止计时，并返回一个Field
func (tc *TimeCost) Stop() Field {
	d := time.Since(tc.t)
	return Duration(tc.key, d)
}
