// Copyright(C) 2020 Baidu Inc. All Rights Reserved.
// Author: Wei Du (duwei04@baidu.com)
// Date: 2020/8/10

package logit

import (
	"strconv"
	"sync/atomic"
	"time"
)

// NewLogID 获取一个新的logid，默认是 uint32 的字符串
var NewLogID = func() string {
	return strconv.FormatUint(uint64(NewLogIDUint32()), 10)
}

var idx uint32

func rid() uint32 {
	id := atomic.AddUint32(&idx, 1)
	if id < 65535 {
		return id
	}
	atomic.StoreUint32(&idx, 0)
	return rid()
}

// NewLogIDUint32 获取一个新的logid
func NewLogIDUint32() uint32 {
	usec := now().UnixNano()
	logid := usec&0x7FFFFFFF | 0x80000000
	// 通过rid()让同一时间生成的logid不重复
	return uint32(logid) + rid()
}

var now = func() time.Time {
	return time.Now()
}
